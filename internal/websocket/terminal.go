package websocket

import (
	"context"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

func TerminalHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer func(conn *websocket.Conn, code websocket.StatusCode, reason string) {
		err := conn.Close(code, reason)
		if err != nil {

		}
	}(conn, websocket.StatusNormalClosure, "")

	for {
		_, msg, err := conn.Read(context.Background())
		if err != nil {
			log.Println(err)
			return
		}

		_ = msg
	}
}
