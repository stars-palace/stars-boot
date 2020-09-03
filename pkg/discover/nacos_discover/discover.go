package nacos_discover

import (
	"context"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/stars-palace/stars-boot/pkg/client/xnacos_client"
	"github.com/stars-palace/stars-boot/pkg/discover"
	"github.com/stars-palace/stars-boot/pkg/server"
)

/**
 * Copyright (C) @2020 hugo network Co. Ltd
 * nacos 的服务发现
 * @author: hugo
 * @version: 1.0
 * @date: 2020/8/27
 * @time: 21:35
 * @description:
 */

// nacos的实体
type NacoseDiscover struct {
	naming_client.INamingClient
}

func CreateNacoseDiscover(client *xnacos_client.NacosClient) *NacoseDiscover {
	return &NacoseDiscover{client.GetNamingClient()}
}

// 服务发现
func (n NacoseDiscover) GetServerInstance(con context.Context, param *discover.ServerInstancesParam) ([]*server.ServiceInfo, error) {
	// SelectOneHealthyInstance将会按加权随机轮训的负载均衡策略返回一个健康的实例	// 实例必须满足的条件：health=true,enable=true and weight>0
	instances, err := n.SelectInstances(vo.SelectInstancesParam{
		ServiceName: param.ServiceName,
		GroupName:   param.GroupName, // 默认值DEFAULT_GROUP
		Clusters:    param.Clusters,  // 默认值DEFAULT
		HealthyOnly: true,
	})
	if nil != err {
		return nil, err
	}
	var sers []*server.ServiceInfo
	if len(instances) <= 0 {
		return sers, nil
	}
	//对获取到的对象进行转换
	for _, v := range instances {
		ser := &server.ServiceInfo{
			Name:        v.ServiceName,
			Scheme:      v.ServiceName,
			IP:          v.Ip,
			Port:        int(v.Port),
			Weight:      v.Weight,
			Enable:      v.Enable,
			Healthy:     v.Healthy,
			Ephemeral:   v.Ephemeral,
			Metadata:    v.Metadata,
			Region:      "",
			Zone:        "",
			GroupName:   "",
			ClusterName: "",
		}
		sers = append(sers, ser)
	}
	return sers, nil
}
