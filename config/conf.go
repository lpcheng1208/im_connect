package config

import (
	"fmt"
	"os"
	"time"
)

var (
	Connect ConnectConf
)

// ConnectConf Connect配置
type ConnectConf struct {
	TCPListenAddr string
	WSListenAddr  string
	RedisIP       string
	RedisPassword string
	SubscribeNum  int
}

func init() {
	env := os.Getenv("gim_env")
	nowTime := time.Now().Format("2006-01-02 15:04:05")
	switch env {
	case "prod":
		fmt.Printf("%s case env: prod ---- >> %s \n", nowTime, env)
		initProdConf()
	case "test":
		fmt.Printf("%s case env: test ---- >> %s \n", nowTime, env)
		initTestConf()
	default:
		fmt.Printf("%s case env: default ---- >> %s \n", nowTime, env)
		initLocalConf()
	}
}
