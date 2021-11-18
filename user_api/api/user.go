package api

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mxshop_api/user_api/middlewares"
	"mxshop_api/user_api/models"
	"strings"

	"mxshop_api/user_api/forms"
	"mxshop_api/user_api/global"
	"mxshop_api/user_api/global/response"
	"mxshop_api/user_api/proto"
	"net/http"
	"strconv"
	"time"
)

func removeTopStruct(fileds map[string]string) map[string]string {
	rsp := map[string]string{}
	for field, err := range fileds {
		rsp[field[strings.Index(field, ".")+1:]] = err
	}
	return rsp
}

func HandleValidatorError(c *gin.Context, err error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"msg": err.Error(),
		})
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"error": removeTopStruct(errs.Translate(global.Trans)),
	})
	return
}

func HandleGrpcErrorToHttp(err error, c *gin.Context) {
	// 将grpc的code转为HTTP的状态码
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"msg": e.Message(),
				})
			case codes.Internal:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "内部错误",
				})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "参数错误",
				})
			case codes.Unavailable:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "用户服务不可用",
				})

			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "其它错误" + e.Message(),
				})
			}
		}
	}
}

func GetUserList(ctx *gin.Context) {
	ip := global.ServerConfig.UserSrvInfo.Host
	port := global.ServerConfig.UserSrvInfo.Port
	// 拨号连接用户grpc服务器
	userConn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[GetUserList] 连接用户服务失败",
			"msg", err.Error())
	}
	userSrvClient := proto.NewUserClient(userConn)

	// 获取查询参数
	pn := ctx.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	pSize := ctx.DefaultQuery("pSize", "10")
	pSizeInt, _ := strconv.Atoi(pSize)
	// 调用接口
	rsp, err := userSrvClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    uint32(pnInt),
		PSize: uint32(pSizeInt),
	})
	if err != nil {
		zap.S().Error("[GetUserList] 查询用户列表失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	result := make([]interface{}, 0)
	for _, value := range rsp.Data {
		user := response.UserResponse{
			Id:       value.Id,
			NickName: value.NickName,
			Birthday: response.JsonTime(time.Unix(int64(value.BirthDay), 0)),
			Gender:   value.Gender,
			Mobile:   value.Mobile,
		}
		result = append(result, user)
	}
	ctx.JSON(http.StatusOK, result)
}

func PassWordLogin(c *gin.Context) {
	fmt.Println("登录...")
	// 表单验证
	passwordLogin := forms.PassWordLoginForm{}
	if err := c.ShouldBind(&passwordLogin); err != nil {
		HandleValidatorError(c, err)
		return
	}
	ip := global.ServerConfig.UserSrvInfo.Host
	port := global.ServerConfig.UserSrvInfo.Port
	// 验证验证码
	pass := store.Verify(passwordLogin.CaptchaId, passwordLogin.Captcha, true)
	if !pass {
		c.JSON(http.StatusBadRequest, gin.H{
			"captcha": "验证码错误",
		})
		return
	}
	// 拨号连接用户grpc服务器
	userConn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[GetUserList] 连接用户服务失败",
			"msg", err.Error())
	}
	userSrvClient := proto.NewUserClient(userConn)
	// 登录的逻辑
	rsp, err := userSrvClient.GetUserByMobile(context.Background(), &proto.MobileRequest{
		Mobile: passwordLogin.Mobile,
	})
	if err != nil {
		e, ok := status.FromError(err)
		if ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusBadRequest, map[string]string{
					"mobile": "用户不存在",
				})
			default:
				c.JSON(http.StatusInternalServerError, map[string]string{
					"mobile": "登录失败",
				})
			}
			return
		}
	}
	// 检查密码是否正确
	passRsp, passErr := userSrvClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
		Password:          passwordLogin.PassWord,
		EncryptedPassword: rsp.PassWord,
	})
	if passErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "登录失败",
		})
		return
	}
	if passRsp.Success == false {
		c.JSON(http.StatusOK, gin.H{
			"msg": "密码错误",
		})
		return
	}
	// 登录成功, 返回token
	now_unix := time.Now().Unix()
	claims := models.CustomClaims{
		ID:          uint(rsp.Id),
		NickName:    rsp.NickName,
		AuthorityId: uint(rsp.Role),
		StandardClaims: jwt.StandardClaims{
			NotBefore: now_unix,               // 签名的生效时间
			ExpiresAt: now_unix + 60*60*24*30, // 30天过期
			Issuer:    "ws",
		},
	}
	j := middlewares.NewJWT()
	token, err := j.CreateToken(claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "生成token失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   "登录成功",
		"token": token,
	})
}
