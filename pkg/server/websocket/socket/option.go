package xsocket

import (
	"context"
	"time"

	"github.com/douyu/jupiter/pkg/util/xcompressor"
	"github.com/douyu/jupiter/pkg/xlog"
)

type SocketOptions struct {
	ctx    context.Context
	cancel context.CancelFunc
	logger *xlog.Logger

	// 读到消息时的执行项
	onRecvMsg func([]byte) []byte
	// 读消息时发生错误的执行项
	onRecvErr []func(error)

	// 写入消息前的执行项
	beforeSendMsg func([]byte)
	// 写入消息后的执行项
	afterSendMsg func([]byte)
	// 写消息时发生错误的执行项
	onSendErr []func(error)

	// 数据压缩、解压缩方式
	sendCompressor xcompressor.XCompressor
	recvCompressor xcompressor.XCompressor

	// 发送消息通道
	sendChan       chan []byte
	sendChanBuffer int

	// 检测心跳间隔
	heartbeatMsg      []byte
	heartbeatInterval time.Duration
	hbTicker          *time.Ticker
}

type SocketOption interface{ apply(*SocketOptions) }

func DefaultSocketOptions() *SocketOptions {
	var s SocketOptions
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.sendChanBuffer = 1
	s.sendChan = make(chan []byte, 1)
	s.logger = xlog.JupiterLogger
	s.heartbeatInterval = time.Second * 30
	s.heartbeatMsg = []byte("ping")
	return &s
}

func (o *SocketOptions) Apply(options ...SocketOption) {
	for _, opt := range options {
		opt.apply(o)
	}
}

type funcSocketOption func(*SocketOptions)

func (f funcSocketOption) apply(o *SocketOptions) {
	f(o)
}

func newFuncSocketOption(f func(o *SocketOptions)) SocketOption {
	return funcSocketOption(f)
}

// 发送消息通道缓冲
// 默认 1
func WithSendChanBuffer(size int) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.sendChanBuffer = size
		o.sendChan = make(chan []byte, o.sendChanBuffer)
	})
}

func WithLogger(l *xlog.Logger) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.logger = l
	})
}

func WithContext(ctx context.Context) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.ctx, o.cancel = context.WithCancel(ctx)
	})
}

// 注册收到消息时的处理项
func WithOnRecvMsg(f func([]byte) []byte) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.onRecvMsg = f
	})
}

// 注册收到消息时发生错误的处理项
func WithOnRecvErr(f func(error)) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.onRecvErr = append(o.onRecvErr, f)
	})
}

// 注册写入消息前的处理项
func WithBeforeSendMsg(f func([]byte)) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.beforeSendMsg = f
	})
}

// 注册写入消息后的处理项
func WithAfterSendMsg(f func([]byte)) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.afterSendMsg = f
	})
}

// 注册写入消息时发生错误的处理项
func WithOnSendErr(f func(error)) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.onSendErr = append(o.onSendErr, f)
	})
}

// 数据压缩
func WithCompressor(send, recv xcompressor.XCompressor) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.sendCompressor, o.recvCompressor = send, recv
	})
}

// 设置心跳检测
func WithHeartbeat(data []byte, interval time.Duration) SocketOption {
	return newFuncSocketOption(func(o *SocketOptions) {
		o.heartbeatMsg, o.heartbeatInterval = data, interval
	})
}
