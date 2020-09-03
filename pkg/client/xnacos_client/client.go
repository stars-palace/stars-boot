package xnacos_client

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/sirupsen/logrus"
	"github.com/stars-palace/stars-boot/pkg/registry"
	"sync"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/27 3:33 下午
 * @updateDate           2020/8/27 3:33 下午
 * @version              1.0
**/
// nacos客户端
type NacosClient struct {
	//服务的client
	namingClient *naming_client.INamingClient
	//配置的client
	configClient *config_client.IConfigClient
	//客户端配置
	clientConf *constant.ClientConfig
	kvs        sync.Map
	//服务配置
	serConfigs []constant.ServerConfig
	//logger
	log *logrus.Logger
}

// 获取INamingClient
func (client *NacosClient) GetNamingClient() naming_client.INamingClient {
	return *client.namingClient
}

/**
TimeoutMs            uint64 //timeout for requesting Nacos server, default value is 10000ms
ListenInterval       uint64 //Deprecated
BeatInterval         int64  //the time interval for sending beat to server,default value is 5000ms
NamespaceId          string //the namespaceId of Nacos
Endpoint             string //the endpoint for get Nacos server addresses
RegionId             string //the regionId for kms
AccessKey            string //the AccessKey for kms
SecretKey            string //the SecretKey for kms
OpenKMS              bool   //it's to open kms,default is false. https://help.aliyun.com/product/28933.html
CacheDir             string //the directory for persist nacos service info,default value is current path
UpdateThreadNum      int    //the number of gorutine for update nacos service info,default value is 20
NotLoadCacheAtStart  bool   //not to load persistent nacos service info in CacheDir at start time
UpdateCacheWhenEmpty bool   //update cache when get empty service instance from server
Username             string //the username for nacos auth
Password             string //the password for nacos auth
LogDir               string //the directory for log, default is current path
RotateTime           string //the rotate time for log, eg: 30m, 1h, 24h, default is 24h
MaxAge               int64  //the max age of a log file, default value is 3
LogLevel

*/
//NewNacosClientConfig 创建一个默认的client配置
func NewNacosClientConfig(config *registry.RegistryConfig) constant.ClientConfig {
	clientConfig := constant.ClientConfig{
		TimeoutMs:            config.TimeoutMs, // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId
		BeatInterval:         config.BeatInterval,
		NamespaceId:          config.NamespaceId,
		Endpoint:             config.Endpoint,
		RegionId:             config.RegionId,
		AccessKey:            config.AccessKey,
		SecretKey:            config.SecretKey,
		OpenKMS:              config.OpenKMS, // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId
		CacheDir:             config.CacheDir,
		UpdateThreadNum:      config.UpdateThreadNum,
		NotLoadCacheAtStart:  config.NotLoadCacheAtStart,
		UpdateCacheWhenEmpty: config.UpdateCacheWhenEmpty,
		Username:             config.Username,
		Password:             config.Password,
		LogDir:               config.LogDir,
		RotateTime:           config.RotateTime, // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId
		MaxAge:               config.MaxAge,
		LogLevel:             config.LogLevel,
	}
	return clientConfig
}

//创建nacosclient
func NewNacosClient(client constant.ClientConfig, server []constant.ServerConfig) (*NacosClient, error) {
	// 创建服务发现客户端
	namingClient, err := clients.CreateNamingClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: server,
		constant.KEY_CLIENT_CONFIG:  client,
	})
	if nil != err {
		return nil, err
	}

	// 创建动态配置客户端
	configClient, err1 := clients.CreateConfigClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: server,
		constant.KEY_CLIENT_CONFIG:  client,
	})
	if nil != err1 {
		return nil, err
	}
	nacosClient := &NacosClient{
		namingClient: &namingClient,
		configClient: &configClient,
		clientConf:   &client,
		serConfigs:   server,
		log:          logrus.New(),
	}
	return nacosClient, nil
}
