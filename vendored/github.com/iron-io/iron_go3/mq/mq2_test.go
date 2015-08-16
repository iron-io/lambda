package mq

import (
	"fmt"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	// use a queue named "test_queue" to push/get messages
	q := New("test_queue")
	err := q.Clear()
	if err != nil {
		t.Error("Unexpected error in clearing a queue: ", err)
	}

	_, err = q.PushString("Hello, World!")
	if err != nil {
		t.Error("Unexpected error in pushing a message: ", err)
	}

	// You can also pass multiple messages in a single call.
	ids, err := q.PushStrings("Message 1", "Message 2")
	if err != nil {
		t.Error("Unexpected error in pushing a message: ", err)
	}
	if len(ids) != 2 {
		t.Error("Expected 2 id got: ", len(ids))
	}

	msgs, err := q.GetN(100)
	if err != nil {
		t.Error("Unexpected error while dequeueing", err)
	}
	if len(msgs) != 3 {
		t.Error("Expected 3 got: ", len(msgs))
	}
	q.Clear()
}

func TestQueueSize(t *testing.T) {
	q := New("queuename")
	q.Clear()
	strings := []string{}
	for n := 0; n < 100; n++ {
		strings = append(strings, fmt.Sprint("test: ", n))
	}

	ids, err := q.PushStrings(strings...)
	info, err := q.Info()
	if err != nil {
		t.Error("Unexpected error in getting qinfo: ", err)
	}
	if info.Size != 100 {
		t.Error("Expected 100 in size got: ", info.Size)
	}

	for i := 0; i < 10; i++ {
		err := q.DeleteMessage(ids[i], "0")
		if err != nil {
			t.Error("Unexpected error while deleting message: ", err)
		}
	}
	info, err = q.Info()
	if err != nil {
		t.Error("")
	}

	msgs, err := q.GetN(90)
	if err != nil {
		t.Error("Unexpected error while getting message: ", err)
	}
	if len(msgs) != 90 {
		t.Error("Expected to be able to pull 90 message got: ", len(msgs))
	}

	for i := 0; i < 10; i++ {
		err := q.DeleteMessage(msgs[i].Id, msgs[i].ReservationId)
		if err != nil {
			t.Error("Unexpected error while deleting message: ", err)
		}
	}
	info, err = q.Info()
	if err != nil {
		t.Error("Unexpected error in getting qinfo: ", err)
	}

	if info.Size != 80 {
		t.Error("Expected 80 in size got: ", info.Size)
	}

	err = q.Clear()
	if err != nil {
		t.Error("Unexpected error in clearing queue", err)
	}

	info, err = q.Info()
	if err != nil {
		t.Error("Unexpected error in getting qinfo: ", err)
	}
	if info.Size != 0 {
		t.Error("Expected 0 in size got: ", info.Size)
	}
}

func TestRelease(t *testing.T) {
	q := New("queuename")

	_, err := q.PushString("trying")
	if err != nil {
		t.Error("Unexpected error while pushing message: ", err)
	}

	msg, err := q.Get()
	if err != nil {
		t.Error("Unexpected error while getting message: ", err)
	}

	err = msg.Release(3)
	if err != nil {
		t.Error("Unexpected error while releasing message: ", err)
	}
	msg, err = q.Get()
	if err != nil {
		t.Error("Unexpected error while getting message: ", err)
	}
	//if msg != nil {
	//t.Error("Should have not released message within delay: ", msg)
	//}

	time.Sleep(4 * time.Second)
	msg, err = q.Get()
	if err != nil {
		t.Error("Unexpected error while getting message: ", err)
	}
	//if msg == nil {
	//t.Error("Should have released message: ", msg)
	//}
}
