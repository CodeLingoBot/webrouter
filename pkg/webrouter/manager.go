package webrouter

import (
	"errors"
	"github.com/CodisLabs/codis/pkg/utils/atomic2"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/net"
	ws "golang.org/x/net/websocket"
	"strconv"
	"strings"
	"sync"
	"time"
)

var counterI atomic2.Int64

var (
	CMD_QUERY_STATUS = "GET_LOGINSTATUS"
)

var (
	ERR_BACKEND_RESP_TIMEOUT = errors.New("backend response timeout")
	ERR_SESSION_NOT_FOUND    = errors.New("session not found")
	ERR_NOT_JSON             = errors.New("not a json")
)

type LongpollingManager struct {
	pollingLock *sync.Mutex
	pollings    map[int64]*Longpolling
	wsMapping   map[*ws.Conn]int64
	cnf         *Conf
	wheel       *net.HashedTimeWheel

	backend *Backend

	chs map[int64]chan map[string]interface{}
}

func NewLongpollingManager(cnf *Conf) *LongpollingManager {
	m := &LongpollingManager{
		cnf:         cnf,
		pollingLock: &sync.Mutex{},
		pollings:    make(map[int64]*Longpolling),
		wsMapping:   make(map[*ws.Conn]int64),
		wheel:       net.NewHashedTimeWheel(time.Second, 60, 3),
		chs:         make(map[int64]chan map[string]interface{}),
	}

	backend := NewBackend(cnf.Backends, cnf.Name, cnf.ConnectType, cnf.MsgType, m)
	m.backend = backend

	go backend.Loop()

	m.wheel.Start()

	return m
}

func (self *LongpollingManager) NewWebSocketLongpolling(conn *ws.Conn) *Longpolling {
	l := self.NewLongpolling()

	self.wsMapping[conn] = l.Sid
	l.conn = conn

	return l
}

func (self *LongpollingManager) NewLongpolling() *Longpolling {
	longpolling := newLongpolling(self)

	self.pollingLock.Lock()
	defer self.pollingLock.Unlock()

	self.pollings[longpolling.Sid] = longpolling

	return longpolling
}

func (self *LongpollingManager) GetLongpolling(sid int64) *Longpolling {
	self.pollingLock.Lock()
	defer self.pollingLock.Unlock()

	longpolling, _ := self.pollings[sid]

	return longpolling
}

func (self *LongpollingManager) DeleteLongpolling(sid int64) {
	self.pollingLock.Lock()
	defer self.pollingLock.Unlock()

	delete(self.pollings, sid)

	log.Info(formatSession(sid, MODULE_SERVER_MGR, "Removed"))

	self.backend.NotifyClosed(sid)
}

func (self *LongpollingManager) DeleteWebSocketLongpolling(conn *ws.Conn) {
	sid, ok := self.wsMapping[conn]
	if !ok {
		return
	}

	delete(self.wsMapping, conn)
	self.DeleteLongpolling(sid)
}

func (self *LongpollingManager) Exist(sid int64) bool {
	self.pollingLock.Lock()
	defer self.pollingLock.Unlock()

	_, ok := self.pollings[sid]

	return ok
}

func (self *LongpollingManager) SendToBackend(sid int64, data *Data) error {
	for _, r := range data.Value {
		err := appendSessionId(r, sid)

		if nil != err {
			return err
		}

		err = self.backend.Send(r, sid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (self *LongpollingManager) SendToFrondend(sid int64, msg string) error {
	l := self.GetLongpolling(sid)

	if nil == l {
		return ERR_SESSION_NOT_FOUND
	}

	if l.conn != nil {
		defer l.RefreshVid()
		return ws.Message.Send(l.conn, newData(sid, l.Vid, msg).Marshal())
	}

	return nil
}

func (self *LongpollingManager) QueryUserStatus(userIds string) (map[string]interface{}, error) {
	ids := strings.Split(userIds, ",")
	syncI := counterI.Incr()
	ch := make(chan map[string]interface{})

	self.PutSync(syncI, ch)
	defer func() {
		close(ch)
		delete(self.chs, syncI)
	}()

	userStatus := make([]map[string]int, len(ids))

	for index, idStr := range ids {
		id, _ := strconv.Atoi(idStr)

		userStatus[index] = map[string]int{
			"id": id,
		}
	}

	msg := map[string]interface{}{
		"src":        1,
		"cmd":        CMD_QUERY_STATUS,
		"I":          syncI,
		"userstatus": userStatus,
	}

	err := self.backend.Send(msg, 0)

	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(self.cnf.BackendSyncTimeout):
		return nil, ERR_BACKEND_RESP_TIMEOUT
	case rsp := <-ch:
		return rsp, nil
	}
}

func (self *LongpollingManager) PutSync(flag int64, ch chan map[string]interface{}) {
	self.chs[flag] = ch
}

func (self *LongpollingManager) AckSync(flag int64, msg map[string]interface{}) {
	ch, ok := self.chs[flag]

	if ok {
		ch <- msg
	}
}

func appendSessionId(info interface{}, sid int64) error {
	m, ok := info.(map[string]interface{})

	if !ok {
		return ERR_NOT_JSON
	}

	m["sessionid"] = sid

	return nil
}
