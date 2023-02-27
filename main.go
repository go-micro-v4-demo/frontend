package main

import (
	"fmt"
	"github.com/go-micro-v4-demo/frontend/handler"
	helloworldPb "github.com/go-micro-v4-demo/helloworld/proto"
	userPb "github.com/go-micro-v4-demo/user/proto"
	mgrpc "github.com/go-micro/plugins/v4/client/grpc"
	mhttp "github.com/go-micro/plugins/v4/server/http"
	regs "go-micro.dev/v4/registry"
	"go-micro.dev/v4/util/log"

	"github.com/gorilla/mux"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"net/http"
)

var (
	service = "frontend"
	version = "latest"
)

func main() {
	// Create service

	reg := regs.NewMemoryRegistry() //内存registry
	if err := reg.Register(&regs.Service{
		Name:    "user",
		Version: "latest",
		Nodes: []*regs.Node{
			{
				//k8s pod ip
				Address: "10.32.0.3:8080",
			},
			{ // k8s pod ip
				Address: "10.38.0.1:8080",
			},
		},
	}); err != nil {
		log.Fatalf("registry failed, err: %v", err)
	}
	srv := micro.NewService(
		micro.Server(mhttp.NewServer()),
		micro.Client(mgrpc.NewClient()))
	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Address("0.0.0.0:8080"), //对外暴漏8000端口
		micro.Registry(reg),
	)
	client := srv.Client()
	svc := &handler.Frontend{
		UserService:       userPb.NewUserService("user", client),
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
