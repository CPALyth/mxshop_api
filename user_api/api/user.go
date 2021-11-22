package api

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/hashicorp/consul/api"
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
	// 从注册中心获取服务信息
	cfg := api.DefaultConfig()
	consulInfo := global.ServerConfig.ConsulInfo
	cfg.Address = fmt.Sprintf("%s:%d", consulInfo.Host, consulInfo.Port)

	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	data, err := client.Agent().ServicesWithFilter(fmt.Sprintf(`Service == "%s"`, global.ServerConfig.UserSrvInfo.Name))
	if err != nil {
		panic(err)
	}
	userSrvHost := ""
	userSrvPort := 0
	for _, val := range data {
		userSrvHost = val.Address
		userSrvPort = val.Port
		break
	}

	// 拨号连接用户grpc服务器
	userConn, err := grpc.Dial(fmt.Sprintf("%s:%d", userSrvHost, userSrvPort), grpc.WithInsecure())
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

	// 验证验证码
	pass := store.Verify(passwordLogin.CaptchaId, passwordLogin.Captcha, false)
	if !pass {
		c.JSON(http.StatusBadRequest, gin.H{
			"captcha": "验证码错误",
		})
		return
	}
	// 拨号连接用户grpc服务器
	ip := global.ServerConfig.UserSrvInfo.Host
	port := global.ServerConfig.UserSrvInfo.Port
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

func Register(ctx *gin.Context) {
	// 用户注册
	registerForm := forms.RegisterForm{}
	if err := ctx.ShouldBind(&registerForm); err != nil {
		HandleValidatorError(ctx, err)
		return
	}
	// 将验证码保存到redis
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host, global.ServerConfig.RedisInfo.Port),
	})
	value, err := rdb.Get(context.Background(), registerForm.Mobile).Result()
	if err == redis.Nil || value != registerForm.Code {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "验证码错误",
		})
		return
	}

	// 拨号连接用户grpc服务器
	ip := global.ServerConfig.UserSrvInfo.Host
	port := global.ServerConfig.UserSrvInfo.Port
	userConn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[Register] 连接用户服务失败",
			"msg", err.Error())
	}
	userSrvClient := proto.NewUserClient(userConn)
	// 调用grpc接口
	rsp, err := userSrvClient.CreateUser(context.Background(), &proto.CreateUserInfo{
		NickName: registerForm.Mobile,
		Password: registerForm.PassWord,
		Mobile:   registerForm.Mobile,
	})
	if err != nil {
		zap.S().Errorf("[Register] 调用grpc接口失败, %s", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	// 注册成功, 返回token
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
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "生成token失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":   "登录成功",
		"token": token,
	})
}
