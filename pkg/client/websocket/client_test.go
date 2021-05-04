package xwsclient

import (
	"fmt"
	"testing"
	"time"

	"github.com/douyu/jupiter/pkg/util/xcompressor"
)

// 测试 ws 服务：https://www.idcd.com/tool/socket
func TestClient(t *testing.T) {
	var compressor = xcompressor.NewZlibCompressor(5)
	for i := 0; i < 20; i++ {
		var c = NewClient(
			fmt.Sprintf("ws://127.0.0.1:60100/ws%d/", i%2+1),
			WithDialTimeout(time.Second*5),
			WithOnRecvMsg(func(data []byte) []byte {
				fmt.Printf("Recieve %d bytes: %s\n", len(data), data)
				if string(data) == "pong" {
					return []byte("i am ws client.")
				}
				return nil
			}),
			// WithOnConnect(func() { fmt.Printf("Client Connected\n") }),
			WithReconnect(true),
			// WithPing([]byte("ping"), time.Second*20),
			WithBeforeSendMsg(func(data []byte) {
				fmt.Printf("Send %d bytes msg: %s\n", len(data), data)
			}),
			WithAfterSendMsg(func(data []byte) {
				fmt.Printf("Send success %d bytes\n", len(data))
			}),
			WithCompressor(compressor, compressor),
		)

		var err = c.Run()
		if err != nil {
			panic(err)
		}
	}

	<-make(chan struct{})
}
