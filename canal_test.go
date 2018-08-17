package canal_test

import (
	"sync"
	"testing"
	"time"

	"github.com/zyxar/canal"
)

const (
	num = 100
)

func TestSendNonBlocking(t *testing.T) {
	t.Parallel()
	can := canal.New()
	done := make(chan bool)
	go func() {
		for i := 0; i < num; i++ {
			can.Send(i)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("blocking!")
	}
	can.Close()
}

func TestRecvNonBlocking(t *testing.T) {
	t.Parallel()
	can := canal.New()
	done := make(chan bool)
	go func() {
		for i := 0; i < num; i++ {
			v, err := can.Recv()
			if err != canal.ErrCanalRecvNil || v != nil {
				t.Error("invalid recv", v, err)
			}
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("blocking!")
	}
	can.Close()
}

func TestSendRecvNonBlocking(t *testing.T) {
	t.Parallel()
	can := canal.New()
	for i := 0; i < num; i++ {
		if err := can.Send(i); err != nil {
			t.Error(i, err.Error())
		}
	}
	can.Close()
	for i := 0; i < num; i++ {
		v, err := can.Recv()
		if err != nil || v == nil || v.(int) != i {
			t.Error("invalid recv", v, err)
		}
	}
	// can.Wait()
}

func TestSendRecvWaitNonBlocking(t *testing.T) {
	t.Parallel()
	can := canal.New()
	for i := 0; i < num; i++ {
		if err := can.Send(i); err != nil {
			t.Error(i, err.Error())
		}
	}
	for i := 0; i < num; i++ {
		v, err := can.Recv()
		if err != nil || v == nil || v.(int) != i {
			t.Error("invalid recv", v, err)
		}
	}
	can.Close()
	can.Wait()
}

func TestComprehensiveNonBlocking(t *testing.T) {
	t.Parallel()
	can := canal.New()
	wg := &sync.WaitGroup{}
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func(i int) {
			if err := can.Send(i); err != nil {
				t.Error(i, err.Error())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	<-time.After(time.Microsecond)
	if can.Len() != num {
		t.Errorf("canal length not correct: %d != %d", can.Len(), num)
	}
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func(i int) {
			v, err := can.Recv()
			if err != nil || v == nil {
				t.Error("invalid recv", v, err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	<-time.After(time.Microsecond)
	if can.Len() != 0 {
		t.Errorf("canal length not correct: %d != %d", can.Len(), 0)
	}
	can.Close()
	can.Wait()
}
