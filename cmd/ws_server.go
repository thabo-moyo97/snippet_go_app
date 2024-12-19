package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"log/slog"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

func newUpgrader(logger *slog.Logger) *websocket.Upgrader {
	u := websocket.NewUpgrader()

	u.CheckOrigin = func(request *http.Request) bool {
		host, _, err := net.SplitHostPort(request.RemoteAddr)
		if err != nil {
			return false
		}
		logger.Info("checking origin", "host", host)
		return host == "127.0.0.1" || host == "::1"
	}

	u.OnOpen(func(c *websocket.Conn) {
		logger.Info("websocket opened", "addr", c.RemoteAddr().String())
	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		logger.Info("websocket message received",
			"type", messageType,
			"data", string(data))
		err := c.WriteMessage(messageType, data)
		if err != nil {
			return
		}
	})
	u.OnClose(func(c *websocket.Conn, err error) {
		logger.Info("websocket closed",
			"addr", c.RemoteAddr().String(),
			"error", err)
	})

	return u
}

func onWebsocket(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	conn, err := newUpgrader(logger).Upgrade(w, r, nil)

	if err != nil {
		logger.Error("websocket upgrade failed", "error", err)
	}
	logger.Info("connection upgraded", "addr", conn.RemoteAddr().String())
}

func main() {
	mux := &http.ServeMux{}

	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		onWebsocket(w, r, slog.Default())
	})

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
