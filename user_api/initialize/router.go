package initialize

import (
	"github.com/gin-gonic/gin"
	router2 "mxshop_api/user_api/router"
)

func Routers() *gin.Engine {
	Router := gin.Default()
	ApiGroup := Router.Group("/u/v1")
	router2.InitUserRouter(ApiGroup)
	return Router
}
