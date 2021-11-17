package main

import (
	"fmt"
	"go.uber.org/zap"
	"mxshop_api/user_api/global"
	"mxshop_api/user_api/initialize"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	// 初始化logger
	initialize.InitLogger()

	// 初始化配置文件
	initialize.InitConfig()

	// 初始化routers
	Router := initialize.Routers()

	// 初始化
	_ = initialize.InitTrans("zh")

	port := global.ServerConfig.Port
	zap.S().Infof("启动服务器, 端口:%d", port)

	if err := Router.Run(fmt.Sprintf(":%d", port)); err != nil {
		zap.S().Panic("启动失败:", err.Error())
	}
}
