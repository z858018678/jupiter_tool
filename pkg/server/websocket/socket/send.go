package xsocket

import (
	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/gorilla/websocket"
)

// 发送消息
func (s *Socket) Send(msg []byte) {
	s.opts.sendChan <- msg
}

func (s *Socket) defaultOnSendErr(err error) {
	s.opts.logger.Error("send msg", xlog.FieldMethod("send"), xlog.FieldErr(err))
	// 发送失败，直接关闭连接
	s.Stop()
}

func (s *Socket) send() {
	for {
		select {
		// 如果关闭了
		case <-s.Stopped():
			return

		// 如果收到消息
		case msg := <-s.opts.sendChan:
			if s.opts.beforeSendMsg != nil {
				s.opts.beforeSendMsg(msg)
			}

			// 压缩
			if s.opts.sendCompressor != nil {
				var data, err = s.opts.sendCompressor.Compress(msg)
				if err != nil {
					s.opts.logger.Error("send msg", xlog.FieldMethod("send"), xlog.FieldErr(err))
					continue
				}
				msg = data
			}

			var err = s.conn.WriteMessage(websocket.BinaryMessage, msg)
			if err != nil {
				for _, f := range s.opts.onSendErr {
					f(err)
				}
				continue
			}

			if s.opts.afterSendMsg != nil {
				s.opts.afterSendMsg(msg)
			}
		}
	}
}
