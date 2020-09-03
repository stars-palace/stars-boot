package xgrpc

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stars-palace/stars-boot/pkg/server"
	"github.com/stars-palace/stars-boot/pkg/xconst"
	"github.com/stars-palace/statrs-common/pkg/xlogger"
	"google.golang.org/grpc"
	"net"
	"reflect"
	"runtime"
	"time"
)

/**
 * grpc服务处理类
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/4 3:22 下午
 * @updateDate           2020/8/4 3:22 下午
 * @version              1.0
**/

// Server ...
//grpc 服务的结构体
type Server struct {
	*grpc.Server
	listener net.Listener
	*Config
}

//create server ...
//创建服务
func newServer(config *Config) *Server {
	//流拦截器的切片
	var streamInterceptors = append(
		[]grpc.StreamServerInterceptor{defaultStreamServerInterceptor(config.logger, config.SlowQueryThresholdInMilli)},
		config.streamInterceptors...,
	)
	//一元拦截器的切片
	var unaryInterceptors = append(
		[]grpc.UnaryServerInterceptor{defaultUnaryServerInterceptor(config.logger, config.SlowQueryThresholdInMilli)},
		config.unaryInterceptors...,
	)
	//追加拦截器链
	config.serverOptions = append(config.serverOptions,
		grpc.StreamInterceptor(StreamInterceptorChain(streamInterceptors...)),
		grpc.UnaryInterceptor(UnaryInterceptorChain(unaryInterceptors...)),
	)

	//创建grpc服务
	newServer := grpc.NewServer(config.serverOptions...)
	//创建服务监听指定协议和地址
	listener, err := net.Listen(config.Network, config.Address())
	if err != nil {
		config.logger.Panic("new grpc server err", xlogger.FieldErrKind(xconst.ErrKindListenErr), xlogger.FieldErr(err))
	}
	//设置地址信心
	config.Port = listener.Addr().(*net.TCPAddr).Port
	//创建服务
	return &Server{Server: newServer, listener: listener, Config: config}
}

// Server implements server.Server interface.
//服务，实现服务和接口
func (s *Server) Serve() error {
	//打印输出grpc服务启动成功
	s.colorer.Println(fmt.Sprintf("⇨ grpc server started on %s", s.colorer.Green(s.Config.Address())))
	//开启grpc服务
	err := s.Server.Serve(s.listener)
	if err == grpc.ErrServerStopped {
		return nil
	}
	return err
}

// Stop implements server.Server interface
// it will terminate grpc server immediately
//停止具体服务。服务接口
//将立即停止grpc服务
func (s *Server) Stop() error {
	s.Server.Stop()
	return nil
}

// GracefulStop implements server.Server interface
// it will stop grpc server gracefully
//优雅的停止服务，服务接口
//将优雅的停止grpc服务
func (s *Server) GracefulStop(ctx context.Context) error {
	s.Server.GracefulStop()
	return nil
}

// Info returns server info, used by governor and consumer balancer
// 初始化服务信息
func (s *Server) Info(group, cluster string) *server.ServiceInfo {
	rpcParam := &server.ServiceInfo{
		Name:        s.Name,
		Scheme:      s.Name,
		IP:          s.Host,
		Port:        s.Port,
		Weight:      s.Weight,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		GroupName:   group,
		ClusterName: cluster,
	}
	return rpcParam
}

//写一个服务注册的方法
//srv 服务的实现
//in 服务的接口
func (s *Server) Register(in interface{}, srv interface{}) {
	//获取类型
	elem := reflect.TypeOf(in).Elem()
	if elem.Kind() != reflect.Interface {
		panic(fmt.Sprintf("注册grpc服务的%s必须是一个接口\n", elem.Kind().String()))
	}
	srvType := reflect.TypeOf(srv).Elem()
	if !srvType.Implements(elem) {
		panic(fmt.Sprintf("注册grpc服务的%s必须实现接口%s\n ", srvType.Kind().String(), elem.Kind().String()))
	}
	//定义服务调用的地址
	stype := fmt.Sprintf("%s.%s.Grpc", elem.PkgPath(), elem.Name())
	s.logger.Info(fmt.Sprintf("rpc add server [path:/%s]", stype))
	//服务处理
	var _Grpc_ServiceDesc = grpc.ServiceDesc{
		ServiceName: stype,
		HandlerType: (*interface{})(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "HugoGrpc",
				Handler:    _Index_Handler,
			},
		},
		Streams: []grpc.StreamDesc{},
	}
	//第一个参数为处理的ServiceDesc
	s.RegisterService(&_Grpc_ServiceDesc, srv)
}

//默认的流拦截器定义
func defaultStreamServerInterceptor(logger *logrus.Logger, slowQueryThresholdInMilli int64) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		var beg = time.Now()
		var event = "normal"
		defer func() {
			if slowQueryThresholdInMilli > 0 {
				if int64(time.Since(beg))/1e6 > slowQueryThresholdInMilli {
					event = "slow"
				}
			}

			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				case error:
					err = rec
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
			}
			if err != nil {
				logger.Error("access", err)
				return
			}
			logger.Info("access", "未出现异常")
		}()
		return handler(srv, stream)
	}
}

// 默认的一元拦截定义
func defaultUnaryServerInterceptor(logger *logrus.Logger, slowQueryThresholdInMilli int64) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		//服务调用开始时间
		var beg = time.Now()
		var event = "normal"
		defer func() {
			if slowQueryThresholdInMilli > 0 {
				if int64(time.Since(beg))/1e6 > slowQueryThresholdInMilli {
					event = "slow"
				}
			}
			//错误恢复
			if rec := recover(); rec != nil {
				switch rec := rec.(type) {
				case error:
					err = rec
				default:
					err = fmt.Errorf("%v", rec)
				}

				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
			}
			if err != nil {
				logger.Error("access", err)
				return
			}
			logger.Info("access", "未出现异常")
		}()
		return handler(ctx, req)
	}
}
