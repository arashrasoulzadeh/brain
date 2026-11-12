package plugin_test

import (
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
