package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/husio/berrychat/pkg/berry"
)

func main() {
	chat := berry.NewChat()

	http.Handle("/ws", &wsHandler{chat: chat})
	http.HandleFunc("/", handleIndex)

	fmt.Println("starting HTTP server: http://localhost:8000/")
	http.ListenAndServe("localhost:8000", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "cmd/berryd/index.html")
}

type wsHandler struct {
	chat berry.Chat
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("cannot upgrade client connection: %s", err)
		return
	}
	defer conn.Close()

	_ = berry.HandleClient(r.Context(), h.chat, conn)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
