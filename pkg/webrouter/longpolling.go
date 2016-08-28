package webrouter

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/CodisLabs/codis/pkg/utils/atomic2"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/morya/net"
	"github.com/morya/webrouter/pkg/util"
	ws "golang.org/x/net/websocket"
)

const (
	ERR_NONE      = 0
	ERR_KNOWN_SID = 100
	ERR_KNOWN_VID = 101
)

var counter atomic2.Int64

type Longpolling struct {
	sync.Mutex

	Err int   `json:"err,omitempty"`
	Sid int64 `json:"sid,omitempty"` // sessionId
	Vid int   `json:"vid,omitempty"` // vid

	messages *list.List

	conn *ws.Conn

	createAt      time.Time
	lastConnectAt time.Time

	manager         *LongpollingManager
	wheelKey        string
	timeoutDuration time.Duration
}

func NewErrEmptyLongpolloing(err int) *Longpolling {
	return &Longpolling{
		Err: err,
	}
}

func newLongpolling(manager *LongpollingManager) *Longpolling {
	now := time.Now()

	l := &Longpolling{
		Err:             0,
		createAt:        now,
		lastConnectAt:   now,
		Sid:             getSid(),
		Vid:             0,
		manager:         manager,
		timeoutDuration: manager.cnf.LongpollingInterval + manager.cnf.NetDelay*2,
		messages:        list.New(),
		wheelKey:        net.NewKey(),
	}

	return l
}

func (self *Longpolling) ToString() string {
	return fmt.Sprintf("Sid<%d>, Vid<%d>, CreateAt<%s>, LastConnectAt<%s>", self.Sid, self.Vid, self.createAt, self.lastConnectAt)
}

func (self *Longpolling) ResetTimeout() {
	// longpolling connect operation, cancel timeout, reset timeout, refresh vid
	self.manager.wheel.Cancel(self.wheelKey)
	self.manager.wheel.AddWithId(self.timeoutDuration, self.wheelKey, self.timeout)

	self.RefreshVid()
}

func (self *Longpolling) RefreshVid() {
	self.Vid += 1
	self.lastConnectAt = time.Now()
}

func (self *Longpolling) WaitResponse() {
	select {
	case <-time.After(self.manager.cnf.LongpollingInterval):
		return
	}
}

func (self *Longpolling) GetData() *Data {
	self.Lock()
	defer self.Unlock()

	value := util.ToArray(self.messages)
	self.messages.Init()
	return newDataValues(self.Sid, self.Vid, value)
}

func (self *Longpolling) AppendData(value string) {
	self.Lock()
	defer self.Unlock()

	self.messages.PushBack(value)
}

func (self *Longpolling) timeout(wheelKey string) {
	log.Infof("%s Timeout Sid<%s> last connect<%s>", MODULE_SERVER_FRONTEND, self.Sid, self.lastConnectAt)
	self.manager.DeleteLongpolling(self.Sid)
}

func getSid() int64 {
	return counter.Incr()
}
