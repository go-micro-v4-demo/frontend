package main

import (
	"fmt"
	"github.com/go-micro-v4-demo/frontend/handler"
	helloworldPb "github.com/go-micro-v4-demo/helloworld/proto"
	userPb "github.com/go-micro-v4-demo/user/proto"
	mgrpc "github.com/go-micro/plugins/v4/client/grpc"
	mhttp "github.com/go-micro/plugins/v4/server/http"
	"github.com/gorilla/mux"
	k8sHeadlessSvc "github.com/gsmini/k8s-headless-svc"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"net/http"
)

var (
	service = "frontend"
	version = "latest"
)

const K8sSvcName = "user-svc"

// https://juejin.cn/post/6877424913775329287/
// k8s 插件是是需要k8s 客户端的很冗余 需要client访问apiserver
// 要么仿造k8s 插件写一个 要么机遇mdns自定义封装一个
// 核心思想是去解析svc headless 然后解析返回的endpoint
// NewRegistry returns a new mdns registry.
const UserSvcName = "user-info"       //user模块在k8s service中的metadata.name的名字
const HelloWordSvcName = "helloworld" //user模块在k8s service中的metadata.name的名字
func main() {
	UserSvc := &k8sHeadlessSvc.Service{Namespace: "default", SvcName: UserSvcName, PodPort: 9090}
	HelloWordSvc := &k8sHeadlessSvc.Service{Namespace: "default", SvcName: HelloWordSvcName, PodPort: 9090}
	reg := k8sHeadlessSvc.NewRegistry([]*k8sHeadlessSvc.Service{UserSvc, HelloWordSvc})
	srv := micro.NewService(
		micro.Server(mhttp.NewServer()), //当前服务的类型 http 对外提供http
		micro.Client(mgrpc.NewClient())) //当前client的类型grpc 对内调用grpc
	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Address("0.0.0.0:8081"), //对外暴漏8000端口
		micro.Registry(reg),
	)
	client := srv.Client()
	svc := &handler.Frontend{
		UserService:       userPb.NewUserService(UserSvcName, client),
		HelloworldService: helloworldPb.NewHelloworldService("helloworld", client),
	}
	r := mux.NewRouter()
	r.HandleFunc("/index", svc.HomeHandler).Methods(http.MethodGet)
	//反爬虫
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "User-agent: *\nDisallow: /") })
	//健康检查
	r.HandleFunc("/_healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })

	var httpHandler http.Handler = r
	// Register handler
	if err := micro.RegisterHandler(srv.Server(), httpHandler); err != nil {
		logger.Fatal(err)
	}
	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
