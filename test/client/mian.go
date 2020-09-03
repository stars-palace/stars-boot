package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	stars_boot "github.com/stars-palace/stars-boot"
	"github.com/stars-palace/stars-boot/pkg/client/xgrpc_client"
	"github.com/stars-palace/stars-boot/pkg/server/xhttp"
	"github.com/stars-palace/stars-boot/pkg/worker/task"
	"github.com/stars-palace/stars-boot/test/api"
	"net/http"
	"time"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/31 5:10 下午
 * @updateDate           2020/8/31 5:10 下午
 * @version              1.0
**/

type ClientController struct {
}

func (test *ClientController) Register(server *xhttp.Server) {
	server.GET("/client", test.index)
}

//写入一个测试的方法
func (test *ClientController) index(ctx echo.Context) error {
	//获取客户端链接
	client, err := xgrpc_client.BrianGrpcClient()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	var res string
	//调用服务
	err1 := client.Call("brian-base-rpc", new(api.TestApi), context.Background(), api.SayHello, &res, "张三")
	if err1 != nil {
		return ctx.JSON(http.StatusInternalServerError, err1.Error())
	}
	return ctx.JSON(http.StatusOK, res)
}
func main() {
	//获取客户端链接
	runConfigApp()
	//time2()
}

func time2() {
	backTask := task.BackgroundTask{}
	backTask.Time1 = time.Duration(1) * time.Second
	backTask.AddJob(func() error {
		fmt.Println("后台任务执行")
		return nil
	})
	backTask.Run()
	time.Sleep(time.Second * 90)
}
func runConfigApp() {
	app := stars_boot.RewConfigApplication()
	app.RegisterController(&ClientController{})
	if err := app.Startup(); err != nil {
		fmt.Println("启动有误")
	}
	app.Run()
}
