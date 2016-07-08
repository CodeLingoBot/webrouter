package webrouter

import (
	"encoding/json"
	"fmt"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/funny/link"
	"sync"
	"time"
)

type Backend struct {
	sync.Mutex
	session     *link.Session
	pos         int
	servers     []string
	handshake   string
	serverName  string
	connectType byte
	msgType     byte
	mgr         *LongpollingManager
}

func NewBackend(addrs []string, serverName string, connectType, msgType byte, mgr *LongpollingManager) *Backend {
	bkn := &Backend{
		pos:         0,
		servers:     addrs,
		handshake:   fmt.Sprintf(`{"servername":"%s"}`, serverName),
		serverName:  serverName,
		connectType: connectType,
		msgType:     msgType,
		mgr:         mgr,
	}

	log.Infof("%s handshake is <%s>", MODULE_SERVER_BACKEND, bkn.handshake)

	return bkn
}

func (self *Backend) Loop() {
	for {
		err := self.loop0()

		if nil != err {
			log.InfoErrorf(err, "%s Connect to <%s> failure", MODULE_SERVER_BACKEND, self.currentServer())
			time.Sleep(time.Second * 10)
		}
	}
}

func (self *Backend) loop0() error {
	var err error
	self.session, err = link.Connect("tcp", self.selectServer(), newCodecFactory(self))
	if err != nil {
		return err
	}
	defer self.session.Close()

	log.InfoErrorf(err, "%s Connect to <%s> successed", MODULE_SERVER_BACKEND, self.currentServer())

	err = self.sendStr(self.handshake)
	if err != nil {
		return err
	}

	log.InfoErrorf(err, "%s Handshake to <%s> successed", MODULE_SERVER_BACKEND, self.currentServer())

	info := fmt.Sprintf(`{"src":100,"cmd":"LOGIN","routername":"%s"}`, self.serverName)
	err = self.sendStr(info)
	if err != nil {
		return err
	}

	log.InfoErrorf(err, "%s Login <%s> to <%s> successed", MODULE_SERVER_BACKEND, info, self.currentServer())

	for {
		log.Debugf("%s begin received", MODULE_SERVER_BACKEND)

		data := make(map[string]interface{})

		if err := self.session.Receive(&data); err != nil {
			return err
		}

		log.Debugf("%s end received", MODULE_SERVER_BACKEND)
	}
}

func (self *Backend) sendStr(str string) error {
	err := self.session.Send(int(len(str)))
	if err != nil {
		return err
	}

	err = self.session.Send(self.connectType)
	if err != nil {
		return err
	}
	err = self.session.Send(byte(0))
	if err != nil {
		return err
	}
	err = self.session.Send(byte(0x0d))
	if err != nil {
		return err
	}
	err = self.session.Send(byte(0x0a))
	if err != nil {
		return err
	}
	err = self.session.Send([]byte(str))
	if err != nil {
		return err
	}

	return nil
}

func (self *Backend) selectServer() string {
	self.Lock()

	defer func() {
		self.pos += 1
		if self.pos == len(self.servers) {
			self.pos = 0
		}

		self.Unlock()
	}()
	return self.servers[self.pos]
}

func (self *Backend) currentServer() string {
	self.Lock()
	defer self.Unlock()

	return self.servers[self.pos]
}

func (self *Backend) received(msg map[string]interface{}) {
	cmd := msg["cmd"]

	if cmd == CMD_QUERY_STATUS {
		flag, err := getInt64Value(msg, "I")
		if err != nil {
			log.WarnErrorf(err, "%s Ack Query failure", MODULE_SERVER_BACKEND)
			return
		}

		self.mgr.AckSync(flag, msg)
		return
	}

	sid, err := getInt64Value(msg, "sessionid")

	if nil != err {
		log.Warnf("%s Unmarshal failure <%+v>", MODULE_SERVER_BACKEND, msg)
		return
	}

	longpolling := self.mgr.GetLongpolling(int64(sid))
	if nil == longpolling {
		log.Warnf("%s Session <%d> not found", MODULE_SERVER_BACKEND, sid)
		return
	}

	m, _ := json.Marshal(msg)
	err = self.mgr.SendToFrondend(sid, string(m))

	if err != nil {
		log.InfoErrorf(err, "%s Send msg <%s> to Session <%d> failure", MODULE_SERVER_BACKEND, string(m), sid)
	}
}

func (self *Backend) Send(message interface{}, sid int64) error {
	v, err := json.Marshal(message)

	if nil != err {
		return err
	}

	log.Info(formatSession(sid, MODULE_SERVER_BACKEND, "Send msg <%s> to <%s>", string(v), self.currentServer()))
	return self.session.Send(string(v))
}

func (self *Backend) NotifyClosed(sid int64) error {
	m := map[string]interface{}{
		"src":       100,
		"cmd":       "SESSION_LOGOUT",
		"sessionid": sid,
	}

	return self.Send(m, sid)
}

func getInt64Value(msg map[string]interface{}, key string) (int64, error) {
	v := msg[key]

	t, ok := v.(json.Number)

	if !ok {
		log.Warnf("%s Unmarshal failure <%+v>", MODULE_SERVER_BACKEND, msg)
		return 0, ERR_SESSION_NOT_FOUND
	}

	value, err := t.Int64()

	if nil != err {
		log.Warnf("%s Unmarshal failure <%+v>", MODULE_SERVER_BACKEND, msg)
		return 0, err
	}

	return value, nil
}
