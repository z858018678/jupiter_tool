package xwsclient

import (
	"net/http"

	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/gorilla/websocket"
)

func NewClient(addr string, options ...ClientOption) *WsClient {
	var c WsClient
	c.id = globalGenerator.RandStringBytesMaskImprSrcUnsafe()

	var connected = make(chan struct{})
	var disconnected = make(chan struct{})

	c.addr = addr
	c.options = DefaultClientOptions()

	var opts []ClientOption
	opts = append(opts, WithOnRecvErr(c.defaultOnRecvErr))
	opts = append(opts, WithOnSendErr(c.defaultOnSendErr))
	opts = append(opts, WithOnConnect(c.defaultOnConnected))
	opts = append(opts, WithHeader(http.Header{"X-Ws-ClientID": []string{c.id}}))
	opts = append(opts, options...)

	c.options.Apply(opts...)

	// 设置 logger
	c.options.logger = c.options.logger.With(
		xlog.FieldMod("ws"),
		xlog.String("client", c.id),
	)

	// 设置 dialer
	c.Dialer = websocket.Dialer{HandshakeTimeout: c.options.dialerTimeout}

	c.sendChan = make(chan *sendBody, c.options.sendChanBuffer)
	c.connected = &connected
	c.disconnected = &disconnected

	return &c
}

func (c *WsClient) Run() error {
	// 开启连接
	if err := c.connect(); err != nil {
		return err
	}

	// 是否要断线重连重连
	go c.reconnectMonitor()
	// 接受消息
	go c.recv()
	// 发送消息
	go c.send()
	// ping
	go c.ping()

	return nil
}

func (c *WsClient) Stop() error {
	select {
	case <-c.Stopped():
	default:
		c.options.cancel()
	}
	return nil
}

func (c *WsClient) Stopped() <-chan struct{} {
	return c.options.ctx.Done()
}

func (c *WsClient) GetID() string {
	return c.id
}
