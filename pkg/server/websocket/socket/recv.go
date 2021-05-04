package xsocket

import (
	"time"

	"github.com/douyu/jupiter/pkg/xlog"
)

func (s *Socket) defaultOnRecvErr(err error) {
	s.opts.logger.Error("receive msg", xlog.FieldMethod("receive"), xlog.FieldErr(err))
	s.Stop()
}

func (s *Socket) recv() {
	// 在此处初始化 heartbeat ticker 的原因是
	// 会在后面 reset heartbeat ticker
	s.opts.hbTicker = time.NewTicker(s.opts.heartbeatInterval)
	go s.checkHeartbeat()

	for {
		select {
		// 如果关闭了
		case <-s.Stopped():
			return

		default:
			var _, data, err = s.conn.ReadMessage()
			if err != nil {
				for _, f := range s.opts.onRecvErr {
					f(err)
				}
				continue
			}

			// 解压缩
			if s.opts.recvCompressor != nil {
				data, err = s.opts.recvCompressor.Uncompress(data)
				if err != nil {
					s.opts.logger.Error("uncompress", xlog.FieldMethod("receive"), xlog.FieldErr(err))
					continue
				}
			}

			go s.heartbeat(data)

			if s.opts.onRecvMsg != nil {
				var msgToSend = s.opts.onRecvMsg(data)
				if msgToSend != nil {
					s.Send(msgToSend)
				}
			}
		}
	}
}
