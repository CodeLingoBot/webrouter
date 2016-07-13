package router

import (
	"errors"
	"github.com/morya/im/pkg/protocol"
	"sync"
)

var (
	ERR_NO_ROUTE = errors.New("no route")
)

// 路由表
type RouteTable struct {
	lock sync.RWMutex
}

// 路由
type Route struct {
	Backend string
	rule    []*Rule
}

func NewRouteTable() *RouteTable {
	return &RouteTable{}
}

func (self *RouteTable) GetRoute(head *protocol.MessageHead) (string, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	return "nil", nil
}

func (self *RouteTable) Reload() {
	self.lock.Lock()
	defer self.lock.Unlock()
}
