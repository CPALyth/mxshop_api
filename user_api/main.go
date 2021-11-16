package main

import (
	"fmt"
	"go.uber.org/zap"
	"mxshop_api/user_api/initialize"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	Router := initialize.Routers()

	port := 8021
	zap.S().Infof("启动服务器, 端口:%d", port)

	if err := Router.Run(fmt.Sprintf(":%d", port)); err != nil {
		zap.S().Panic("启动失败:", err.Error())
	}
}
