package initialize

import (
	"github.com/gin-gonic/gin"
	"mxshop_api/user_api/middlewares"
	router2 "mxshop_api/user_api/router"
)

func Routers() *gin.Engine {
	Router := gin.Default()

	// 配置跨域
	Router.Use(middlewares.Cors())

	ApiGroup := Router.Group("/u/v1")
	router2.InitUserRouter(ApiGroup)
	return Router
}
