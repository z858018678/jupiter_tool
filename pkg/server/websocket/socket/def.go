package xsocket

import "github.com/gorilla/websocket"

type Socket struct {
	id         string
	remoteAddr string
	conn       *websocket.Conn

	opts *SocketOptions

	sendChan chan []byte
}

type SocketInfo struct {
	ID         string
	RemoteAddr string
}
