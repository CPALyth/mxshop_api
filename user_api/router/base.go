package router

import (
	"github.com/gin-gonic/gin"
	"mxshop_api/user_api/api"
)

func InitBaseRouter(Router *gin.RouterGroup) {
	BaseRouter := Router.Group("base")
	{
		BaseRouter.GET("captcha", api.GetCaptcha)
	}
}
