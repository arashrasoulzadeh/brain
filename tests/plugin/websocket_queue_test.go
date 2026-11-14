package plugin_test

import (
	"sync"
	"testing"

	"brain/internal/plugin"
)

func TestWebSocketSendQueueRejectsOverflowAndPreservesFIFO(t *testing.T) {
	q := plugin.NewWebSocketSendQueue(2)
	if err := q.Enqueue([]byte("one")); err != nil {
		t.Fatal(err)
	}
	if err := q.Enqueue([]byte("two")); err != nil {
		t.Fatal(err)
	}
	if err := q.Enqueue([]byte("three")); err == nil {
		t.Fatal("queue overflow accepted")
	}
	if string(q.Dequeue()) != "one" || string(q.Dequeue()) != "two" {
		t.Fatal("queue did not preserve FIFO")
	}
	if q.Dequeue() != nil {
		t.Fatal("empty queue returned a frame")
	}
}

func TestWebSocketSendQueueSupportsConcurrentProducers(t *testing.T) {
	q := plugin.NewWebSocketSendQueue(32)
	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		index := index
		group.Add(1)
		go func() {
			defer group.Done()
			if err := q.Enqueue([]byte{byte(index)}); err != nil {
				t.Errorf("Enqueue(%d) = %v", index, err)
			}
		}()
	}
	group.Wait()
	for index := 0; index < 32; index++ {
		if q.Dequeue() == nil {
			t.Fatalf("queue lost frame %d", index)
		}
	}
}
