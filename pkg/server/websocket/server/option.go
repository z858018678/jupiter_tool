package xserver

import (
	"context"
	"net/http"

	xsocket "github.com/douyu/jupiter/pkg/server/websocket/socket"
	"github.com/douyu/jupiter/pkg/xlog"
)

type ServerOptions struct {
	logger      *xlog.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	handlers    map[string]http.Handler
	onNewSocket func(*xsocket.Socket)
}

type ServerOption interface{ apply(*ServerOptions) }

func DefaultServerOptions() *ServerOptions {
	var s ServerOptions
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.handlers = make(map[string]http.Handler)
	s.logger = xlog.JupiterLogger

	return &s
}

func (o *ServerOptions) Apply(options ...ServerOption) {
	for _, opt := range options {
		opt.apply(o)
	}
}

type funcServerOption func(*ServerOptions)

func (f funcServerOption) apply(o *ServerOptions) {
	f(o)
}

func newFuncServerOption(f func(o *ServerOptions)) ServerOption {
	return funcServerOption(f)
}

func WithContext(ctx context.Context) ServerOption {
	return newFuncServerOption(func(o *ServerOptions) {
		o.ctx, o.cancel = context.WithCancel(ctx)
	})
}

func WithLogger(l *xlog.Logger) ServerOption {
	return newFuncServerOption(func(o *ServerOptions) {
		o.logger = l
	})
}

// 有新客户端连接时的回调
func WithOnNewSocket(f func(*xsocket.Socket)) ServerOption {
	return newFuncServerOption(func(o *ServerOptions) {
		o.onNewSocket = f
	})
}
