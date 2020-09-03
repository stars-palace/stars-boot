package registry

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stars-palace/stars-boot/pkg/server"
	"github.com/stars-palace/stars-boot/pkg/xconst"
	"github.com/stars-palace/statrs-common/pkg/xlogger"
	conf "github.com/stars-palace/statrs-config"
	"io"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/20 6:03 下午
 * @updateDate           2020/8/20 6:03 下午
 * @version              1.0
**/
// ServerInstance ...
type ServerInstance struct {
	Scheme string
	IP     string
	Port   int
	Labels map[string]string
}

// RegistryConfig 的服务配置
type RegistryConfig struct {
	Type                 string `properties:"brian.registry.type"`                 //注册中心类型
	ContextPath          string `properties:"brian.registry.ContextPath"`          //the  server contextpath
	IpAddr               string `properties:"brian.registry.address"`              //the  server address
	Port                 uint64 `properties:"brian.registry.port"`                 //the  server port
	TimeoutMs            uint64 `properties:"brian.registry.TimeoutMs"`            //timeout for requesting  server, default value is 10000ms
	ListenInterval       uint64 `properties:"brian.registry.ListenInterval"`       //Deprecated
	BeatInterval         int64  `properties:"brian.registry.BeatInterval"`         //the time interval for sending beat to server,default value is 5000ms
	NamespaceId          string `properties:"brian.registry.NamespaceId"`          //the namespaceId of
	Endpoint             string `properties:"brian.registry.Endpoint"`             //the endpoint for get  server addresses
	RegionId             string `properties:"brian.registry.RegionId"`             //the regionId for kms
	AccessKey            string `properties:"brian.registry.AccessKey"`            //the AccessKey for kms
	SecretKey            string `properties:"brian.registry.SecretKey"`            //the SecretKey for kms
	OpenKMS              bool   `properties:"brian.registry.OpenKMS"`              //it's to open kms,default is false. https://help.aliyun.com/product/28933.html
	CacheDir             string `properties:"brian.registry.CacheDir"`             //the directory for persist  service info,default value is current path
	UpdateThreadNum      int    `properties:"brian.registry.UpdateThreadNum"`      //the number of gorutine for update  service info,default value is 20
	NotLoadCacheAtStart  bool   `properties:"brian.registry.NotLoadCacheAtStart"`  //not to load persistent  service info in CacheDir at start time
	UpdateCacheWhenEmpty bool   `properties:"brian.registry.UpdateCacheWhenEmpty"` //update cache when get empty service instance from server
	Username             string `properties:"brian.registry.Username"`             //the username for  auth
	Password             string `properties:"brian.registry.Password"`             //the password for  auth
	LogDir               string `properties:"brian.registry.log.dir"`              //the directory for log, default is current path
	RotateTime           string `properties:"brian.registry.rotate.time"`          //the rotate time for log, eg: 30m, 1h, 24h, default is 24h
	MaxAge               int64  `properties:"brian.registry.max-age"`              //the max age of a log file, default value is 3
	LogLevel             string `properties:"brian.registry.log.level"`            //the level of log, it's must be debug,info,warn,error, default value is info
	ClusterName          string `properties:"brian.registry.cluster.name"`         // 默认值DEFAULT
	GroupName            string `properties:"brian.registry.group.name"`
}

func DefaultConfig() *RegistryConfig {
	return &RegistryConfig{
		Type:                "nacos",
		IpAddr:              "127.0.0.1",
		ContextPath:         "/nacos",
		Port:                80, // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		/*	LogDir:              "/tmp/nacos/log",
			CacheDir:            "/tmp/nacos/cache",*/
		RotateTime: "1h",
		MaxAge:     3,
		LogLevel:   "debug",
	}
}

// RewConfig 读取注册中心的相关配置
func RewConfig() (*RegistryConfig, error) {
	config := DefaultConfig()
	err := conf.UnmarshalToStruct(config)
	if nil != err {
		logrus.Error("Unmarshal config ", xlogger.FieldMod(xconst.ModConfig), xlogger.FieldErrKind(xconst.ErrKindUnmarshalConfigErr), xlogger.FieldErr(err))
	}
	return config, err
}

// Registry register/deregister service
// registry impl should control rpc timeout
type Registry interface {
	RegisterService(context.Context, *server.ServiceInfo) error
	DeregisterService(context.Context, *server.ServiceInfo) error
	io.Closer
}

// Nop registry, used for local development/debugging
// 用于本地开发 不进行注册
type Nop struct{}

// RegisterService ...
func (n Nop) RegisterService(context.Context, *server.ServiceInfo) error { return nil }

// DeregisterService ...
func (n Nop) DeregisterService(context.Context, *server.ServiceInfo) error { return nil }

// Close ...
func (n Nop) Close() error { return nil }

// Configuration ...
type Configuration struct {
}

// Rule ...
type Rule struct {
	Target  string
	Pattern string
}
