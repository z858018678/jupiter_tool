package xserver

import (
	"sync"
)

type WsServer struct {
	addr string
	*randomStringGenerator

	serverOptions *ServerOptions
	sockets       sync.Map
	// key: path, value: room_id
	pathRooms sync.Map
	// key: room_id, value: rooms
	rooms sync.Map
}
