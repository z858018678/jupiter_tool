package xtcp

import (
	"bytes"
	"errors"
	"os"
	"time"

	"github.com/douyu/jupiter/pkg/xlog"
)

// 处理接受到的完整消息
func (c *TcpClient) defaultMsgHandler(msg []byte) {
	if c.options.msgHandler != nil {
		var data = c.options.msgHandler(msg)
		if data == nil {
			return
		}
		c.Send(data)
	}
}

func (c *TcpClient) defaultOnRecvErr(err error) {
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return
	}

	c.options.logger.Error("receive msg", xlog.FieldMethod("receive"), xlog.FieldErr(err))
	c.pause()
}

func (c *TcpClient) defaultOnRecvMsg(msg []byte) {
	var n = len(msg)
	// 如果消息为空
	if n == 0 {
		return
	}

	if len(c.options.msgDelim) == 0 {
		c.defaultMsgHandler(msg)
		return
	}

	// 预留完整的数据空间用来保存上次残留的数据与本次接受到的数据
	var thisData = make([]byte, len(c.lastMsg)+n)
	// 首部填入残留数据
	copy(thisData[:len(c.lastMsg)], c.lastMsg)
	// 尾部填入本次接受到的数据
	copy(thisData[len(c.lastMsg):], msg)

	// 尝试使用 c.msgTail 分隔数据，判断是否接受到了一条完整的数据
	var msgs = bytes.Split(thisData, c.options.msgDelim)

	var l = len(msgs)
	// 更新残留数据
	c.lastMsg = msgs[l-1]

	// 处理分隔出来的完整数据
	for _, msg := range msgs[:l-1] {
		// 这里不能使用 goroutine
		// 不然消息顺序与发送方的发消息顺序可能不一致
		c.defaultMsgHandler(msg)
	}

	// 如果以本次接受到的消息不以 c.msgTail 结尾
	// 说明消息有一部分未结束
	// 继续下次接收消息，直到接收到 c.msgTail
	if !bytes.HasSuffix(msg, c.options.msgDelim) {
		return
	}

	// 否则，说明残留的消息也为一条完成的消息
	c.defaultMsgHandler(c.lastMsg)
	// 清空残留的消息
	c.lastMsg = nil
}

func (c *TcpClient) recv() {
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
			c.SetReadDeadline(time.Now().Add(c.options.recvMsgDuration))

			// 按定长数据去读消息
			// 直到读到消息分隔符
			// 分隔消息，多余的消息放到缓存中
			for {
				var data = make([]byte, c.options.recvMsgLen)
				var n, err = c.Conn.Read(data)
				if err != nil {
					for _, f := range c.options.onRecvErr {
						f(err)
					}
					break
				}

				if n == 0 {
					continue
				}

				c.defaultOnRecvMsg(data[:n])
			}
		}
	}
}
