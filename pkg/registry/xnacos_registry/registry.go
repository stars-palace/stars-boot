package xnacos_registry

import (
	"context"
	"github.com/brian-god/brian-go/pkg/client/xnacos_client"
	"github.com/brian-god/brian-go/pkg/registry"
	"github.com/brian-god/brian-go/pkg/server"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

/**
 * nacos注册中心
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/21 9:25 上午
 * @updateDate           2020/8/21 9:25 上午
 * @version              1.0
**/
// nacos的实体
type NacosRegistery struct {
	naming_client.INamingClient
}

func CreateNacosRegister(client *xnacos_client.NacosClient) *NacosRegistery {
	return &NacosRegistery{client.GetNamingClient()}
}

//DefaultClientConfig创建一个默认的sever配置
func NacosServerConfigs(config *registry.RegistryConfig) []constant.ServerConfig {
	// 至少一个ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      config.IpAddr,
			ContextPath: config.ContextPath,
			Port:        config.Port,
		},
	}
	return serverConfigs
}

// RegisterService ... 服务注册
func (e *NacosRegistery) RegisterService(ctx context.Context, info *server.ServiceInfo) error {
	ok, err := e.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          info.IP,
		Port:        uint64(info.Port),
		ServiceName: info.Name,
		Weight:      info.Weight,
		Enable:      info.Enable,
		Healthy:     info.Healthy,
		Ephemeral:   info.Ephemeral,
		Metadata:    info.Metadata,
		ClusterName: info.ClusterName, // 默认值DEFAULT
		GroupName:   info.GroupName,   // 默认值DEFAULT_GROUP
	})
	if !ok {
		return err
	}
	return nil
}

// DeregisterService ... 注销服务
func (e *NacosRegistery) DeregisterService(ctx context.Context, info *server.ServiceInfo) error {
	ok, err := e.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          info.IP,
		Port:        uint64(info.Port),
		ServiceName: info.Name,
		Ephemeral:   info.Ephemeral,
		Cluster:     info.ClusterName, // 默认值DEFAULT
		GroupName:   info.GroupName,   // 默认值DEFAULT_GROUP
	})
	if !ok {
		return err
	}
	return nil
}

// Close ...
func (e *NacosRegistery) Close() error {
	return nil
}
