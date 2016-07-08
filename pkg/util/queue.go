package util

import (
	"errors"
	"sync"
)

var (
	ERROR_NO_ENOUGH = errors.New("has not enough element")
)

type Queue struct {
	sync.Mutex
	capacity int
	count    int
	data     []interface{}
}

func NewQueue(capacity int) *Queue {
	return &Queue{
		capacity: capacity,
		count:    0,
		data:     make([]interface{}, capacity),
	}
}

func (self *Queue) Append(e interface{}) {
	self.Lock()
	defer self.Unlock()

	self.ensure()

	self.data[self.count] = e
	self.count++
}

func (self *Queue) Get(num int) ([]interface{}, error) {
	if self.count < num {
		return nil, ERROR_NO_ENOUGH
	}

	return self.data[:num], nil
}

func (self *Queue) Remove(num int) int {
	self.Lock()
	defer self.Unlock()

	if num <= 0 {
		return 0
	}

	if num >= self.count {
		self.data = make([]interface{}, self.capacity)
		self.count = 0
		return self.count
	} else {
		newData := make([]interface{}, self.capacity)
		copy(newData, self.data[num+1:])

		self.count -= num

		return num
	}
}

func (self *Queue) Size() int {
	return self.count
}

func (self *Queue) ensure() {
	if !self.isFull() {
		return
	}

	newCapacity := self.capacity/4 + self.capacity
	newData := make([]interface{}, newCapacity)
	copy(newData, self.data)

	self.data = newData
	self.capacity = newCapacity
}

func (self *Queue) isFull() bool {
	return self.count == self.capacity
}
