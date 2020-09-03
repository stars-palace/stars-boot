package stars_boot

import (
	"context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/labstack/gommon/color"
	"github.com/sirupsen/logrus"
	"github.com/stars-palace/stars-boot/pkg/client/xgrpc_client"
	"github.com/stars-palace/stars-boot/pkg/client/xnacos_client"
	"github.com/stars-palace/stars-boot/pkg/discover"
	"github.com/stars-palace/stars-boot/pkg/discover/nacos_discover"
	"github.com/stars-palace/stars-boot/pkg/registry"
	"github.com/stars-palace/stars-boot/pkg/registry/xnacos_registry"
	"github.com/stars-palace/stars-boot/pkg/server"
	"github.com/stars-palace/stars-boot/pkg/server/xgrpc"
	"github.com/stars-palace/stars-boot/pkg/server/xhttp"
	"github.com/stars-palace/stars-boot/pkg/worker"
	"github.com/stars-palace/stars-boot/pkg/worker/task"
	"github.com/stars-palace/stars-boot/pkg/xconst"
	"github.com/stars-palace/statrs-common/pkg/group"
	"github.com/stars-palace/statrs-common/pkg/utils/xgo"
	"github.com/stars-palace/statrs-common/pkg/xcast"
	"github.com/stars-palace/statrs-common/pkg/xfile"
	"github.com/stars-palace/statrs-common/pkg/xflag"
	"github.com/stars-palace/statrs-common/pkg/xlogger"
	conf "github.com/stars-palace/statrs-config"
	fileDatasource "github.com/stars-palace/statrs-config/datasource/file"
	"github.com/stars-palace/statrs-config/properties"
	"github.com/stars-palace/statrs-config/xini"
	"github.com/stars-palace/statrs-config/xyml"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/9/3 12:49 下午
 * @updateDate           2020/9/3 12:49 下午
 * @version              1.0
**/

// Application is the framework's instance, it contains the servers, workers, client and configuration settings.
// Create an instance of Application, by using &Application{}
type Application struct {
	servers     []server.Server
	workers     []worker.Worker
	logger      *logrus.Logger
	stopOnce    sync.Once
	initOnce    sync.Once
	startupOnce sync.Once

	//registerer registry.Registry

	signalHooker func(*Application)
	defers       []func() error

	governor *http.Server
	colorer  *color.Color

	httpServer           *xhttp.Server
	rpcServer            *xgrpc.Server
	registry             registry.Registry
	discover             discover.Discover
	registryConfig       *registry.RegistryConfig
	Name                 string `properties:"brian.application.name"`                  //应用名称
	LogLevel             string `properties:"brian.application.log.level"`             // 日志级别
	EnableRpcServer      bool   `properties:"brian.application.enable.RpcServer"`      //是否开启rpc服务
	EnableRegistryCenter bool   `properties:"brian.application.enable.RegistryCenter"` //是否启用注册中心
	EnableServerClient   bool   `properties:"brian.application.enable.ServerClient"`   //是否开启客户端链接
	RefreshTime          int    `properties:"brian.application.servers.refresh.time"`  //本地服务列表刷新时间
}

// 初始化应用
func (app *Application) initialize() {
	app.initOnce.Do(func() {
		app.servers = make([]server.Server, 0)
		app.workers = make([]worker.Worker, 0)
		app.signalHooker = hookSignals
		app.defers = []func() error{}
	})
}

// 获取默认的应用
func DefaultApplication() *Application {
	//开始使用默认的
	app := &Application{colorer: color.New(), logger: logrus.New(), Name: "brian", LogLevel: "info", EnableRpcServer: false, EnableRegistryCenter: false, RefreshTime: 15}
	//打印logo
	app.printBanner()
	//调用应用初始化的方法
	app.initialize()
	//默认启动方式不进行配置加载
	//app.loadConfig()
	//创建http服务
	app.defaultServerHTTP()
	//判断是否需要开启rpc服务
	if app.EnableRpcServer {
		//创建rpc服务
		app.defaultServeGRPC()
	}
	return app
}

// 创建一个读取配置文件的application
func RewConfigApplication() *Application {
	//开始使用默认的
	app := &Application{colorer: color.New(), logger: logrus.New(), Name: "brian", LogLevel: "info", EnableRpcServer: false, EnableRegistryCenter: false, RefreshTime: 15}
	//打印logo
	app.printBanner()
	//调用应用初始化的方法
	app.initialize()
	//配置加载
	app.loadConfig()
	//读取配置文件中对应用的配置
	if err := conf.UnmarshalToStruct(app); err != nil {
		logrus.Panic("read app config  error", xlogger.FieldMod(xconst.ModApp), xlogger.FieldErrKind(xconst.ReadAppConfigErr), xlogger.FieldErr(err))
	}
	//创建读取配置http服务
	app.serverHTTP()
	//判断是否需要开启rpc服务
	if app.EnableRpcServer {
		//创建读取配置rpc服务
		app.serveGRPC()
	}
	//设置应用的日志级别
	if level, err := logrus.ParseLevel(app.LogLevel); nil == err {
		app.logger.Level = level
	}
	//判断应用是否开启注册中心
	if app.EnableRegistryCenter {
		//初始化配置中心
		app.registryCenter()
		//进行服务的注册
		//app.registryServer()
	}
	//判断是否启用客户端链接
	if app.EnableServerClient {
		//初始化客户端链接
		app.initRpcClient()
		//刷新本地服务列表
		app.refreshServers()
	}
	return app
}

//默认配置启动rpc服务
func (app *Application) defaultServeGRPC() error {
	//获取一个grpc服务
	rpcServer := xgrpc.DefaultConfig().Build()
	app.rpcServer = rpcServer
	return app.Serve(rpcServer)
}

//rpc服务
func (app *Application) serveGRPC() error {
	//获取一个grpc服务
	rpcServer := xgrpc.StdConfig().Build()
	app.rpcServer = rpcServer
	return app.Serve(rpcServer)
}

// RegisterRpcServer 注册rpc服务
func (app *Application) RegisterRpcServer(in interface{}, srv interface{}) {
	if app.rpcServer != nil {
		app.rpcServer.Register(in, srv)
	} else {
		app.logger.Panic("RegisterRpcServer err app rpcServer is nil")
	}
}

//RegisterController 注册controller
func (app *Application) RegisterController(con xhttp.Controller) {
	app.httpServer.UseController(con)
}

//使用默认配置启动http服务
func (app *Application) defaultServerHTTP() error {
	httpServer := xhttp.DefaultConfig().Build()
	app.httpServer = httpServer
	return app.Serve(httpServer)
}

//http服务
func (app *Application) serverHTTP() error {
	httpServer := xhttp.StdConfig("http").Build()
	app.httpServer = httpServer
	return app.Serve(httpServer)
}

// 启动应用内部方法
func (app *Application) startup() (err error) {
	//执行注入的函数
	app.startupOnce.Do(func() {
		err = xgo.SerialUntilError(
			//放入执行的函数
			app.initLogger,
		)()
	})
	return
}

//初始化客户端链接
func (app *Application) initRpcClient() {
	xgrpc_client.InitServerGrpcClient(app.discover, app.registryConfig)
}

//创建注册中心
func (app *Application) registryCenter() {
	registryConfig, err := registry.RewConfig()
	//是否的能够获取到配置文件
	if nil != err {
		app.logger.Panic(xlogger.FieldMod(xconst.ModRegistry), xlogger.FieldErrKind(xconst.ReadRegistryConfigErr), xlogger.FieldErr(err))
	}
	//将注册中心的配置信息放入应用中
	app.registryConfig = registryConfig
	//获取配置中心的类型
	if xconst.Nacos == registryConfig.Type {
		//nacos
		nacosConfig := xnacos_client.NewNacosClientConfig(registryConfig)
		nacosServerConfig := xnacos_registry.NacosServerConfigs(registryConfig)
		//创建一个nacos的client
		nacosClient, err1 := xnacos_client.NewNacosClient(nacosConfig, nacosServerConfig)
		if nil != err1 {
			app.logger.Panic("create nacos client error ", xlogger.FieldMod(xconst.ModConfig), xlogger.FieldErr(err))
		}
		//获取注册中心
		app.registry = xnacos_registry.CreateNacosRegister(nacosClient)
		//获取服务发现的客户端
		app.discover = nacos_discover.CreateNacoseDiscover(nacosClient)
	}
}

func (app *Application) initLogger() error {
	logrus.SetOutput(os.Stdout)
	//日志级别
	if v := conf.Get(xconst.ApplicationLoglevel); v != nil {
		if v, err := xcast.ToStringE(v); nil == err {
			if level, err := logrus.ParseLevel(v); nil == err {
				logrus.SetLevel(level)
			}
		}
	}
	logrus.Debug("debug 日志")
	return nil
}

//提供外部启动应用执行
func (app *Application) Startup(fns ...func() error) error {
	if err := app.startup(); err != nil {
		return err
	}
	return xgo.SerialUntilError(fns...)()
}

// GracefulStop 完成必要的清理后停止应用程序
func (app *Application) GracefulStop(ctx context.Context) (err error) {
	app.beforeStop()
	app.stopOnce.Do(func() {
		//清理注册中心
		err = app.registry.Close()
		if err != nil {
			app.logger.Errorf("graceful stop register close err", xlogger.FieldMod(xconst.ModApp), xlogger.FieldErr(err))
		}
		var eg errgroup.Group
		for _, s := range app.servers {
			s := s
			eg.Go(func() error {
				return s.GracefulStop(ctx)
			})
		}
		err = eg.Wait()
	})
	return err
}

// Stop 完成必要的清理后立即停止程序
func (app *Application) Stop() (err error) {
	app.beforeStop()
	app.stopOnce.Do(func() {
		//清理注册中心
		/*err = app.registerer.Close()*/
		var eg errgroup.Group
		for _, s := range app.servers {
			s := s
			eg.Go(s.Stop)
		}
		for _, w := range app.workers {
			w := w
			eg.Go(w.Stop)
		}
		err = eg.Wait()
	})
	return
}

// Run run application
func (app *Application) Run() error {
	defer app.clean()
	if app.signalHooker == nil {
		app.signalHooker = hookSignals
	}
	//注册
	if app.registry == nil {
		app.registry = registry.Nop{}
	}

	app.signalHooker(app)

	// start govern
	var eg errgroup.Group
	//eg.Go(app.startGovernor)
	eg.Go(app.startServers)
	eg.Go(app.startWorkers)
	return eg.Wait()
}

//开启工作线程
func (app *Application) startWorkers() error {
	var eg group.Group
	// start multi workers
	for _, w := range app.workers {
		w := w
		eg.Go(func() error {
			return w.Run()
		})
	}
	return eg.Wait()
}

// 启动服务
func (app *Application) startServers() error {
	registryConfig := app.registryConfig
	var eg errgroup.Group
	xgo.ParallelWithErrorChan()
	// start multi servers
	for _, s := range app.servers {
		s := s
		eg.Go(func() (err error) {
			//开启注册中心进行服务注册
			if app.EnableRegistryCenter {
				serverInfo := s.Info(registryConfig.GroupName, registryConfig.ClusterName)
				//注册服务
				_ = app.registry.RegisterService(context.TODO(), serverInfo)
				app.logger.Info("start servers", xlogger.FieldMod(xconst.ModApp), xlogger.FieldAddr(serverInfo.Label()), xlogger.Any("scheme", serverInfo.Scheme))

			}
			//注销服务
			//defer app.registry.DeregisterService(context.TODO(), serverInfo)
			//defer app.registry.DeregisterService(context.TODO(), serverInfo)
			//defer app.logger.Info("exit server", logger.FieldMod(xconst.ModApp), logger.FieldErr(err), logger.FieldAddr(serverInfo.Label()))
			return s.Serve()
		})
	}
	return eg.Wait()
}

func (app *Application) clean() {
	for i := len(app.defers) - 1; i >= 0; i-- {
		fn := app.defers[i]
		if err := fn(); err != nil {
			//xlog.Error("clean.defer", xlog.String("func", xstring.FunctionName(fn)))
		}
	}
	//_ = xlog.DefaultLogger.Flush()
	//_ = xlog.JupiterLogger.Flush()
}
func (app *Application) beforeStop() {
	if app.EnableRegistryCenter {
		app.logger.Info("停止服务并注销注册中心的服务")
		//注销服务
		app.deregisterService()
	}
	// 应用停止之前的处理
	//app.logger.Info("leaving jupiter, bye....", xlog.FieldMod(ecode.ModApp))
}

//刷新本地服务列表
func (app *Application) refreshServers() {
	backTask := task.BackgroundTask{}
	backTask.Time1 = time.Duration(app.RefreshTime) * time.Second
	backTask.AddJob(func() error {
		xgrpc_client.RangeConns(func(key, value interface{}) bool {
			//将map的key转为string
			name := key.(string)
			//通过遍历查询服务列表
			servers, err := app.discover.GetServerInstance(context.Background(), &discover.ServerInstancesParam{
				ServiceName: name,
				GroupName:   app.registryConfig.GroupName,
				Clusters:    nil,
			})
			if err != nil {
				app.logger.Error("  refresh local server list  ", xlogger.FieldMod(xconst.ModWork), fmt.Sprintf("server name %s", key), xlogger.Error(err))
				return false
			}
			_, err1 := xgrpc_client.ChangeServerCons(name, servers)
			if err1 != nil {
				app.logger.Error("  refresh local server list  ", xlogger.FieldMod(xconst.ModWork), fmt.Sprintf("server name %s", key), xlogger.Error(err1))
				return false
			}
			app.logger.Info("  refresh local server list  ", xlogger.FieldMod(xconst.ModWork), fmt.Sprintf("server name %s", key))
			return true
		})
		return nil
	})
	app.Work(&backTask)
}

//deregisterService 进行服务注册
func (app *Application) deregisterService() {
	registryConfig := app.registryConfig
	for _, s := range app.servers {
		//获取服务信息
		serverInfo := s.Info(registryConfig.GroupName, registryConfig.ClusterName)
		//注销服务
		err := app.registry.DeregisterService(context.TODO(), serverInfo)
		app.logger.Info("exit server", xlogger.FieldMod(xconst.ModApp), xlogger.FieldErr(err), xlogger.FieldAddr(serverInfo.Label()))
	}
}

//注册服务
func (app *Application) Serve(s server.Server) error {
	app.servers = append(app.servers, s)
	return nil
}

//注册工作刘
func (app *Application) Work(w worker.Worker) error {
	app.workers = append(app.workers, w)
	return nil
}

//加载配置
func (app *Application) loadConfig() error {
	var (
		watchConfig = xflag.Bool("watch")
		configAddr  = xflag.String("config")
	)

	if configAddr == "" {
		app.logger.Warn("no config ... read default config")
		//为空则读取默认文件
		//优先级
		//botostrop.yml
		//application.yml
		//application.properties
		dir, _ := os.Getwd()
		ok, _ := xfile.PathExists(fmt.Sprintf("%s/resources/botostrop.yml", dir))
		if !ok {
			ok, _ = xfile.PathExists(fmt.Sprintf("%s/resources/application.yml", dir))
			if !ok {
				ok, _ = xfile.PathExists(fmt.Sprintf("%s/resources/application.properties", dir))
				if !ok {
					return nil
				} else {
					conf.SetConfigType("properties")
					configAddr = fmt.Sprintf("%s/resources/application.properties", dir)
				}
			} else {
				conf.SetConfigType("yml")
				configAddr = fmt.Sprintf("%s/resources/application.yml", dir)
			}
		} else {
			conf.SetConfigType("yml")
			configAddr = fmt.Sprintf("%s/botostrop.yml", dir)
		}
	}
	switch {
	case strings.HasPrefix(configAddr, "http://"),
		strings.HasPrefix(configAddr, "https://"):
		provider := fileDatasource.NewDataSource(configAddr, watchConfig)
		if err := conf.LoadFromDataSource(provider, toml.Unmarshal); err != nil {
			app.logger.Panic("load remote config ", xlogger.FieldMod(xconst.ModConfig), xlogger.FieldErrKind(xconst.ErrKindUnmarshalConfigErr), xlogger.FieldErr(err))
		}
		app.logger.Info("load remote config ", xlogger.FieldMod(xconst.ModConfig), xlogger.FieldAddr(configAddr))
	default:
		//provider := file_datasource.NewDataSource(configAddr, watchConfig)
		f, ferr := os.Open(configAddr)
		defer f.Close()
		if nil != ferr {
			app.logger.Panic(fmt.Sprintf("Error opening configuration file %s ", configAddr), xlogger.FieldMod(xconst.ModConfig), xlogger.FieldErrKind(xconst.ErrKindUnmarshalConfigErr), xlogger.Error(ferr))
		}
		fileSuffix := xfile.GetFileSuffix(f)
		confTyep, err2 := fileDatasource.ParseConfigType(fileSuffix[1:])
		if nil != err2 {
			app.logger.Panic(fmt.Sprintf("read config err  %s ", configAddr), xlogger.FieldMod(xconst.ModConfig), xlogger.FieldErrKind(xconst.ErrKindUnmarshalConfigErr), xlogger.Error(err2))
		}
		switch confTyep {
		case fileDatasource.Properties:
			if err := conf.LoadFromFile(f, properties.ReadFileConfig, properties.Unmarshal); err != nil {
				app.logger.Panic("load local file ", xlogger.FieldMod(xconst.ModConfig), xlogger.FieldErrKind(xconst.ErrKindUnmarshalConfigErr), xlogger.FieldErr(err))
			}
			break
		case fileDatasource.Yml:
			if err := conf.LoadFromFile(f, xyml.ReadFileConfig, xyml.Unmarshal); err != nil {
				app.logger.Panic("load local file ", xlogger.FieldMod(xconst.ModConfig), xlogger.FieldErrKind(xconst.ErrKindUnmarshalConfigErr), xlogger.FieldErr(err))
			}
			break
		case fileDatasource.InI:
			if err := conf.LoadFromFile(f, xini.ReadFileConfig, xini.Unmarshal); err != nil {
				app.logger.Panic("load local file ", xlogger.FieldMod(xconst.ModConfig), xlogger.FieldErrKind(xconst.ErrKindUnmarshalConfigErr), xlogger.FieldErr(err))
			}
			break
		default:
			app.logger.Panic(fmt.Sprintf("Cannot parse files of type %s ", fileSuffix), xlogger.FieldMod(xconst.ModConfig), xlogger.FieldErrKind(xconst.ErrKindUnmarshalConfigErr), xlogger.FieldErrKind(xconst.ErrConfigType))
		}
		app.logger.Info("load local file ", xlogger.FieldMod(xconst.ModConfig), xlogger.FieldAddr(configAddr))
	}
	return nil
}
func (app *Application) printBanner() error {
	const banner = `
	          _____                    _____                    _____                   _______
	         /\    \                  /\    \                  /\    \                 /::\    \
	        /::\____\                /::\____\                /::\    \               /::::\    \
	       /:::/    /               /:::/    /               /::::\    \             /::::::\    \
	      /:::/    /               /:::/    /               /::::::\    \           /::::::::\    \
	     /:::/    /               /:::/    /               /:::/\:::\    \         /:::/~~\:::\    \
	    /:::/____/               /:::/    /               /:::/  \:::\    \       /:::/    \:::\    \
	   /::::\    \              /:::/    /               /:::/    \:::\    \     /:::/    / \:::\    \
	  /::::::\    \   _____    /:::/    /      _____    /:::/    / \:::\    \   /:::/____/   \:::\____\
	 /:::/\:::\    \ /\    \  /:::/____/      /\    \  /:::/    /   \:::\ ___\ |:::|    |     |:::|    |
	/:::/  \:::\    /::\____\|:::|    /      /::\____\/:::/____/  ___\:::|    ||:::|____|     |:::|    |
	\::/    \:::\  /:::/    /|:::|____\     /:::/    /\:::\    \ /\  /:::|____| \:::\    \   /:::/    /
	 \/____/ \:::\/:::/    /  \:::\    \   /:::/    /  \:::\    /::\ \::/    /   \:::\    \ /:::/    /
	          \::::::/    /    \:::\    \ /:::/    /    \:::\   \:::\ \/____/     \:::\    /:::/    /
	           \::::/    /      \:::\    /:::/    /      \:::\   \:::\____\        \:::\__/:::/    /
	           /:::/    /        \:::\__/:::/    /        \:::\  /:::/    /         \::::::::/    /
	          /:::/    /          \::::::::/    /          \:::\/:::/    /           \::::::/    /
	         /:::/    /            \::::::/    /            \::::::/    /             \::::/    /
	        /:::/    /              \::::/    /              \::::/    /               \::/____/
	        \::/    /                \::/____/                \::/____/                 ~~
	         \/____/                  ~~

	 Welcome to hugo, starting application ...
	`
	/*const banner = `
	  o
	 <|>
	 / >
	 \o__ __o     o       o     o__ __o/    o__ __o
	  |     v\   <|>     <|>   /v     |    /v     v\
	 / \     <\  < >     < >  />     / \  />       <\
	 \o/     o/   |       |   \      \o/  \         /
	  |     <|    o       o    o      |    o       o
	 / \    / \   <\__ __/>    <\__  < >   <\__ __/>
									  |
							  o__     o
							  <\__ __/>

	Welcome to hugo, starting application ...
	`*/
	if app.colorer == nil {
		app.colorer = color.New()
	}
	app.colorer.Printf("%s\n", app.colorer.Blue(banner))
	return nil
}
