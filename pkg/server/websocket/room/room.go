package xroom

import (
	xsocket "github.com/douyu/jupiter/pkg/server/websocket/socket"
	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/gorilla/websocket"
)

func NewRoom(id string, options ...RoomOption) *Room {
	var r Room
	r.id = id
	r.opts = DefaultRoomOptions()
	r.opts.Apply(options...)

	r.opts.broadcastChan = make(chan []byte, r.opts.broadcastChanBuffer)

	// 设置 logger
	r.opts.logger = r.opts.logger.With(
		xlog.String("room", id),
	)

	return &r
}

func (r *Room) GetID() string {
	return r.id
}

func (r *Room) AddSocket(s *xsocket.Socket) {
	var info = s.Info()
	var logger = r.opts.logger.With(xlog.String("socket", info.ID), xlog.String("from", info.RemoteAddr))
	logger.Info("connnected")
	r.sockets.Store(info.ID, s)
	go func(id string) {
		select {
		case <-r.Stopped():
		case <-s.Stopped():
		}
		r.sockets.Delete(id)
		logger.Info("leave room")
	}(info.ID)
}

func (r *Room) NewSocket(id string, conn *websocket.Conn) *xsocket.Socket {
	var opts []xsocket.SocketOption
	opts = append(opts, xsocket.WithLogger(r.opts.logger))
	opts = append(opts, xsocket.WithContext(r.opts.ctx))
	opts = append(opts, r.opts.socketOptions...)

	var socket = xsocket.NewSocket(id, conn, opts...)
	socket.Run()
	r.AddSocket(socket)
	return socket
}

func (r *Room) Run() error {
	go r.broadcastMonitor()
	return nil
}

func (r *Room) Stop() error {
	r.opts.logger.Info("room close", xlog.FieldName(r.id))
	r.opts.cancel()
	return nil
}

func (r *Room) Stopped() <-chan struct{} {
	return r.opts.ctx.Done()
}

func (r *Room) ListSockets() []*xsocket.Socket {
	var sockets []*xsocket.Socket
	r.sockets.Range(func(_, v interface{}) bool {
		sockets = append(sockets, v.(*xsocket.Socket))
		return true
	})
	return sockets
}
