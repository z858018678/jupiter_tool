package xserver

import (
	"context"
	"fmt"
	"testing"
	"time"

	xroom "github.com/douyu/jupiter/pkg/server/websocket/room"
	xsocket "github.com/douyu/jupiter/pkg/server/websocket/socket"
	"github.com/douyu/jupiter/pkg/xlog"
)

func TestServer(t *testing.T) {
	var server = NewServer(
		"127.0.0.1:60100",
		WithContext(context.Background()),
		WithLogger(xlog.JupiterLogger),
	)

	// var compressor = xcompressor.NewZlibCompressor(5)
	var room1 = server.RegisterPath(
		"/ws1/",
		xroom.WithBroadcastChanBuffer(100),
		xroom.WithSocketOptions(
			xsocket.WithOnRecvMsg(func(data []byte) []byte {
				fmt.Printf("Receive %d bytes: %s\n", len(data), data)
				if string(data) == "ping" {
					return []byte("pong")
				}
				return []byte("i am ws server.")
			}),
			xsocket.WithBeforeSendMsg(func(data []byte) {
				fmt.Printf("send msg %d bytes: %s\n", len(data), data)
			}),
			xsocket.WithAfterSendMsg(func(data []byte) {
				fmt.Printf("Send success %d bytes\n", len(data))
			}),
			// xsocket.WithCompressor(compressor, compressor),
		),
	)

	var room2 = server.RegisterPath(
		"/ws2/",
		xroom.WithBroadcastChanBuffer(100),
		xroom.WithSocketOptions(
			xsocket.WithOnRecvMsg(func(data []byte) []byte {
				fmt.Printf("Receive %d bytes: %s\n", len(data), data)
				if string(data) == "ping" {
					return []byte("pong")
				}
				return []byte("i am ws server.")
			}),
			xsocket.WithBeforeSendMsg(func(data []byte) {
				fmt.Printf("send msg %d bytes: %s\n", len(data), data)
			}),
			xsocket.WithAfterSendMsg(func(data []byte) {
				fmt.Printf("Send success %d bytes\n", len(data))
			}),
			// xsocket.WithCompressor(compressor, compressor),
		),
	)

	go server.Run()

	time.Sleep(time.Second * 5)

	room1.Broadcast([]byte("Welcome to room 1."))
	room2.Broadcast([]byte("Welcome to room 2."))
	server.BroadcastRoom([]byte("Welcome to ws server"))

	<-make(chan struct{})
}
