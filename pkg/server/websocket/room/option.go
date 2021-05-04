package xroom

import (
	"context"
	"net/http"

	xsocket "github.com/douyu/jupiter/pkg/server/websocket/socket"
	"github.com/douyu/jupiter/pkg/xlog"
)

type RoomOptions struct {
	logger        *xlog.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	socketOptions []xsocket.SocketOption
	handler       http.Handler

	broadcastChanBuffer int
	broadcastChan       chan []byte
}

type RoomOption interface{ apply(*RoomOptions) }

func DefaultRoomOptions() *RoomOptions {
	var s RoomOptions
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.logger = xlog.JupiterLogger

	return &s
}

func (o *RoomOptions) Apply(options ...RoomOption) {
	for _, opt := range options {
		opt.apply(o)
	}
}

type funcRoomOption func(*RoomOptions)

func (f funcRoomOption) apply(o *RoomOptions) {
	f(o)
}

func newFuncRoomOption(f func(o *RoomOptions)) RoomOption {
	return funcRoomOption(f)
}

func WithContext(ctx context.Context) RoomOption {
	return newFuncRoomOption(func(o *RoomOptions) {
		o.ctx, o.cancel = context.WithCancel(ctx)
	})
}

func WithLogger(l *xlog.Logger) RoomOption {
	return newFuncRoomOption(func(o *RoomOptions) {
		o.logger = l
	})
}

func WithBroadcastChanBuffer(buffer int) RoomOption {
	return newFuncRoomOption(func(o *RoomOptions) {
		o.broadcastChanBuffer = buffer
	})
}

func WithSocketOptions(opts ...xsocket.SocketOption) RoomOption {
	return newFuncRoomOption(func(o *RoomOptions) {
		o.socketOptions = append(o.socketOptions, opts...)
	})
}
