package xwsclient

import (
	"time"

	"github.com/douyu/jupiter/pkg/xlog"
)

func (c *WsClient) defaultOnConnected() {
	c.options.logger.Info("connected")
}

// 建立 tcp 连接
func (c *WsClient) connect() error {
	// 开启新的连接
	var conn, _, err = c.DialContext(c.options.ctx, c.addr, c.options.header)

	// 连接失败
	if err != nil {
		c.options.logger.Error("dial websocket failed", xlog.FieldErr(err))
		c.pause()
		return err
	}

	c.Conn = conn
	c.SetReadLimit(c.options.readLimit)

	// 恢复连接
	c.restore()
	for _, f := range c.options.onConnect {
		go f()
	}

	return nil
}

// pause 断开连接
func (c *WsClient) pause() {
	// 检测连接信号
	select {
	// 如果显示连接正常
	case <-*c.connected:
		// 停止连接正常信号
		*c.connected = make(chan struct{})
	default:
	}

	// 检测断开连接信号
	select {
	case <-*c.disconnected:
	// 如果连接未断开
	default:
		// 断开连接
		close(*c.disconnected)
	}
}

// 恢复连接
func (c *WsClient) restore() {
	// 检测断开连接信号
	select {
	// 如果连接已断开
	case <-*c.disconnected:
		// 关闭断开信号
		*c.disconnected = make(chan struct{})
	default:
	}

	// 检测连接信号
	select {
	case <-*c.connected:
	// 如果未发送连接正常信号
	default:
		// 发送连接正常信号
		close(*c.connected)
	}
}

func (c *WsClient) reconnectMonitor() {
	for {
		select {
		// 服务停止，断开连接
		case <-c.Stopped():
			// 关闭连接
			if c.Conn != nil {
				c.Conn.Close()
			}

			return

		// 如果连接断开了
		case <-*c.disconnected:
			// 关闭连接
			if c.Conn != nil {
				c.Conn.Close()
			}

			if !c.options.mustReconnect {
				c.Stop()
				continue
			}

			time.Sleep(time.Second * 3)

			// 尝试重连
			c.connect()
		}
	}
}
