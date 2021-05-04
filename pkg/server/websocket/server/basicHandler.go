package xserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/gorilla/websocket"
)

func (s *WsServer) basicHandler(w http.ResponseWriter, r *http.Request) {
	var logger = s.serverOptions.logger.With(xlog.FieldMethod("handler"))

	var clientID = r.Header.Get("X-Ws-ClientID")
	if clientID == "" {
		logger.Error("invalid client id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var now = time.Now()
	var subprotocols = websocket.Subprotocols(r)
	// 升级连接协议
	var upgrader = websocket.Upgrader{
		CheckOrigin:       func(*http.Request) bool { return true },
		Subprotocols:      subprotocols,
		EnableCompression: true,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("upgrade", xlog.FieldErr(err))
		w.WriteHeader(http.StatusBadRequest)
		conn.Close()
		return
	}

	// 生成连接ID
	var id = s.RandStringBytesMaskImprSrcUnsafe()
	var connID = fmt.Sprintf("%d-%s-%s", now.Unix(), clientID, id)
	var _, ok = s.sockets.Load(connID)

	var n int
	for ok {
		if n >= 3 {
			logger.Error("gen conn id", xlog.FieldErr(err))
			w.WriteHeader(http.StatusInternalServerError)
			conn.Close()
			return
		}

		id = s.RandStringBytesMaskImprSrcUnsafe()
		connID = fmt.Sprintf("%d-%s-%s", now.Unix(), clientID, id)
		_, ok = s.sockets.Load(connID)
		n++
	}

	go s.addSocket(r.URL.Path, connID, conn)
}
