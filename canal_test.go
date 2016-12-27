package canal_test

import (
	"."
	"sync"
	"testing"
	"time"
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
			can.Chan() <- i
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
			v, ok, opened := can.Recv()
			if !opened || ok || v != nil {
				t.Error("invalid recv")
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
		can.Chan() <- i
	}
	can.Close()
	for i := 0; i < num; i++ {
		v, ok, opened := can.Recv()
		if !opened || !ok || v == nil || v.(int) != i {
			t.Error("invalid recv")
		}
	}
	<-time.After(time.Microsecond)
	if _, _, opened := can.Recv(); opened {
		t.Error("invalid recv")
	}
}

func TestSendRecvWaitNonBlocking(t *testing.T) {
	t.Parallel()
	can := canal.New()
	for i := 0; i < num; i++ {
		can.Chan() <- i
	}
	for i := 0; i < num; i++ {
		v, ok, opened := can.Recv()
		if !opened || !ok || v == nil || v.(int) != i {
			t.Error("invalid recv")
		}
	}
	can.CloseAndWait()
	if _, _, opened := can.Recv(); opened {
		t.Error("invalid recv")
	}
}

func TestComprehensiveNonBlocking(t *testing.T) {
	t.Parallel()
	can := canal.New()
	wg := &sync.WaitGroup{}
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func(i int) {
			can.Chan() <- i
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
			v, ok, opened := can.Recv()
			if !opened || !ok || v == nil {
				t.Error("invalid recv")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	<-time.After(time.Microsecond)
	if can.Len() != 0 {
		t.Errorf("canal length not correct: %d != %d", can.Len(), 0)
	}
	can.CloseAndWait()
	if _, _, opened := can.Recv(); opened {
		t.Error("invalid recv")
	}
}
