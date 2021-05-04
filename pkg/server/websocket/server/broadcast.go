package xserver

import xroom "github.com/douyu/jupiter/pkg/server/websocket/room"

// 广播给所有的 room
// room 会广播给所有 room 中的 socket
func (s *WsServer) BroadcastRoom(msg []byte) {
	s.rooms.Range(func(key, value interface{}) bool {
		go value.(*xroom.Room).Broadcast(msg)
		return true
	})
}
