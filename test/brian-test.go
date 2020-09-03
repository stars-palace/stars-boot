package main

import (
	"cloud.google.com/go/trace/testdata/helloworld"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/sirupsen/logrus"
	stars_boot "github.com/stars-palace/stars-boot"
	"github.com/stars-palace/stars-boot/pkg/client/xgrpc_client"
	"github.com/stars-palace/stars-boot/pkg/server/xgrpc"
	"github.com/stars-palace/stars-boot/pkg/server/xhttp"
	"github.com/stars-palace/stars-boot/test/api"
	"net/http"
	"time"
)

var ser = stars_boot.Application{}

//构建一个controller

type TestController struct {
}

func (test *TestController) Register(server *xhttp.Server) {
	server.GET("/index", test.index)
}

//写入一个测试的方法
func (test *TestController) index(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "hello hugo")
}

func main() {
	//dir, _ := os.Getwd()
	//out := conf.InitConfig(fmt.Sprintf("%s/resources/application.properties", dir))
	//fmt.Println(out)
	/*r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	go r.Run() // listen and serve on 0.0.0.0:8080
	server := grpc.NewServer()

	helloworld.RegisterGreeterServer(server, new(Greeter))

	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err.Error())
	}
	server.Serve(lis)*/

	/*if err := ser.Startup(serverHTTP, serveGRPC); err != nil {
		fmt.Println("启动有误")
	}
	ser.Run()*/
	//nacos()
	runConfigApp()
	//and()
}

var serverClient *xgrpc_client.ConnGrpcClient

func and() {
	ticker := time.NewTicker(time.Second * 2)
	for _ = range ticker.C {
		fmt.Printf("ticked at %v\n", time.Now())
	}
	/*for i:=0;i<100;i++{
		b := rand.Intn(100)  //生成0-99之间的随机数
		fmt.Println(b)
	}*/
}

func runConfigApp() {
	app := stars_boot.RewConfigApplication()
	//注册rpc服务
	app.RegisterRpcServer(new(api.TestApi), new(TestApiImpl))
	//注册http controller
	app.RegisterController(&TestController{})
	if err := app.Startup(); err != nil {
		fmt.Println("启动有误")
	}
	app.Run()
}
func runApp() {
	app := stars_boot.DefaultApplication()
	//注册rpc服务
	app.RegisterRpcServer(new(api.TestApi), new(TestApiImpl))
	//注册http controller
	app.RegisterController(&TestController{})
	if err := app.Startup(); err != nil {
		fmt.Println("启动有误")
	}
	app.Run()
}
func Hello() error {
	fmt.Printf("你好")
	return nil
}
func nacos() {
	clientConfig := constant.ClientConfig{
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		/*LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",*/
		RotateTime: "1h",
		MaxAge:     3,
		LogLevel:   "debug",
		//Username: "nacos",
		//Password: "nacos",
	}

	// 至少一个ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "localhost",
			ContextPath: "/nacos",
			Port:        8848,
		},
	}

	// 创建服务发现客户端
	namingClient, err := clients.CreateNamingClient(map[string]interface{}{
		constant.KEY_SERVER_CONFIGS: serverConfigs,
		constant.KEY_CLIENT_CONFIG:  clientConfig,
	})
	if nil != err {
		logrus.Panic(err.Error())
	}
	success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          "10.0.0.11",
		Port:        8848,
		ServiceName: "demo.go",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "shanghai"},
		ClusterName: "cluster-a", // 默认值DEFAULT
		GroupName:   "group-a",   // 默认值DEFAULT_GROUP
	})
	if !success {
		logrus.Panic(err.Error())
	}
	success1, err1 := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          "10.0.0.11",
		Port:        8848,
		ServiceName: "demo.go",
		Ephemeral:   true,
		Cluster:     "cluster-a", // 默认值DEFAULT
		GroupName:   "group-a",   // 默认值DEFAULT_GROUP
	})
	if !success1 {
		logrus.Panic(err1.Error())
	}
	fmt.Println(namingClient)
}

//rpc服务
func serveGRPC() error {
	//获取一个grpc服务
	grpcServer := xgrpc.DefaultConfig().Build()
	grpcServer.Register(new(api.TestApi), new(TestApiImpl))
	//注册服务
	return ser.Serve(grpcServer)
}
func serverHTTP() error {
	httpServer := xhttp.StdConfig("http").Build()
	//使用
	httpServer.UseController(&TestController{})
	//启动服务
	ser.Serve(httpServer)
	return nil
}

type Greeter struct {
	helloworld.GreeterServer
}

//grpc
func (g Greeter) SayHello(context context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{
		Message: "Hello Jupiter",
	}, nil
}

type TestApiImpl struct {
}

// SayHello
// TODO rpc接口的实现必须使用值接收者不能够使用指针接收者，使用指针接收者会造成结构体是否实现接口的判断出错
func (test TestApiImpl) SayHello(name string) string {
	return fmt.Sprintf("hell %s", name)
}
