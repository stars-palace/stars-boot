package xconst

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/9/3 12:53 下午
 * @updateDate           2020/9/3 12:53 下午
 * @version              1.0
**/

// 统一 Err Kind
const (
	// ErrKindUnmarshalConfigErr ...
	ErrKindUnmarshalConfigErr = "unmarshal config err"
	ReadAppConfigErr          = "read app config err"
	ReadRegistryConfigErr     = "read registry config err"
	// ErrKindRegisterErr ...
	ErrKindRegisterErr = "register err"
	// ErrKindUriErr ...
	ErrKindUriErr = "uri err"
	// ErrKindRequestErr ...
	ErrKindRequestErr = "request err"
	// ErrKindFlagErr ...
	ErrKindFlagErr = "flag err"
	// ErrKindListenErr ...
	ErrKindListenErr = "listen err"
	// ErrKindAny ...
	ErrKindAny         = "any"
	ErrREQNotMethod    = "请求体中未携带method信息"
	ErrServerException = "调用服务端异常"
	ErrConfigType      = "错误的配置文件类型"
)

// 统一模块信息
const (
	// ModConfig ...
	ModConfig = "config"
	// ModApp ...
	ModApp = "app"
	// ModProc ...
	ModProc = "proc"
	// ModGrpcServer ...
	ModGrpcServer = "server.grpc"
	// ModRegistryETCD ...
	ModRegistryETCD = "registry.etcd"
	// ModClientETCD ...
	ModClientETCD = "client.etcd"
	// ModClientGrpc ...
	ModClientGrpc = "client.grpc"
	// ModClientMySQL ...
	ModClientMySQL = "client.mysql"
	// ModRegistryNacos ...
	ModRegistryNacos = "registry.nacos"
	// ModRegistry
	ModRegistry = "registry"
	ModWork     = "work"
)

//配置中心类型
const (
	Nacos     = "nacos"
	Etcd      = "etcd"
	Zookppeer = "zookppeer"
)

const FrameName = "brian"
const (
	//应用名称
	ApplicationName = FrameName + ".application.name"
	//应用全局日志级别
	ApplicationLoglevel = FrameName + ".log.level"
)
