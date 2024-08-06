package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()

	u.CheckOrigin = func(request *http.Request) bool {
		host, _, err := net.SplitHostPort(request.RemoteAddr)
		if err != nil {
			return false
		}
		fmt.Println("CheckOrigin:", host)
		return host == "127.0.0.1" || host == "::1"
	}

	u.OnOpen(func(c *websocket.Conn) {
		fmt.Println("OnOpen:", c.RemoteAddr().String())
	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		fmt.Println("OnMessage:", messageType, string(data))
		err := c.WriteMessage(messageType, data)
		if err != nil {
			return
		}
	})
	u.OnClose(func(c *websocket.Conn, err error) {
		fmt.Println("OnClose:", c.RemoteAddr().String(), err)
	})

	return u
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := newUpgrader().Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("Failed:", err)
	}
	fmt.Println("Upgraded:", conn.RemoteAddr().String())
}

func main() {
	mux := &http.ServeMux{}

	mux.HandleFunc("GET /ws", onWebsocket)

	engine := nbhttp.NewEngine(nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{"localhost:8888"},
		MaxLoad:                 1000000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	})

	err := engine.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)

		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	engineErr := engine.Shutdown(ctx)
	if engineErr != nil {
		panic(engineErr)
	}
}
