package xgrpc_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stars-palace/stars-boot/pkg/server/xgrpc"
	stars_registry_center "github.com/stars-palace/stars-registry-center"
	"github.com/stars-palace/stars-registry-center/discover"
	"github.com/stars-palace/stars-registry-center/registry"
	"github.com/stars-palace/statrs-common/pkg/xcodec"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"math/rand"
	"reflect"
	"strings"
	"sync"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * grpc client
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/17 9:51 上午
 * @updateDate           2020/8/17 9:51 上午
 * @version              1.0
**/

// 定义一个存放server链接的集合
var conn sync.Map
var serverClient *ServerGrpcClient

// GrpcClient grpc 客户端
type GrpcClient interface {
	//调用远程的服务
	Call(serverName string, int interface{}, ctx context.Context, method string, result interface{}, params ...interface{}) error
}

// ConnGrpcClient 需要传入构建好的链接进行服务的调用
type ConnGrpcClient struct {
	cc *grpc.ClientConn
}

// ServerGrpcClient 服务使用的客户端
type ServerGrpcClient struct {
	//服务发现
	discover discover.Discover
	registry *registry.RegistryConfig
}

//获取客户端链接

func BrianGrpcClient() (*ServerGrpcClient, error) {
	if nil == serverClient {
		return nil, errors.New("GrpcClient does not init or Application not enable ServerClient")
	}
	return serverClient, nil
}

// RangeConns 遍历本地所有服务
func RangeConns(f func(key, value interface{}) bool) {
	conn.Range(f)
}

