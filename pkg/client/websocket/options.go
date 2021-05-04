package xwsclient

import (
	"context"
	"net/http"
	"time"

	"github.com/douyu/jupiter/pkg/util/xcompressor"
	"github.com/douyu/jupiter/pkg/xlog"
)

type ClientOptions struct {
	ctx           context.Context
	cancel        context.CancelFunc
	header        http.Header
	mustReconnect bool

	// 连接超时
	dialerTimeout time.Duration

	logger *xlog.Logger

	// 连接时的执行项
	onConnect []func()

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

	// 读取数据的最大长度
	readLimit int64

	// 发送消息通道缓冲
	sendChanBuffer int

	// 数据压缩、解压缩方式
	sendCompressor xcompressor.XCompressor
	recvCompressor xcompressor.XCompressor

	// ping 消息
	pingData []byte
	// ping 间隔
	pingInterval time.Duration
}

func (o *ClientOptions) Apply(options ...ClientOption) {
	for _, opt := range options {
		opt.apply(o)
	}
}

func DefaultClientOptions() *ClientOptions {
	var c ClientOptions
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.logger = xlog.JupiterLogger
	c.sendChanBuffer = 1
	c.header = make(http.Header)

	return &c
}

type ClientOption interface {
	apply(*ClientOptions)
}

type funcClientOption func(*ClientOptions)

func (f funcClientOption) apply(o *ClientOptions) {
	f(o)
}

func newFuncClientOption(f func(*ClientOptions)) ClientOption {
	return funcClientOption(f)
}

func WithLogger(l *xlog.Logger) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.logger = l
	})
}

// 客户端生命周期
func WithContext(ctx context.Context) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.ctx, o.cancel = context.WithCancel(ctx)
	})
}

// 连接时的 header
func WithHeader(header http.Header) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		for k, v := range header {
			for _, s := range v {
				o.header.Add(k, s)
			}
		}
	})
}

// 是否需要断线重连
func WithReconnect(on bool) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.mustReconnect = on
	})
}

// 连接成功时要运行的任务
func WithOnConnect(funcs ...func()) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.onConnect = append(o.onConnect, funcs...)
	})
}

// 注册收到消息时的处理项
func WithOnRecvMsg(f func([]byte) []byte) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.onRecvMsg = f
	})
}

// 注册收到消息时发生错误的处理项
func WithOnRecvErr(f func(error)) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.onRecvErr = append(o.onRecvErr, f)
	})
}

// 注册写入消息前的处理项
func WithBeforeSendMsg(f func([]byte)) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.beforeSendMsg = f
	})
}

// 注册写入消息后的处理项
func WithAfterSendMsg(f func([]byte)) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.afterSendMsg = f
	})
}

// 注册写入消息时发生错误的处理项
func WithOnSendErr(f func(error)) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.onSendErr = append(o.onSendErr, f)
	})
}

// 设置读取数据的最大长度
func WithRecvMsgMaxBytes(max int64) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.readLimit = max
	})
}

// 连接超时
// 如果不设置，则为 10s
func WithDialTimeout(t time.Duration) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.dialerTimeout = t
	})
}

// 发送消息通道缓冲
func WithSendChanBuffer(size int) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.sendChanBuffer = size
	})
}

// 数据压缩
func WithCompressor(send, recv xcompressor.XCompressor) ClientOption {
	return newFuncClientOption(func(o *ClientOptions) {
		o.sendCompressor, o.recvCompressor = send, recv
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
