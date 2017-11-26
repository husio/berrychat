package chat

import (
	"context"
	"fmt"
	"log"
)

// compatible with gorilla's websocket connection
type ClientConnection interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	Close() error
}

func HandleClient(ctx context.Context, chat Chat, conn ClientConnection) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer conn.Close()

	out := make(chan Message, 8)

	user := chat.CreateUser(out)
	defer chat.RemoveUser(user)

	log.Printf("CLIENT CONNECT %s", user.UserID())
	defer log.Printf("CLIENT DISCONNECT %s", user.UserID())

	go func() {
		for {
			select {
			case m := <-out:
				if err := conn.WriteMessage(textMessage, m.JSON()); err != nil {
					log.Printf("ERR %s: cannot write message: %v", user.UserID(), err)
					cancel()
				}
			case <-ctx.Done():
				// deplete output buffer before quitting
				for {
					select {
					case m := <-out:
						if err := conn.WriteMessage(textMessage, m.JSON()); err != nil {
							log.Printf("ERR %s: cannot write message: %v", user.UserID(), err)
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
			log.Printf("ERR %s: cannot read message: %v", user.UserID(), err)
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
			// continue processing
		}

		msg, err := ParseMessage(raw)
		if err != nil {
			emsg := ErrMessage(fmt.Sprintf("invalid message: %v", err))
			if err := user.Send(ctx, emsg); err != nil {
				log.Printf("ERR %s: cannot send message: %v", user.UserID(), err)
				return
			}
			continue handleNextMessage
		}

		log.Printf("message: %q %v", msg.Kind(), msg.Content())

		if err := user.Send(ctx, OKMessage()); err != nil {
			log.Printf("ERR %s: cannot send message: %v", user.UserID(), err)
		}
	}
}

const textMessage = 1