// ChangeServerCons 改变服务的链接
func ChangeServerCons(serverName string, servers []*stars_registry_center.ServiceInfo) ([]*grpc.ClientConn, error) {
	var conns []*grpc.ClientConn
	//通过遍历实例获取创建链接
	for _, v := range servers {
		//获取一个grpc的链接
		conni, err := grpc.Dial(fmt.Sprintf("%s:%d", v.IP, v.Port), grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		conns = append(conns, conni)
	}
	//并将数据存储进集合中
	conn.Store(serverName, conns)
	return conns, nil
}

//创建客户端
func NewConnGrpcClient(cc *grpc.ClientConn) *ConnGrpcClient {
	return &ConnGrpcClient{cc}
}

//创建客户端
func InitServerGrpcClient(dis discover.Discover, config *registry.RegistryConfig) {
	serverClient = &ServerGrpcClient{discover: dis, registry: config}
}

//int 接口
//ctx上下文
//method 调用的方法
//result 返回值 该调用方式暂时只支持单个返回
//params 请求参数
func (c *ServerGrpcClient) Call(serverName string, int interface{}, ctx context.Context, method string, result interface{}, params ...interface{}) error {
	//判断是否有链接了
	conn, err := c.getConn(serverName)
	//获取链接是否出错
	if nil != err {
		return err
	}
	return invoke(conn, int, ctx, method, result, params...)
}

//getConn 获取一个链接
func (c *ServerGrpcClient) getConn(serverName string) (*grpc.ClientConn, error) {
	//判断是否有链接了
	serConns, ok := conn.Load(serverName)
	if !ok {
		//获取根据serverName 获取链接
		conns, err := c.getConnByServerName(serverName)
		if nil != err {
			return nil, err
		}
		//将新获取到的链接赋值给前原链接
		serConns = conns
	}
	//进行类型的转换
	serChConns := serConns.([]*grpc.ClientConn)
	//对链接进行筛选
	conni, err1 := c.selectConn(serverName, serChConns)
	if err1 != nil {
		//获取根据serverName 获取链接
		conns1, err := c.getConnByServerName(serverName)
		if nil != err {
			return nil, err
		}
		val := rand.Intn(len(conns1))
		//获取一个链接
		conn1 := conns1[val]
		//随机获取一个
		return conn1, nil
	}
	//随机获取一个
	return conni, nil
}

//进行链接的筛选
/**
// Idle indicates the ClientConn is idle.
	Idle State = iota  空闲链接
	// Connecting indicates the ClientConn is connecting.
	Connecting  链接中
	// Ready indicates the ClientConn is ready for work.
	Ready 忙的链接
	// TransientFailure indicates the ClientConn has seen a failure but expects to recover.
	TransientFailure 瞬间故障链接
	// Shutdown indicates the ClientConn has started shutting down.
	Shutdown 停止的链接

*/
func (c *ServerGrpcClient) selectConn(serverName string, conns []*grpc.ClientConn) (*grpc.ClientConn, error) {
	//默认使用随机的算法，获取一个随机的数
	val := rand.Intn(len(conns))
	//获取一个链接
	conni := conns[val]
	//获取服务状态
	status := conni.GetState()
	switch status {
	case connectivity.Idle:
		//空闲直接返回
		return conni, nil
	case connectivity.Connecting:
		//链接中判断是否是最后一个，如果是返回
		if len(conns) == 1 {
			return conni, nil
		} else {
			conns = append(conns[:val], conns[val+1:]...)
			return c.selectConn(serverName, conns)
		}
	case connectivity.Ready:
		//工作中，判断是否是最后一个是的话返回
		if len(conns) == 1 {
			return conni, nil
		} else {
			conns = append(conns[:val], conns[val+1:]...)
			return c.selectConn(serverName, conns)
		}
	case connectivity.TransientFailure:
		//瞬间故障，判断是否是最后一个是的话就直接返回
		if len(conns) == 1 {
			return conni, nil
		} else {
			conns = append(conns[:val], conns[val+1:]...)
			return c.selectConn(serverName, conns)
		}
	case connectivity.Shutdown:
		//停止，判断是否是最后一个，如果是则返回空并报错
		if len(conns) == 1 {
			return nil, errors.New("No link available")
		} else {
			conns = append(conns[:val], conns[val+1:]...)
			//将不可用的删除并更新本地列表
			conn.Store(serverName, conns)
			return c.selectConn(serverName, conns)
		}
	default:
		return nil, nil
	}
}

//int 接口
//ctx上下文
//method 调用的方法
//result 返回值 该调用方式暂时只支持单个返回
//params 请求参数
func (c *ConnGrpcClient) Call(serverName string, int interface{}, ctx context.Context, method string, result interface{}, params ...interface{}) error {
	return invoke(c.cc, int, ctx, method, result, params...)
}

// getConnByServerName根据服务名称获取链接，主要用于服务的发现
func (c *ServerGrpcClient) getConnByServerName(serverName string) ([]*grpc.ClientConn, error) {
	//获取所有的服务实例
	servers, err := c.discover.GetServerInstance(context.Background(), &discover.ServerInstancesParam{
		ServiceName: serverName,
		//Clusters: c.registry.ClusterName,
		GroupName: c.registry.GroupName,
	})
	if err != nil {
		return nil, err
	}
	//判断是否有获取到实例
	if len(servers) <= 0 {
		return nil, errors.New(fmt.Sprintf("Can not  ServerInstance by server name: %s ", serverName))
	}
	return ChangeServerCons(serverName, servers)
}

// Invoke 具体调用
func invoke(cc *grpc.ClientConn, int interface{}, ctx context.Context, method string, result interface{}, params ...interface{}) error {
	out := new(xgrpc.HugoResponse)
	par := make([]string, 0)
	//request data
	data, err := json.Marshal(params)
	if err != nil {
		return errors.New(err.Error())
	}
	par = append(par, string(data))
	request := &xgrpc.HugoRequest{MethodName: method, Parameters: par}
	//获取类型
	elem := reflect.TypeOf(int).Elem()
	//定义服务调用的地址
	stype := fmt.Sprintf("%s.%s.Grpc/HugoGrpc", elem.PkgPath(), elem.Name())
	err = cc.Invoke(ctx, stype, request, out)
	//如果调用报错直接抛出
	if err != nil {
		return err
	}
	if err := xgrpc.ChickResponse(out); err != nil {
		return err
	}
	//判断返回的是否有数据
	if len(out.Data) == 0 {
		return nil
	}
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("输出参数必须是指针类型")
	}
	//对返回数据进行解码操作
	var resData []interface{}
	if err := json.Unmarshal([]byte(out.Data), &resData); err != nil {
		return err
	}
	//返回值的长度
	lenth := len(resData)
	//定义一个 error
	var reSerr error
	//判断是否有返回
	if lenth > 0 {
		//获取最后一位是返回值类型
		resType := resData[lenth-1]
		//获取真正的返回数据
		resData = resData[:lenth-1]
		//将数据类型转成string
		strResType := resType.(string)
		//将string截取长切片
		strResTypes := strings.Split(strResType, ",")
		for i, v := range strResTypes {
			//返回的数据
			resDataValue := resData[i]
			if "error" == v {
				if nil != resDataValue {
					reSerr = errors.New(resDataValue.(string))
				}
			} else {
				resVale, rerr := xcodec.UnmarshalByType(resDataValue, rv.Elem().Type())
				if nil != rerr {
					return rerr
				}
				//给返回结果赋值
				rv.Elem().Set(resVale)
			}
		}
	}
	//如果结果为nil,这里返回null
	return reSerr
}
