package xhttp

/**
 * Copyright (C) @2020 hugo network Co. Ltd
 * 定义http服务的处理类
 * @author: hugo
 * @version: 1.0
 * @date: 2020/8/2
 * @time: 11:56
 * @description:
 */

//服务处理的controller接口
type Controller interface {
	//controller的注册方法实现往服务中注册函数
	Register(server *Server)
}
