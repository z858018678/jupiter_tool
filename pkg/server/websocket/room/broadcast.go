package xroom

import xsocket "github.com/douyu/jupiter/pkg/server/websocket/socket"

func (r *Room) Broadcast(msg []byte) {
	r.opts.broadcastChan <- msg
}

func (r *Room) broadcastMonitor() {
	for msg := range r.opts.broadcastChan {
		r.sockets.Range(func(key, value interface{}) bool {
			go value.(*xsocket.Socket).Send(msg)
			return true
		})
	}
}
