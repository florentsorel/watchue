package stream_test

import (
	"testing"
	"time"

	"github.com/florentsorel/watchue/internal/stream"
)

func TestHub_PublishDeliversToSubscriber(t *testing.T) {
	hub := stream.NewHub()
	ch, unsubscribe := hub.Subscribe()
	defer unsubscribe()

	hub.Publish([]byte("hello"))

	select {
	case msg := <-ch:
		if string(msg) != "hello" {
			t.Errorf("got %q, want %q", msg, "hello")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestHub_PublishFanOutToAllSubscribers(t *testing.T) {
	hub := stream.NewHub()
	ch1, unsub1 := hub.Subscribe()
	defer unsub1()
	ch2, unsub2 := hub.Subscribe()
	defer unsub2()

	hub.Publish([]byte("hi"))

	for _, ch := range []<-chan []byte{ch1, ch2} {
		select {
		case msg := <-ch:
			if string(msg) != "hi" {
				t.Errorf("got %q, want %q", msg, "hi")
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for message")
		}
	}
}

func TestHub_UnsubscribeStopsDeliveryAndClosesChannel(t *testing.T) {
	hub := stream.NewHub()
	ch, unsubscribe := hub.Subscribe()
	unsubscribe()

	hub.Publish([]byte("should not arrive"))

	select {
	case msg, ok := <-ch:
		if ok {
			t.Errorf("received %q after unsubscribe, want closed channel", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for channel to close")
	}
}

func TestHub_PublishWithNoSubscribersDoesNotBlock(t *testing.T) {
	hub := stream.NewHub()
	done := make(chan struct{})
	go func() {
		hub.Publish([]byte("into the void"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked with no subscribers")
	}
}

func TestHub_SlowSubscriberGetsMessagesDroppedNotBlocked(t *testing.T) {
	hub := stream.NewHub()
	ch, unsubscribe := hub.Subscribe()
	defer unsubscribe()

	// fill the subscriber's buffer without ever reading it
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			hub.Publish([]byte("msg"))
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Publish blocked on a slow/unread subscriber instead of dropping")
	}

	// the channel should still have at least one buffered message we can read
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("expected at least one buffered message to still be readable")
	}
}
