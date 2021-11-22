package global

import (
	ut "github.com/go-playground/universal-translator"
	"mxshop_api/user_api/config"
	"mxshop_api/user_api/proto"
)

var (
	Trans         ut.Translator
	ServerConfig  = &config.ServerConfig{}
	UserSrvClient proto.UserClient
)
