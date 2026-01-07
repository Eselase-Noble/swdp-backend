package websocket

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/moby/moby/api/types"
	"github.com/moby/moby/client"
)

func TerminalHandler(docker *client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		workspaceID := chi.URLParam(r, "workspaceID")
		containerName := "ws-" + workspaceID

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx := context.Background()

		execResp, err := docker.ContainerExecCreate(
			ctx,
			containerName,
			types.ExecConfig{
				Cmd:          []string{"/bin/sh"},
				AttachStdin:  true,
				AttachStdout: true,
				AttachStderr: true,
				Tty:          true,
			},
		)
		if err != nil {
			log.Println("exec create:", err)
			return
		}

		hijack, err := docker.ContainerExecAttach(
			ctx,
			execResp.ID,
			types.ExecStartCheck{Tty: true},
		)
		if err != nil {
			log.Println("exec attach:", err)
			return
		}
		defer hijack.Close()

		// WS → Container stdin
		go func() {
			for {
				_, msg, err := conn.Read(ctx)
				if err != nil {
					return
				}
				_, _ = hijack.Conn.Write(msg)
			}
		}()

		// Container stdout → WS
		buf := make([]byte, 1024)
		for {
			n, err := hijack.Reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Println("container read:", err)
				}
				return
			}
			_ = conn.Write(ctx, websocket.MessageText, buf[:n])
		}
	}
}
