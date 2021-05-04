package xtcp

import (
	"errors"
	"time"

	"github.com/douyu/jupiter/pkg/xlog"
)

// 发送消息

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

func (c *TcpClient) Write(msg []byte) {
	c.sendChan <- newSendBody(msg, nil)
}

// 同步返回发送消息错误
func (c *TcpClient) Send(msg []byte) error {
	var ec = make(chan error, 1)
	var b = newSendBody(msg, ec)
	c.sendChan <- b
	return b.getErr()
}

func (c *TcpClient) defaultOnSendErr(err error) {
	if err == nil {
		return
	}
	c.options.logger.Error("send msg", xlog.FieldMethod("send"), xlog.FieldErr(err))
	c.pause()
}

func (c *TcpClient) send() {
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
				c.options.logger.Debug("send msg", xlog.FieldMethod("send"), xlog.FieldValue(string(msg)))
				msg = append(msg, c.options.msgDelim...)
				var _, err = c.Conn.Write(msg)

				if msgBody.setErr != nil {
					go msgBody.setErr(err)
				}

				for _, f := range c.options.onSendErr {
					f(err)
				}
			}
		}
	}
}

func (c *TcpClient) ping() {
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
				c.Write(c.options.pingData)
			}
		}
	}
}
