package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mxshop_api/user_api/models"
	"net/http"
)

func IsAdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, _ := ctx.Get("claims")
		zap.S().Infof("访问用户: %d", claims.(*models.CustomClaims).ID)
		curUser := claims.(*models.CustomClaims)
		if curUser.AuthorityId == 2 { // 2是普通用户, 1 是管理员
			ctx.JSON(http.StatusForbidden, gin.H{
				"msg": "无权限",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
