package xtcp

import (
	"net"

	"github.com/douyu/jupiter/pkg/xlog"
)

func NewClient(addr string, options ...ClientOption) *TcpClient {
	var tc TcpClient
	var connected = make(chan struct{})
	var disconnected = make(chan struct{})
	var opts []ClientOption
	opts = append(opts, WithOnRecvErr(tc.defaultOnRecvErr))
	opts = append(opts, WithOnSendErr(tc.defaultOnSendErr))
	opts = append(opts, options...)

	tc.addr = addr
	tc.options = DefaultClientOptions()
	tc.options.Apply(opts...)

	// 设置 logger
	tc.options.logger = tc.options.logger.With(
		xlog.FieldMod("tcp"),
		xlog.FieldType("client"),
	)

	tc.d = net.Dialer{Timeout: tc.options.dialerTimeout}
	tc.connected = &connected
	tc.disconnected = &disconnected

	tc.sendChan = make(chan *sendBody, tc.options.sendChanBuffer)

	return &tc
}

func (c *TcpClient) Run() error {
	// 首次连接
	if err := c.connect(); err != nil {
		return err
	}

	// 重连
	go c.reconnectMonitor()
	// 发消息
	go c.send()
	// 收消息
	go c.recv()
	// ping
	go c.ping()
	return nil
}

// 关闭 TCP 客户端
func (c *TcpClient) Stop() error {
	c.options.cancel()
	return nil
}

func (c *TcpClient) Stopped() <-chan struct{} { return c.options.ctx.Done() }

func (c *TcpClient) WithOnConnect(funcs ...func()) {
	c.options.onConnect = append(c.options.onConnect, funcs...)
}
