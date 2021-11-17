package initialize

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"mxshop_api/user_api/global"
	"os"
)

func InitConfig() {
	configFileName := "config-online.yaml"
	fmt.Println(os.Getenv("MXSHOP_DEBUG"))
	if os.Getenv("MXSHOP_DEBUG") == "true" {
		configFileName = "config-dev.yaml"
	}

	v := viper.New()
	v.SetConfigFile(configFileName)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(global.ServerConfig); err != nil {
		panic(err)
	}
	fmt.Println(global.ServerConfig)

	// viper动态监控配置文件变化
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		_ = v.ReadInConfig()
		_ = v.Unmarshal(global.ServerConfig)
		fmt.Println(global.ServerConfig)
	})
}
