package berry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

type User interface {
	UserID() string
	Nickname() string
	Send(context.Context, Message) error
}

type Room interface {
	RoomID() string
	Users() []User
	UsersCount() int
	Broadcast(context.Context, Message) error
	Subscribe(User)
	Unsubscribe(User)
}

type Message interface {
	Kind() string
	Content() []string
	JSON() []byte
}

type Chat interface {
	CreateUser(out chan<- Message) User
	RemoveUser(User)
	Room(roomID string) Room
	// RenameUser(User, string) error
}

func NewChat() Chat {
	return &chat{
		rooms: make(map[string]Room),
		users: make(map[string]User),
	}
}

type chat struct {
	mu    sync.Mutex
	rooms map[string]Room
	users map[string]User
}

func (c *chat) Room(roomID string) Room {
	c.mu.Lock()
	defer c.mu.Unlock()

	if r, ok := c.rooms[roomID]; ok {
		return r
	}

	r := &room{
		id:    roomID,
		users: make(map[string]User),
	}
	c.rooms[roomID] = r
	return r
}

func (c *chat) CreateUser(out chan<- Message) User {
	id := generateID()
	u := &user{
		id:       id,
		nickname: id,
		out:      out,
	}

	c.mu.Lock()
	c.users[u.UserID()] = u
	c.mu.Unlock()

	return u
}

func (c *chat) RemoveUser(u User) {
	// TODO
}

type room struct {
	id string

	mu    sync.Mutex
	users map[string]User
}

func (r *room) RoomID() string {
	return r.id
}

func (r *room) String() string {
	return fmt.Sprintf("<Room:%s>", r.id)
}

func (r *room) Broadcast(ctx context.Context, msg Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.users {
		if err := u.Send(ctx, msg); err != nil {
			log.Printf("cannot write to user %q: %v", u.UserID(), err)
		}
	}
	return nil
}

func (r *room) Users() []User {
	panic("x")
}

func (r *room) UsersCount() int {
	return 0
}

func (r *room) Subscribe(u User) {
	r.mu.Lock()
	r.users[u.UserID()] = u
	r.mu.Unlock()
}

func (r *room) Unsubscribe(u User) {
	r.mu.Lock()
	delete(r.users, u.UserID())
	r.mu.Unlock()
}

type user struct {
	id       string
	nickname string
	out      chan<- Message
}

func (u *user) UserID() string {
	return u.id
}

func (u *user) Nickname() string {
	return u.nickname
}

func (u *user) Send(ctx context.Context, msg Message) error {
	select {
	case u.out <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return ErrSlowClient
	}
}

func (u *user) String() string {
	return fmt.Sprintf("<User:%s>", u.id)
}

var ErrSlowClient = errors.New("slow client")
