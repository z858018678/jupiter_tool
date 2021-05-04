package xserver

import (
	"net"
	"net/http"

	xroom "github.com/douyu/jupiter/pkg/server/websocket/room"
	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/gorilla/websocket"
)

// 创建一个 ws 服务
func NewServer(addr string, options ...ServerOption) *WsServer {
	var ws WsServer
	ws.addr = addr
	ws.serverOptions = DefaultServerOptions()
	ws.serverOptions.Apply(options...)
	ws.randomStringGenerator = NewRandomStringGenerator(8)

	// 设置 logger
	ws.serverOptions.logger = ws.serverOptions.logger.With(
		xlog.FieldMod("ws"),
		xlog.FieldType("server"),
	)

	return &ws
}

// 创建一个群组
// 可以通过群组对群组内的 socket 连接群发消息
func (w *WsServer) addRoom(id string, room *xroom.Room) {
	var logger = w.serverOptions.logger.With(xlog.String("room", id))
	w.rooms.Store(id, room)
	logger.Info("add room")
	go func(id string) {
		<-room.Stopped()
		w.rooms.Delete(id)
		logger.Info("del room")
	}(id)
}

// 添加一个 socket 连接
func (w *WsServer) addSocket(path, id string, conn *websocket.Conn) {
	var logger = w.serverOptions.logger.With(xlog.String("socket", id))
	var room, ok = w.GetRoomByPath(path)
	if !ok {
		logger.Error("path room not found", xlog.String("path", path))
		conn.Close()
		return
	}

	var socket = room.NewSocket(id, conn)
	w.sockets.Store(id, socket)
	if w.serverOptions.onNewSocket != nil {
		w.serverOptions.onNewSocket(socket)
	}
	<-socket.Stopped()
	w.sockets.Delete(id)
}

func (s *WsServer) NewRoom(options ...xroom.RoomOption) *xroom.Room {
	var id = s.RandStringBytesMaskImprSrcUnsafe()
	var _, ok = s.rooms.Load(id)
	for ok {
		id = s.RandStringBytesMaskImprSrcUnsafe()
		_, ok = s.sockets.Load(id)
	}

	var opts []xroom.RoomOption
	opts = append(opts, xroom.WithContext(s.serverOptions.ctx))
	opts = append(opts, xroom.WithLogger(s.serverOptions.logger))
	opts = append(opts, options...)

	var room = xroom.NewRoom(id, opts...)
	room.Run()
	s.addRoom(id, room)
	return room
}

// 注册 ws handler
// 一个 path 对应一个群组
func (s *WsServer) RegisterPath(path string, options ...xroom.RoomOption) *xroom.Room {
	s.serverOptions.handlers[path] = http.HandlerFunc(s.basicHandler)

	var room = s.NewRoom(options...)

	s.pathRooms.Store(path, room.GetID())
	s.serverOptions.logger.Info("room handler", xlog.String("room", room.GetID()), xlog.String("path", path))
	go func(path string) {
		<-room.Stopped()
		s.pathRooms.Delete(path)
	}(path)

	return room
}

// 运行 ws 服务
func (s *WsServer) Run() error {
	var logger = s.serverOptions.logger.With(xlog.FieldAddr(s.addr))
	listenHTTP, err := net.Listen("tcp", s.addr)
	if err != nil {
		logger.Panic("new server", xlog.FieldErr(err))
	}

	logger.Info("new server")
	var mux = http.NewServeMux()
	for k, v := range s.serverOptions.handlers {
		mux.Handle(k, v)
	}

	go http.Serve(listenHTTP, mux)
	return nil
}

func (s *WsServer) Stop() error {
	s.serverOptions.cancel()
	return nil
}

// 依据 room id 获取 room
func (s *WsServer) GetRoomByID(id string) (*xroom.Room, bool) {
	var room, ok = s.rooms.Load(id)
	if !ok {
		return nil, false
	}
	return room.(*xroom.Room), true
}

// 依据 path 来获取 room
func (s *WsServer) GetRoomByPath(path string) (*xroom.Room, bool) {
	var id, ok = s.pathRooms.Load(path)
	if !ok {
		return nil, false
	}

	return s.GetRoomByID(id.(string))
}

// 列出所有 room
func (s *WsServer) ListRooms() []*xroom.Room {
	var rooms []*xroom.Room
	s.rooms.Range(func(_, v interface{}) bool {
		rooms = append(rooms, v.(*xroom.Room))
		return true
	})
	return rooms
}
