package xsocket

import (
	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/gorilla/websocket"
)

func NewSocket(id string, conn *websocket.Conn, options ...SocketOption) *Socket {
	var s Socket
	s.id = id
	s.conn = conn
	s.opts = DefaultSocketOptions()
	var opts []SocketOption
	opts = append(opts, WithOnRecvErr(s.defaultOnRecvErr))
	opts = append(opts, WithOnSendErr(s.defaultOnSendErr))
	opts = append(opts, options...)

	s.opts.Apply(opts...)

	s.remoteAddr = conn.RemoteAddr().String()

	// 设置 logger
	s.opts.logger = s.opts.logger.With(
		xlog.String("socket", id),
		xlog.String("from", s.remoteAddr),
	)

	go func() {
		<-s.Stopped()
		conn.Close()
	}()

	return &s
}

func (s *Socket) Info() *SocketInfo {
	return &SocketInfo{
		ID:         s.id,
		RemoteAddr: s.remoteAddr,
	}
}
