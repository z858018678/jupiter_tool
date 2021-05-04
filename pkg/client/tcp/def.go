package xtcp

import (
	"net"
)

type sendBody struct {
	msg    []byte
	setErr func(error)
	getErr func() error
}

type TcpClient struct {
	// tcp 连接地址
	addr string

	// tcp 连接
	d net.Dialer
	net.Conn

	// 断开信号
	disconnected *chan struct{}
	// 连接信号
	connected *chan struct{}

	// 发消息通道
	sendChan chan *sendBody

	// 每次读完消息，剩下的消息碎片
	lastMsg []byte

	options *ClientOptions
}
