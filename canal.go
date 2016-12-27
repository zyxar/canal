package canal

import (
	"container/list"
	"sync"
)

type Canal interface {
	Close()
	CloseAndWait()
	Recv() (val interface{}, ok bool, opened bool)
	Chan() chan<- interface{}
	Len() int
}

type canalImpl struct {
	closed chan bool
	*sync.WaitGroup
	*list.List
	send chan<- interface{}
	recv <-chan *list.Element
}

func (c *canalImpl) Close() {
	close(c.closed)
	close(c.send)
}

func (c *canalImpl) CloseAndWait() {
	close(c.closed)
	close(c.send)
	c.WaitGroup.Wait()
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
		make(chan bool),
		&sync.WaitGroup{},
		list.New(),
		send,
		recv,
	}
	go func() {
		for {
			// select {
			// case <-c.closed:
			// 	goto closed
			// default:
			// }
			select {
			case <-c.closed:
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
	c.WaitGroup.Add(1)
	return c
}
