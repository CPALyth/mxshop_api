package main

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"time"
)

func main() {
	// 服务端配置
	sc := []constant.ServerConfig{
		{
			IpAddr: "192.168.1.103",
			Port:   8848,
		},
	}
	// 客户端配置
	cc := constant.ClientConfig{
		NamespaceId:         "cc002bf5-cf66-4e9f-bda9-d56d74cce2f3", //we can create multiple clients with different namespaceId to support multiple namespace.When namespace is public, fill in the blank string here.
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "tmp/nacos/log",
		CacheDir:            "tmp/nacos/cache",
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
	}
	// 创建动态配置客户端
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})
	if err != nil {
		panic(err)
	}
	// 读取配置
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: "user_api.yaml",
		Group:  "dev",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(content)
	// 监听配置变化
	err = configClient.ListenConfig(vo.ConfigParam{
		DataId: "user_api.yaml",
		Group:  "dev",
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("配置文件产生变化")
			fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
		},
	})
	time.Sleep(3000 * time.Second)
}
