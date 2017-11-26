package chat

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"
)

func TestHandleClientConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	chat := NewChat()
	conn := &testClientConnection{
		Err: nil,
		ResRead: []string{
			`["ping"]`,
			`["nick", "bobby"]`,
			`["join", "dark-room"]`,
			`["nick", "Bobby"]`,
			`["say", "dark-room", "hello there!"]`,
			`["quit", "dark-room"]`,
		},
	}

	HandleClient(ctx, chat, conn)

	// wait for the handler to finish - flush all writes
	time.Sleep(50 * time.Millisecond)

	conn.Lock()
	defer conn.Unlock()

	if got, want := len(conn.Written), 6; got != want {
		t.Errorf("want %d messages to be written by client, got %d", want, got)
	}
}

type testClientConnection struct {
	sync.Mutex

	ResRead []string
	Written []string
	Err     error
}

var _ ClientConnection = (*testClientConnection)(nil)

func (c *testClientConnection) ReadMessage() (int, []byte, error) {
	c.Lock()
	defer c.Unlock()

	if c.Err != nil {
		return 0, nil, c.Err
	}

	if len(c.ResRead) == 0 {
		return 0, nil, io.EOF
	}

	raw := c.ResRead[0]
	c.ResRead = c.ResRead[1:]

	return textMessage, []byte(raw), nil
}

func (c *testClientConnection) WriteMessage(messageType int, body []byte) error {
	c.Lock()
	defer c.Unlock()

	c.Written = append(c.Written, string(body))
	return c.Err
}

func (c *testClientConnection) Close() error {
	c.Lock()
	defer c.Unlock()
	return c.Err
}
