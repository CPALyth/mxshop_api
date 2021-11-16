package router

import (
	"github.com/gin-gonic/gin"
	"mxshop_api/user_api/api"
)

func InitUserRouter(Router *gin.RouterGroup) {
	UserRouter := Router.Group("user")
	{
		UserRouter.GET("list", api.GetUserList)
	}
}
