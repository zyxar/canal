package canal

import (
	"container/list"
	"sync"
)

type Canal interface {
	Pipe() (SendChan, RecvChan)
	IsClosed() bool
	Close()
	Wait()
	Len() int
}

type SendChan chan<- interface{}
type RecvChan <-chan interface{}

type canalImpl struct {
	closeChan chan struct{}
	closeOnce sync.Once
	sync.WaitGroup
	*list.List
	send chan<- interface{}
	recv <-chan interface{}
}

func (c *canalImpl) Close() {
	c.closeOnce.Do(func() {
		close(c.closeChan)
		// close(c.send)
	})
}

func (c *canalImpl) IsClosed() bool {
	select {
	case <-c.closeChan:
		return true
	default:
	}
	return false
}

func (c *canalImpl) Pipe() (SendChan, RecvChan) {
	return c.send, c.recv
}

func New() Canal {
	send := make(chan interface{})
	recv := make(chan interface{})
	c := &canalImpl{
		closeChan: make(chan struct{}),
		List:      list.New(),
		send:      send,
		recv:      recv,
	}
	c.WaitGroup.Add(1)
	go func() {
		for {
			if c.IsClosed() {
				goto closed
			}

			front := c.Front()
			var val interface{}
			if front != nil {
				val = front.Value
			}

			select {
			case <-c.closeChan:
				goto closed
			case v, ok := <-send:
				if ok {
					c.PushBack(v)
				} else {
					goto closed
				}
			case recv <- val:
				if front != nil {
					c.Remove(front)
				}
			}
		}
	closed:
		for c.List.Len() > 0 {
			select {
			case recv <- c.Front().Value:
				c.Remove(c.Front())
			}
		}
		close(recv)
		c.WaitGroup.Done()
	}()
	return c
}
