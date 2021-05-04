package xwsclient

import (
	"github.com/gorilla/websocket"
)

type sendBody struct {
	msg    []byte
	setErr func(error)
	getErr func() error
}

type WsClient struct {
	id string

	addr    string
	options *ClientOptions

	disconnected *chan struct{}
	connected    *chan struct{}

	websocket.Dialer
	*websocket.Conn

	// 发送消息通道
	sendChan chan *sendBody
}
