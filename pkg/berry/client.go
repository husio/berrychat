package berry

import (
	"context"
	"errors"
	"fmt"
)

// compatible with gorilla's websocket connection
type ClientConnection interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
}

func HandleClient(ctx context.Context, chat Chat, conn ClientConnection) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make(chan Message, 8)

	user := chat.CreateUser(out)
	defer chat.RemoveUser(user)

	go func() {
		for {
			select {
			case m := <-out:
				if err := conn.WriteMessage(textMessage, m.JSON()); err != nil {
					cancel()
				}
			case <-ctx.Done():
				// deplete output buffer before quitting
				for {
					select {
					case m := <-out:
						if err := conn.WriteMessage(textMessage, m.JSON()); err != nil {
							return
						}
					default:
						return
					}
				}
			}
		}
	}()

handleNextMessage:
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("cannot read message: %s", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// continue processing
		}

		msg, err := ParseMessage(raw)
		if err != nil {
			emsg := ErrMessage(fmt.Sprintf("invalid message: %v", err))
			if err := user.Send(ctx, emsg); err != nil {
				return fmt.Errorf("cannot send message: %s", err)
			}
			continue handleNextMessage
		}

		var response Message
		if handle, ok := messageHandler[msg.Kind()]; !ok {
			response = ErrMessage("unknown message")
		} else {
			if err := handle(ctx, chat, user, msg); err != nil {
				response = ErrMessage(err.Error())
			} else {
				response = OKMessage()
			}
		}

		if err := user.Send(ctx, response); err != nil {
			return fmt.Errorf("cannot send message: %s", err)
		}
	}
}

const textMessage = 1

var messageHandler = map[string]func(context.Context, Chat, User, Message) error{
	"nick": func(ctx context.Context, chat Chat, user User, msg Message) error {
		return errors.New("not implemented")
	},
	"join": func(ctx context.Context, chat Chat, user User, msg Message) error {
		chat.Room(msg.Content()[0]).Subscribe(user)
		return nil
	},
	"quit": func(ctx context.Context, chat Chat, user User, msg Message) error {
		chat.Room(msg.Content()[0]).Unsubscribe(user)
		return nil
	},
	"say": func(ctx context.Context, chat Chat, user User, msg Message) error {
		m, err := NewMessage("msg", user.UserID(), msg.Content()[0], msg.Content()[1])
		if err != nil {
			return err
		}
		return chat.Room(msg.Content()[0]).Broadcast(ctx, m)
	},
	"ping": func(ctx context.Context, chat Chat, user User, msg Message) error {
		return nil
	},
	"ok": func(ctx context.Context, chat Chat, user User, msg Message) error {
		return nil
	},
	"err": func(ctx context.Context, chat Chat, user User, msg Message) error {
		return nil
	},
}
