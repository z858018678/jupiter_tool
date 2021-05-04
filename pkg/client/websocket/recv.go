package xwsclient

import (
	"github.com/douyu/jupiter/pkg/xlog"
)

func (c *WsClient) defaultOnRecvErr(err error) {
	c.options.logger.Error("receive msg", xlog.FieldMethod("receive"), xlog.FieldErr(err))
	c.pause()
}

func (c *WsClient) recv() {
	for {
		select {
		// 如果断开了
		case <-*c.disconnected:
			select {
			// 结束
			case <-c.options.ctx.Done():
			// 等待连接
			case <-*c.connected:
			}

			// 如果关闭了
		case <-c.options.ctx.Done():
			return

		default:
			var _, data, err = c.ReadMessage()
			if err != nil {
				for _, f := range c.options.onRecvErr {
					f(err)
				}
				continue
			}

			// 解压缩
			if c.options.recvCompressor != nil {
				data, err = c.options.recvCompressor.Uncompress(data)
				if err != nil {
					c.options.logger.Error("uncompress", xlog.FieldMethod("receive"), xlog.FieldErr(err))
					continue
				}
			}

			if c.options.onRecvMsg != nil {
				var msgToSend = c.options.onRecvMsg(data)
				if msgToSend != nil {
					c.Send(c.options.ctx, msgToSend)
				}
			}
		}
	}
}
