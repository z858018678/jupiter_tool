package xwsclient

import (
	"context"
	"errors"
	"time"

	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/gorilla/websocket"
)

func newSendBody(msg []byte, errChan chan error) *sendBody {
	var s sendBody
	s.msg = msg
	if errChan == nil {
		return &s
	}

	s.getErr = func() error { return <-errChan }
	s.setErr = func(err error) { errChan <- err }
	return &s
}

func (c *WsClient) Write(ctx context.Context, msg []byte) {
	select {
	case c.sendChan <- newSendBody(msg, nil):
	case <-ctx.Done():
	}
}

// 同步返回发送消息错误
func (c *WsClient) Send(ctx context.Context, msg []byte) error {
	var ec = make(chan error, 1)
	var b = newSendBody(msg, ec)
	select {
	case c.sendChan <- b:
		return b.getErr()

	case <-ctx.Done():
	}

	return nil
}

func (c *WsClient) defaultOnSendErr(err error) {
	if err == nil {
		return
	}

	c.options.logger.Error("send msg", xlog.FieldMethod("send"), xlog.FieldErr(err))
	c.pause()
}

func (c *WsClient) send() {
	for {
		select {
		// 如果断开了
		case <-*c.disconnected:
			select {
			// 结束
			case <-c.Stopped():
			// 等待连接
			case <-*c.connected:
			}

			// 如果关闭了
		case <-c.Stopped():
			return

		// 如果收到消息
		case msgBody := <-c.sendChan:
			select {
			// 等待连接
			case <-*c.disconnected:
				msgBody.setErr(errors.New("connection invalid"))
				continue

			// 结束
			case <-c.Stopped():
				msgBody.setErr(errors.New("client closed"))
				return

			default:
				var msg = msgBody.msg

				if c.options.beforeSendMsg != nil {
					c.options.beforeSendMsg(msg)
				}

				// 压缩
				if c.options.sendCompressor != nil {
					var data, err = c.options.sendCompressor.Compress(msg)
					if err != nil {
						c.options.logger.Error("compress", xlog.FieldMethod("send"), xlog.FieldErr(err))
						continue
					}

					msg = data
				}

				var err = c.WriteMessage(websocket.BinaryMessage, msg)

				if msgBody.setErr != nil {
					go msgBody.setErr(err)
				}

				for _, f := range c.options.onSendErr {
					f(err)
				}

				if err != nil {
					continue
				}

				if c.options.afterSendMsg != nil {
					c.options.afterSendMsg(msg)
				}
			}
		}
	}
}

func (c *WsClient) ping() {
	if len(c.options.pingData) == 0 {
		return
	}

	var ticker = time.NewTicker(c.options.pingInterval)

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

		// 到期
		case <-ticker.C:
			select {
			// 等待连接
			case <-*c.disconnected:
				continue

			// 结束
			case <-c.options.ctx.Done():
				return

			// ping
			default:
				c.Write(c.options.ctx, c.options.pingData)
			}
		}
	}

}
