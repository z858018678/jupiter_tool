package xtcp

import (
	"fmt"
	"testing"
)

// 登录 http://tcplab.openluat.com/ 获取 tcp 服务地址之后测试
func TestTcpClient(t *testing.T) {
	var receiver, handler = NewDefaultReceiver(1)

	var c = NewClient(
		"180.97.81.180:58235",
		// WithMsgDelim([]byte("\n")),
		WithOnConnect(func() {
			fmt.Printf("Connected\n")
		}),
		WithMsgHandler(func(b []byte) []byte {
			fmt.Printf("Read msg: %s\n", b)
			return append([]byte("Receive: "), b...)
		}),
		WithMsgHandler(handler),
		WithReconnect(true),
	)

	c.Run()

	go func() {
		for {
			var msg = receiver()
			fmt.Printf("Read msg: %s\n", msg)
			var err = c.Send(append([]byte("Receive: "), msg...))
			if err != nil {
				fmt.Printf("Send failed: %v\n", err)
			}
		}
	}()
	<-make(chan struct{})
}
