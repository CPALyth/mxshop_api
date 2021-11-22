package initialize

import (
	"fmt"

	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"mxshop_api/user_api/global"
	"mxshop_api/user_api/proto"
)

func InitSrvConn() {
	consulInfo := global.ServerConfig.ConsulInfo
	userConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", consulInfo.Host, consulInfo.Port, global.ServerConfig.UserSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		zap.S().Fatal("InitSrvConn 连接用户服务失败")
	}
	global.UserSrvClient = proto.NewUserClient(userConn)
}

//func InitSrvConn2() {
//	// 从注册中心获取服务信息
//	cfg := api.DefaultConfig()
//	consulInfo := global.ServerConfig.ConsulInfo
//	cfg.Address = fmt.Sprintf("%s:%d", consulInfo.Host, consulInfo.Port)
//
//	client, err := api.NewClient(cfg)
//	if err != nil {
//		panic(err)
//	}
//
//	data, err := client.Agent().ServicesWithFilter(fmt.Sprintf(`Service == "%s"`, global.ServerConfig.UserSrvInfo.Name))
//	if err != nil {
//		panic(err)
//	}
//	userSrvHost := ""
//	userSrvPort := 0
//	for _, val := range data {
//		userSrvHost = val.Address
//		userSrvPort = val.Port
//		break
//	}
//	if userSrvHost == "" {
//		zap.S().Fatal("InitSrvConn 连接用户服务失败")
//		return
//	}
//
//	// 拨号连接用户grpc服务器
//	userConn, err := grpc.Dial(fmt.Sprintf("%s:%d", userSrvHost, userSrvPort), grpc.WithInsecure())
//	if err != nil {
//		zap.S().Errorw("[GetUserList] 连接用户服务失败",
//			"msg", err.Error())
//	}
//	global.UserSrvClient = proto.NewUserClient(userConn)
//
//}
