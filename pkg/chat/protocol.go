package chat

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Each message is JSON array of string. First element is always message kind.
// Number of additional array elements depend on the message. When send, each
// message is followed by \r\n
//
// Client can send one of the following:
//
//    ["nick", "<new nick>"]
//
//    ["join", "<room id>"]
//
//    ["quit", "<room id>"]
//
//    ["say", "<room id>",  "<body>"]
//
//    ["ping"]
//
//    ["ok"]
//
//  Server response to any message is always either OK or ERR:
//
//    ["ok"]
//    ["err", "<description>"]
//
// Server can send one of the following. If user send response, it is ignored
// by the server:
//
//    ["ping"]
//
//    ["msg", "<user id>", "<room id>", "<unix time>", "<body>"]
//
//    ["joined", "<user id>", "<room id>"]
//
//    ["parted", "<user id>", "<room id>"]
//
//    ["renamed", "<user id>", "<new nick>"]
//
//    ["usernick", "<user id>", "<nick>"]
//
//    ["roomusers", "<room id>", "<user 1 id>", "<user 2 id>", ...]
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
	msg := message{
		json: raw,
	}
	if err := json.Unmarshal(raw, &msg.data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal message: %s", err)
	}
	if len(msg.data) == 0 {
		return msg, errors.New("empty message")
	}

	args, ok := clientMessageSchema[msg.data[0]]
	if !ok {
		return nil, fmt.Errorf("unknown message kind: %q", msg.data[0])
	}

	if len(msg.data)-1 != args {
		return nil, fmt.Errorf("invalid format: message of kind %q required %d arguments, got %d", msg.data[0], args, len(msg.data)-1)
	}

	return msg, nil
}

func OKMessage() Message {
	return &message{
		data: []string{"ok"},
		json: []byte(`["ok"]`),
	}
}

func ErrMessage(description string) Message {
	data := []string{"err", description}
	rawjson, err := json.Marshal(data)
	if err != nil {
		panic("cannot JSON serialize")
	}
	return &message{
		data: data,
		json: rawjson,
	}
}

func NewMessage(args ...string) (Message, error) {
	rawjson, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("cannot JSON serialize: %v", err)
	}
	m := &message{
		data: args,
		json: rawjson,
	}
	return m, nil
}

type message struct {
	data []string
	json []byte
}

func (m message) Kind() string {
	return m.data[0]
}

func (m message) Content() []string {
	return m.data[1:]
}

func (m message) JSON() []byte {
	return m.json
}
