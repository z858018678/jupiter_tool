package xroom

import "sync"

type Room struct {
	id      string
	sockets sync.Map
	opts    *RoomOptions
}
