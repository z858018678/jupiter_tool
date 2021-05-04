package xtcp

import (
	"time"

	"github.com/douyu/jupiter/pkg/xlog"
)

func (c *TcpClient) connect() error {
	// 开启新的连接
	var conn, err = c.d.DialContext(c.options.ctx, "tcp", c.addr)

	// 连接失败
	if err != nil {
		c.options.logger.Error("connect", xlog.FieldErr(err))
		return err
	}

	// 连接成功
	c.Conn = conn

	// 恢复连接
	c.restore()
	for _, f := range c.options.onConnect {
		go f()
	}

	return nil
}

// 建立 tcp 连接
func (c *TcpClient) reconnectMonitor() {

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

			// 等待重连
			time.Sleep(c.options.reconnectInterval)

			c.connect()
		}
	}
}

// pause 断开连接
func (c *TcpClient) pause() {
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
func (c *TcpClient) restore() {
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
