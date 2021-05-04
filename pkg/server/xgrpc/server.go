// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xgrpc

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/douyu/jupiter/pkg/constant"
	"github.com/douyu/jupiter/pkg/ecode"
	"github.com/douyu/jupiter/pkg/server"
	"github.com/douyu/jupiter/pkg/xlog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// Server ...
type Server struct {
	*grpc.Server
	listener        net.Listener
	gatewayListener net.Listener
	gatewayMux      *runtime.ServeMux
	*Config
}

func newServer(config *Config) *Server {
	var s Server
	s.Config = config
	var streamInterceptors = append(
		[]grpc.StreamServerInterceptor{defaultStreamServerInterceptor(config.logger, config.SlowQueryThresholdInMilli)},
		config.streamInterceptors...,
	)

	var unaryInterceptors = append(
		[]grpc.UnaryServerInterceptor{defaultUnaryServerInterceptor(config.logger, config.SlowQueryThresholdInMilli)},
		config.unaryInterceptors...,
	)

	config.serverOptions = append(config.serverOptions,
		grpc.StreamInterceptor(StreamInterceptorChain(streamInterceptors...)),
		grpc.UnaryInterceptor(UnaryInterceptorChain(unaryInterceptors...)),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			// 连接最大空闲时长
			MaxConnectionIdle: time.Second * 30,
			// 连接无活动之后，达到这个时间后，服务会检查连接是否 alive
			Time: time.Second * 30,
			// 服务检查连接 alive 之后，达到这个时间之后，连接如果依然不活跃，服务会关闭连接
			Timeout: time.Second * 5,
		}),
	)

	newServer := grpc.NewServer(config.serverOptions...)
	s.Server = newServer

	listener, err := net.Listen(config.Network, config.Address())
	if err != nil {
		config.logger.Panic("new grpc server err", xlog.FieldErrKind(ecode.ErrKindListenErr), xlog.FieldErr(err))
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port
	s.listener = listener

	if config.gwRegister != nil && config.GatewayPort != 0 {
		gwListener, err := net.Listen(config.Network, config.GwAddress())
		if err != nil {
			config.logger.Panic("new grpc gateway server err", xlog.FieldErrKind(ecode.ErrKindListenErr), xlog.FieldErr(err))
		}

		var muxOptions = []runtime.ServeMuxOption{
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
				EmitDefaults: true,
				EnumsAsInts:  true,
			}),
		}

		if config.withMetadataFunc != nil {
			muxOptions = append(muxOptions, runtime.WithMetadata(config.withMetadataFunc))
		}

		var mux = runtime.NewServeMux(muxOptions...)

		err = config.gwRegister(context.Background(), mux, config.Address(), []grpc.DialOption{grpc.WithInsecure()})
		if err != nil {
			config.logger.Panic("register grpc gateway server err", xlog.FieldErrKind(ecode.ErrKindListenErr), xlog.FieldErr(err))
		}

		config.GatewayPort = gwListener.Addr().(*net.TCPAddr).Port
		s.gatewayListener = gwListener
		s.gatewayMux = mux

		config.logger.Info("start gateway", xlog.FieldEvent("init"), xlog.FieldAddr(config.GwAddress()), xlog.String("scheme", "http"))
	}

	return &s
}

// Server implements server.Server interface.
func (s *Server) Serve() error {
	var errChan = make(chan error, 1)
	if s.gatewayListener != nil {
		go func() {
			errChan <- http.Serve(s.gatewayListener, s.gatewayMux)
		}()
	}

	go func() {
		errChan <- s.Server.Serve(s.listener)
	}()

	return <-errChan
}

// Stop implements server.Server interface
// it will terminate echo server immediately
func (s *Server) Stop() error {
	s.Server.Stop()
	return nil
}

// GracefulStop implements server.Server interface
// it will stop echo server gracefully
func (s *Server) GracefulStop(ctx context.Context) error {
	s.Server.GracefulStop()
	return nil
}

// Info returns server info, used by governor and consumer balancer
func (s *Server) Info() *server.ServiceInfo {
	serviceAddress := s.listener.Addr().String()
	if s.Config.ServiceAddress != "" {
		serviceAddress = s.Config.ServiceAddress
	}

	var gwAddress string
	if s.gatewayListener != nil {
		gwAddress = s.gatewayListener.Addr().String()
	}

	info := server.ApplyOptions(
		server.WithScheme("grpc"),
		server.WithAddress(serviceAddress),
		server.WithAddressGW(gwAddress),
		server.WithKind(constant.ServiceProvider),
		server.WithWeight(s.Weight),
	)
	return &info
}
