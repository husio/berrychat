package chat

import (
	"context"
	"testing"
	"time"
)

func TestChat(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	chat := NewChat()

	out1 := make(chan Message, 4)
	user1 := chat.CreateUser(out1)

	out2 := make(chan Message, 4)
	user2 := chat.CreateUser(out2)

	room1 := chat.Room("room-1")
	room1.Subscribe(user1)
	defer room1.Unsubscribe(user1)
	room1.Subscribe(user2)

	pingMsg, err := NewMessage("ping")
	if err != nil {
		t.Fatalf("cannot create ping message: %v", err)
	}
	if err := room1.Broadcast(ctx, pingMsg); err != nil {
		t.Fatalf("cannot broadcast ping message: %v", err)
	}

	select {
	case msg := <-out1:
		if msg.Kind() != "ping" {
			t.Errorf("user 1 should get ping message, got %#v", msg)
		}
	default:
		t.Error("user 1 got no message")
	}

	select {
	case msg := <-out2:
		if msg.Kind() != "ping" {
			t.Errorf("user 2 should get ping message, got %#v", msg)
		}
	default:
		t.Error("user 2 got no message")
	}

	room1.Unsubscribe(user2)

	if err := room1.Broadcast(ctx, OKMessage()); err != nil {
		t.Fatalf("cannot broadcast ok message: %v", err)
	}

	select {
	case msg := <-out1:
		if msg.Kind() != "ok" {
			t.Errorf("user 1 should get ok message, got %#v", msg)
		}
	default:
		t.Error("user 1 got no message")
	}

	select {
	case msg := <-out2:
		t.Errorf("user 2 should not get message, got %#v", msg)
	default:
	}
}
