package xtcp

import (
	"context"
	"time"

	"github.com/douyu/jupiter/pkg/xlog"
)

type ClientOptions struct {
	// 连接生命管理
	ctx    context.Context
	cancel context.CancelFunc
	// 发送消息通道缓冲
	sendChanBuffer int
	sendErrBuffer  int

	// 连接开启时要运行的任务
	// 重连时也会运行
	onConnect []func()
	// 消息处理器
	// 用以处理被分隔符分开的完整消息
	msgHandler         func([]byte) []byte
	recvMsgChanHandler func(chan []byte)
	// 处理接收消息时发生的错误
	onRecvErr []func(error)
	// 处理发送消息时发生的错误
	onSendErr []func(error)
	// 一次循环接受消息最大时长
	recvMsgDuration time.Duration
	// 每次读消息的最大长度
	recvMsgLen int
	// 完整消息分隔符
	msgDelim []byte
	// 连接超时
	dialerTimeout time.Duration

	// ping
	pingData     []byte
	pingInterval time.Duration

	mustReconnect     bool
	reconnectInterval time.Duration

	logger *xlog.Logger
}

// 默认配置
func DefaultClientOptions() *ClientOptions {
	var c ClientOptions
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.sendChanBuffer = 1
	c.sendErrBuffer = 1
	c.recvMsgDuration = time.Second
	c.recvMsgLen = 64
	c.dialerTimeout = time.Second * 10
	c.reconnectInterval = time.Second
	c.logger = xlog.JupiterLogger
	return &c
}

func (o *ClientOptions) Apply(options ...ClientOption) {
	for _, opt := range options {
		opt.apply(o)
	}
}

type ClientOption interface {
	apply(*ClientOptions)
}

type funcClientOption func(*ClientOptions)

func (o funcClientOption) apply(option *ClientOptions) {
	o(option)
}

func newFuncClientOption(f func(*ClientOptions)) *funcClientOption {
	var fo = funcClientOption(f)
	return &fo
}

func WithContext(ctx context.Context) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.ctx, o.cancel = context.WithCancel(ctx)
	})
}

// 发消息通道缓冲
// 如果不设置，则为 1
func WithSendChanBuffer(buffer int) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.sendChanBuffer = buffer
		o.sendErrBuffer = buffer
	})
}

// 连接成功时启动的任务
func WithOnConnect(funcs ...func()) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.onConnect = append(o.onConnect, funcs...)
	})
}

// 注册方法用以在 tcp 客户端接收到消息时处理错误
func WithOnRecvErr(f ...func(error)) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.onRecvErr = append(o.onRecvErr, f...)
	})
}

// 注册方法用以在 tcp 客户端发送消息时处理错误
func WithOnSendErr(f ...func(error)) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.onSendErr = append(o.onSendErr, f...)
	})
}

// 注册消息处理器
// 用以处理被分隔符分开的完整消息
func WithMsgHandler(f func([]byte) []byte) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.msgHandler = f
	})
}

// 每次读消息的时长
// 如果不设置，则为 1s
func WithRecvMsgDuration(t time.Duration) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.recvMsgDuration = t
	})
}

// 每次读消息的最大长度
// 如果不设置，则为 64
func WithRecvMsgLen(l int) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.recvMsgLen = l
	})
}

// 完整消息之间的分隔符
func WithMsgDelim(b []byte) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.msgDelim = b
	})
}

// 连接超时
// 如果不设置，则为 10s
func WithDialTimeout(t time.Duration) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.dialerTimeout = t
	})
}

// 添加 ping 的配置
func WithPing(ping []byte, interval time.Duration) ClientOption {
	if interval < time.Second*5 {
		interval = time.Second * 5
	}

	return newFuncClientOption(func(o *ClientOptions) {
		o.pingData = ping
		o.pingInterval = interval
	})
}

// 断线重连
func WithReconnect(on bool) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.mustReconnect = on
	})
}

// 断线重连失败后重试间隔
// 如果不设置，则为 1s
// 范围: >= 500ms
func WithReconnectInterval(d time.Duration) ClientOption {
	if d < time.Millisecond*500 {
		d = time.Millisecond * 500
	}

	return newFuncClientOption(func(o *ClientOptions) {
		o.reconnectInterval = d
	})
}

func WithLogger(l *xlog.Logger) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.logger = l
	})
}
