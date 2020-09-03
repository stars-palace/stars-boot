package xhttp

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/stars-palace/stars-boot/pkg/server"
	"github.com/stars-palace/stars-boot/pkg/xconst"
	"github.com/stars-palace/statrs-common/pkg/xlogger"
	"log"
	"net"
	"net/http"
	"os"
)

// http Server struct
type Server struct {
	*echo.Echo
	Config   *Config
	listener net.Listener
}

func newServer(config *Config) *Server {
	listener, err := net.Listen("tcp", config.Address())
	if err != nil {
		config.logger.Panic("new xecho server err", xlogger.FieldErrKind(xconst.ErrKindListenErr), xlogger.FieldErr(err))
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port
	return &Server{
		Echo:     echo.New(),
		Config:   config,
		listener: listener,
	}
}

// Server implements server.Server interface.
func (s *Server) Serve() error {
	s.Echo.Logger.SetOutput(os.Stdout)
	s.Echo.Debug = s.Config.Debug
	s.Echo.HideBanner = true
	//TODO std日志
	s.Echo.StdLogger = log.New(os.Stdout, "hugo", 1)
	for _, route := range s.Echo.Routes() {
		//输出地址信息和处理的方法
		s.Config.logger.Info("add route", xlogger.FieldMethod(route.Method), xlogger.String("path", route.Path))
	}
	s.Echo.Listener = s.listener
	err := s.Echo.Start("")
	if err != http.ErrServerClosed {
		return err
	}

	s.Config.logger.Info("close echo", xlogger.FieldAddr(s.Config.Address()))
	return nil
}

// Stop implements server.Server interface
// it will terminate echo server immediately
//停止具体服务。服务接口
//将立即停止echo服务
func (s *Server) Stop() error {
	return s.Echo.Close()
}

// GracefulStop implements server.Server interface
// it will stop echo server gracefully
//优雅的停止服务，服务接口
//将优雅的停止echo服务
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.Echo.Shutdown(ctx)
}

// Info returns server info, used by governor and consumer balancer
// 初始化服务信息
func (s *Server) Info(group, cluster string) *server.ServiceInfo {
	//注http服务
	httpSeverConfig := s.Config
	httpParam := &server.ServiceInfo{
		Name:        httpSeverConfig.Name,
		Scheme:      httpSeverConfig.Name,
		IP:          httpSeverConfig.Host,
		Port:        httpSeverConfig.Port,
		Weight:      httpSeverConfig.Weight,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		GroupName:   group,
		ClusterName: cluster,
	}
	return httpParam
}

//向服务中注册控制器
func (s *Server) UseController(con Controller) {
	//调用controller的注册方法将接口注册到系统中
	con.Register(s)
}
