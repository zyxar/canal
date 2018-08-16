package canal

import (
	"container/list"
	"sync"
)

type Canal interface {
	IsClosed() bool
	Close()
	Wait()
	Recv() (val interface{}, ok bool, opened bool)
	Chan() chan<- interface{}
	Len() int
}

type canalImpl struct {
	closeChan chan struct{}
	closeOnce sync.Once
	sync.WaitGroup
	*list.List
	send chan<- interface{}
	recv <-chan *list.Element
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

func (c *canalImpl) Chan() chan<- interface{} {
	return c.send
}

func (c *canalImpl) Recv() (val interface{}, ok bool, opened bool) {
	v, opened := <-c.recv
	if !opened {
		return
	}
	if v == nil {
		return
	}
	ok = true
	val = v.Value
	return
}

func New() Canal {
	send := make(chan interface{})
	recv := make(chan *list.Element)
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

			select {
			case <-c.closeChan:
				goto closed
			case v, ok := <-send:
				if ok {
					c.PushBack(v)
				} else {
					goto closed
				}
			case recv <- c.Front():
				if c.Front() != nil {
					c.Remove(c.Front())
				}
			}
		}
	closed:
		for c.List.Len() > 0 {
			select {
			case recv <- c.Front():
				c.Remove(c.Front())
			}
		}
		close(recv)
		c.WaitGroup.Done()
	}()
	return c
}
