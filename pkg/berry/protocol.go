package berry

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Each message is JSON array of string. First element is always request
// message id. Second is always message kind.
// Number of additional array elements depend on the message. When send, each
// message is followed by \r\n
//
// Client can send one of the following:
//
//    ["<request id>", "nick", "<new nick>"]
//
//    ["<request id>", "join", "<room id>"]
//
//    ["<request id>", "quit", "<room id>"]
//
//    ["<request id>", "say", "<room id>",  "<body>"]
//
//    ["<request id>", "ping"]
//
//    ["<request id>", "ok"]
//
//  Server response to any message is always either OK or ERR. First argument
//  is always ID of the request it is responding to:
//
//    ["<request id>", "ok"]
//    ["<request id>", "err", "<description>"]
//
// Server can send one of the following. If user send response, it is ignored
// by the server:
//
//    ["<message id>", "ping"]
//
//    ["<message id>", "msg", "<user id>", "<room id>", "<unix time>", "<body>"]
//
//    ["<message id>", "joined", "<user id>", "<room id>"]
//
//    ["<message id>", "parted", "<user id>", "<room id>"]
//
//    ["<message id>", "renamed", "<user id>", "<new nick>"]
//
//    ["<message id>", "roomusers", "<room id>", "<user 1 id>", "<user 2 id>", ...]
//
// Responses comes always in the same order as requests.

// declaration for all client messages
// <message kind>: <num of args>
var clientMessageSchema = map[string]int{
	"nick": 1,
	"join": 1,
	"quit": 1,
	"say":  2,
	"ping": 0,
	"ok":   0,
}

func ParseMessage(raw []byte) (Message, error) {
	msg := &message{
		json: raw,
	}
	if err := json.Unmarshal(raw, &msg.data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal message: %s", err)
	}
	switch len(msg.data) {
	case 0:
		return msg, errors.New("empty message")
	case 1:
		return msg, errors.New("incomplete message")
	}

	kind := msg.data[1]
	args, ok := clientMessageSchema[kind]
	if !ok {
		return nil, fmt.Errorf("unknown message kind: %q", kind)
	}

	if len(msg.data)-2 != args {
		return nil, fmt.Errorf("invalid format: message of kind %q required %d arguments, got %d", kind, args, len(msg.data)-2)
	}

	switch n := len(msg.RequestID()); {
	case n > 8:
		return nil, errors.New("invalid format: request ID must not be longer than 8 characters")
	case n < 4:
		return nil, errors.New("invalid format: request ID must be at least 4 characters long")
	}

	return msg, nil
}

func OKMessage(requestID string) Message {
	data := []string{requestID, "ok"}
	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return &message{
		data: data,
		json: raw,
	}
}

func ErrMessage(requestID, description string) Message {
	data := []string{requestID, "err", description}
	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return &message{
		data: data,
		json: raw,
	}
}

func NewMessage(kind string, args ...string) (Message, error) {
	data := make([]string, 0, len(args)+2)
	data = append(data, generateID(), kind)
	data = append(data, args...)

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("cannot JSON serialize: %v", err)
	}
	m := &message{
		data: data,
		json: raw,
	}
	return m, nil
}

type message struct {
	data []string
	json []byte
}

func (m *message) RequestID() string {
	if len(m.data) == 0 {
		return ""
	}
	return m.data[0]
}

func (m *message) Kind() string {
	if len(m.data) < 2 {
		return ""
	}
	return m.data[1]
}

func (m *message) Content() []string {
	return m.data[2:]
}

func (m *message) JSON() []byte {
	return m.json
}
