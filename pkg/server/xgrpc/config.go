package xgrpc

import (
	"fmt"
	"github.com/labstack/gommon/color"
	"github.com/sirupsen/logrus"
	"github.com/stars-palace/statrs-common/pkg/xlogger"
	conf "github.com/stars-palace/statrs-config"
	"google.golang.org/grpc"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/4 3:17 下午
 * @updateDate           2020/8/4 3:17 下午
 * @version              1.0
**/
// TODO 日志需要单独处理
//Grpc 的配置结构体
type Config struct {
	Name string `properties:"brian.rpc.server.name"`
	Host string `properties:"brian.rpc.server.host"`
	Port int    `properties:"brian.rpc.server.port"`
	// Network network type, tcp4 by default
	Network string `properties:"brian.rpc.server.Network"`
	// DisableTrace disbale Trace Interceptor, false by default
	//禁用跟踪器默认为true
	DisableTrace bool `properties:"brian.rpc.server.DisableTrace"`
	// DisableMetric disable Metric Interceptor, false by default
	//禁用监听器默认为true
	DisableMetric bool `properties:"brian.rpc.server.DisableMetric"`
	// SlowQueryThresholdInMilli, request will be colored if cost over this threshold value
	SlowQueryThresholdInMilli int64 `properties:"brian.rpc.server.timeout"`
	//日志级别
	logLevel           string  `properties:"brian.rpc.server.log.level"`
	Weight             float64 `properties:"brian.rpc.server.registry.weight"`
	serverOptions      []grpc.ServerOption
	streamInterceptors []grpc.StreamServerInterceptor
	unaryInterceptors  []grpc.UnaryServerInterceptor
	colorer            *color.Color
	//TODO 日志
	logger *logrus.Logger
}

// DefaultConfig represents default config
// User should construct config base on DefaultConfig
//grpc默认的配置
//用户不做调整则使用默认的配置
func DefaultConfig() *Config {
	return &Config{
		Network:                   "tcp4",
		Host:                      "127.0.0.1",
		Port:                      9090,
		DisableMetric:             true,
		DisableTrace:              true,
		SlowQueryThresholdInMilli: 500,
		logger:                    logrus.New(),
		colorer:                   color.New(),
		serverOptions:             []grpc.ServerOption{},
		//流方法，流拦截器
		streamInterceptors: []grpc.StreamServerInterceptor{},
		//grpc中使用一元拦截器
		unaryInterceptors: []grpc.UnaryServerInterceptor{},
	}
}

// hugo Standard HTTP Server config
func StdConfig() *Config {
	return RawConfig()
}

// RawConfig ...
func RawConfig() *Config {
	var config = DefaultConfig()
	err := conf.UnmarshalToStruct(config)
	if nil != err {
		logrus.Panic("Unmarshal config ", xlogger.FieldMod(xcodec.ModConfig), xlogger.FieldErrKind(xcodec.ErrKindUnmarshalConfigErr), xlogger.FieldErr(err))
	}
	if level, err := logrus.ParseLevel(config.logLevel); nil == err {
		config.logger.Level = level
	}
	return config
}

// Build ...
func (config *Config) Build() *Server {
	//TODO
	if !config.DisableTrace {
		//config.unaryInterceptors = append(config.unaryInterceptors, traceUnaryServerInterceptor)
		//config.streamInterceptors = append(config.streamInterceptors, traceStreamServerInterceptor)
	}

	if !config.DisableMetric {
		//config.unaryInterceptors = append(config.unaryInterceptors, prometheusUnaryServerInterceptor)
		//config.streamInterceptors = append(config.streamInterceptors, prometheusStreamServerInterceptor)
	}

	return newServer(config)
}

// Address ...
//用来获取组装完成的服务地址
func (config Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
