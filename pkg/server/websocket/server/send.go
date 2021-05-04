package xserver

import (
	xroom "github.com/douyu/jupiter/pkg/server/websocket/room"
	xsocket "github.com/douyu/jupiter/pkg/server/websocket/socket"
)

func (s *WsServer) SendToSocket(msg []byte, ids ...string) {
	for _, id := range ids {
		var socket, ok = s.sockets.Load(id)
		if !ok {
			continue
		}
		go socket.(*xsocket.Socket).Send(msg)
	}
}

func (s *WsServer) SendToRoom(msg []byte, ids ...string) {
	for _, id := range ids {
		var room, ok = s.rooms.Load(id)
		if !ok {
			continue
		}
		go room.(*xroom.Room).Broadcast(msg)
	}
}
