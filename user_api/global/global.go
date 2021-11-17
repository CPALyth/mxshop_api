package global

import (
	ut "github.com/go-playground/universal-translator"
	"mxshop_api/user_api/config"
)

var (
	Trans        ut.Translator
	ServerConfig *config.ServerConfig = &config.ServerConfig{}
)
