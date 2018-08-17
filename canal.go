package canal

import (
	"container/list"
	"errors"
	"sync"
)

var (
	ErrCanalClosed  = errors.New("canal closed")         // returned when canal is closed
	ErrCanalRecvNil = errors.New("canal recv value nil") // returned when there is nothing to be retrieved from canal
)

type Canal interface {
	Send(v interface{}) error   // chan<-
	Recv() (interface{}, error) // <-chan
	Len() int                   // len(chan)
	Close()                     // close(chan)
	Wait()
}

type canalImpl struct {
	*list.List
	send chan<- interface{}
	recv <-chan *list.Element

	closeChan chan struct{}
	closeOnce sync.Once
	sync.WaitGroup
}

func (c *canalImpl) Close() {
	c.closeOnce.Do(func() {
		close(c.closeChan)
	})
}

func (c *canalImpl) Send(v interface{}) error {
	select {
	case <-c.closeChan:
		return ErrCanalClosed
	default:
	}

	select {
	case <-c.closeChan:
		return ErrCanalClosed
	case c.send <- v:
		return nil
	}
}

func (c *canalImpl) Recv() (interface{}, error) {
	v, ok := <-c.recv
	if !ok {
		return nil, ErrCanalClosed
	}
	if v == nil {
		return nil, ErrCanalRecvNil
	}
	return v.Value, nil
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
			select {
			case <-c.closeChan:
				goto cleanup
			default:
			}

			select {
			case <-c.closeChan:
				goto cleanup
			case v, ok := <-send:
				if ok {
					c.PushBack(v)
				}
			case recv <- c.Front():
				if c.Front() != nil {
					c.Remove(c.Front())
				}
			}
		}
	cleanup:
		for c.List.Len() > 0 {
			val := c.Front()
			recv <- val
			c.Remove(val)
		}
		close(recv)
		c.WaitGroup.Done()
	}()
	return c
}
