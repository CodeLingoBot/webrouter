package router

import (
	_ "github.com/morya/im/pkg/protocol"
)

const (
	TYPE_RANGE = iota
	TYPE_EQ
	TYPE_NOT
)

// rule 格式：json

type Rule struct {
	Type  int
	value interface{}
}

func NewRule(rule string) *Rule {
	return &Rule{}
}

// func (self *Rule) exec(head *protocol.MessageHead) bool {
// 	return false
// }

// func (self *Rule) execByIp(head *protocol.MessageHead) bool {
// 	v, _ := value.(string)
// 	return head.GetIp() == v
// }

// func (self *Rule) execByCmd(head *protocol.MessageHead) bool {
// 	v, _ := value.(int32)
// 	return head.GetCmd() == v
// }
