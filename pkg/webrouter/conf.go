package webrouter

import (
	"time"
)

type Conf struct {
	Addr                string
	LongpollingInterval time.Duration
	NetDelay            time.Duration
	BackendSyncTimeout  time.Duration
	Backends            []string
	Name                string
	ConnectType         byte
	MsgType             byte
}
