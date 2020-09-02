package xgrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stars-palace/statrs-common/pkg/xcodec"
	"google.golang.org/grpc"
	"net/http"
	"reflect"
	"strings"
)

/**
 * grpc handler
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/17 9:19 上午
 * @updateDate           2020/8/17 9:19 上午
 * @version              1.0
**/

//处理的handler
func _Index_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HugoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	//地址定义
	//获取类型
	elem := reflect.TypeOf(srv).Elem()
	stype := fmt.Sprintf("/%s/%s/Grpc/%s", elem.PkgPath(), elem.Name(), in.MethodName)
	//拦截器处理
	if interceptor == nil {
		//return srv.(TestGrpcInterface).Index(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: stype,
	}

	//该处确定调用的具体方法
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		var (
			//rer error
			rValues2 []reflect.Value
		)
		request := req.(*HugoRequest)
		if err := ChickRequest(request); err != nil {
			return &HugoResponse{
				RequestId: request.RequestId,
				Code:      http.StatusInternalServerError,
				Message:   err.Error(),
			}, nil
		}
		//获取name
		methodName := request.MethodName
		//获取方法
		// TODO 方法有可能为空，后续处理
		method := reflect.ValueOf(srv).MethodByName(methodName)
		//服务端没有需要调用的方法
		if !method.IsValid() {
			return &HugoResponse{
				RequestId: request.RequestId,
				Code:      http.StatusInternalServerError,
				Message:   http.StatusText(http.StatusInternalServerError) + "-->" + reflect.ValueOf(srv).Elem().Kind().String() + " don't have " + methodName + "  method",
			}, nil
		}
		indx := method.Type().NumIn()
		//fmt.Printf("方法参数1的类型：%d\n", method.Type().In(0).Name())
		//fmt.Printf("方法参数2的类型：%d\n", method.Type().In(1).Name())
		//fmt.Printf("方法参数的个数：%d\n", indx)
		var data []interface{}
		if err := json.Unmarshal([]byte(request.Parameters[0]), &data); err != nil {
			// TODO 解析参数可能报错后续处理
			//将返回值转换成对应的类型
			return &HugoResponse{
				RequestId: request.RequestId,
				Code:      http.StatusInternalServerError,
				Message:   http.StatusText(http.StatusInternalServerError) + "-->" + err.Error(),
			}, nil
		}
		//TODO 正确写法
		for i := 0; i < indx; i++ {
			//for i,v:= range data{
			//获取客户端传来的值
			v := data[i]
			//fmt.Printf("接收到的类型：%d\n", reflect.ValueOf(v).String())
			rev, e := xcodec.UnmarshalByType(v, method.Type().In(i))
			if e != nil {
				return &HugoResponse{
					RequestId: request.RequestId,
					Code:      http.StatusInternalServerError,
					Message:   http.StatusText(http.StatusInternalServerError) + "-->" + e.Error(),
				}, nil
			}
			rValues2 = append(rValues2, rev)
		}
		//调用方法
		result := method.Call(rValues2)
		//定义一个返回数据
		resData := make([]interface{}, 0)
		//记录返回值类型
		resParType := strings.Builder{}
		//判断是否有返回
		if len(result) > 0 {
			for i, v := range result {
				data := v.Interface()
				if i == (len(result) - 1) {
					resParType.WriteString(v.Type().Name())
				} else {
					resParType.WriteString(v.Type().Name() + ",")
				}
				if nil != data {
					//判断返回值是不是error
					if v.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
						//强制转换成error
						data = data.(error).Error()
						//fmt.Printf("返回值第%d是error类型\n",i)
					}
				}
				//fmt.Printf("返回值类型:%s \n",v.Type().Name())
				resData = append(resData, data)
			}
		}
		//将数据类型追加到最后
		resData = append(resData, resParType.String())
		bt, er := json.Marshal(resData)
		if nil != er {
			//将返回值转换成对应的类型
			return &HugoResponse{
				RequestId: request.RequestId,
				Code:      http.StatusInternalServerError,
				Message:   http.StatusText(http.StatusInternalServerError) + "-->" + er.Error(),
				Data:      string(bt),
			}, nil
		}
		//将返回值转换成对应的类型
		return &HugoResponse{
			RequestId: request.RequestId,
			Code:      http.StatusOK,
			Message:   http.StatusText(http.StatusOK),
			Data:      string(bt),
		}, nil
	}
	return interceptor(ctx, in, info, handler)
}

// ChickRequest 校验请求信息
func ChickRequest(req *HugoRequest) error {
	if len(req.MethodName) == 0 {
		return errors.New(xcodec.ErrREQNotMethod)
	}
	return nil
}

// ChickResponse 检查返回信息
func ChickResponse(out *HugoResponse) error {
	//判断是否调用成功
	if out.Code != http.StatusOK {
		//如果有错判断返回消息是否有
		if !(len(out.Message) == 0) {
			return errors.New(out.Message)
		} else {
			return errors.New(xcodec.ErrServerException)
		}
	}
	return nil
}
