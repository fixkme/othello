package env

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/fixkme/gokit/framework/config"
	"github.com/fixkme/gokit/mlog"
)

const (
	APP_ConfigPath     = "CONFIG_PATH"
	APP_ServerId       = "SERVER_ID"
	APP_TimezoneOffset = "TIMEZONE_OFFSET"
	APP_LogPath        = "LOG_PATH"
	APP_LogLevel       = "LOG_LEVEL"
	APP_GateListenAddr = "GATE_LISTEN_ADDR" //gate监听地址
	APP_RpcListenAddr  = "RPC_LISTEN_ADDR"  //用于服务端监听的地址
	APP_RpcAddr        = "RPC_ADDR"         //用于rpc客户端连接服务的地址
	APP_EtcdEndpoints  = "ETCD_ENDPOINTS"   //etcd服务地址，用','隔开
	APP_MongoUri       = "MONGO_URI"
	APP_RedisAddr      = "REDIS_Addr"
)

func GetEnvInt(key string) int {
	strVal := os.Getenv(key)
	if strVal == "" {
		return 0
	}
	val, err := strconv.Atoi(strVal)
	if err != nil {
		fmt.Println("GetEnvInt Atoi:", err)
		return 0
	}
	return val
}

func GetEnvStr(key string) string {
	return os.Getenv(key)
}

func LoadConfigFromEnv(conf *config.AppConfig) error {
	replaceInt(&conf.ServerId, APP_ServerId)
	replaceInt(&conf.TimezoneOffset, APP_TimezoneOffset)
	replaceString(&conf.LogPath, APP_LogPath)
	replaceInt(&conf.LogLevel, APP_LogLevel)
	replaceString(&conf.RpcAddr, APP_RpcAddr)
	replaceString(&conf.RpcListenAddr, APP_RpcListenAddr)
	replaceString(&conf.EtcdEndpoints, APP_EtcdEndpoints)
	replaceString(&conf.MongoUri, APP_MongoUri)
	replaceString(&conf.RedisAddr, APP_RedisAddr)
	// default
	setDefault(&conf.TimezoneOffset, 28800)
	setDefault(&conf.LogPath, "logs")
	setDefault(&conf.LogLevel, int(mlog.DebugLevel))
	setDefault(&conf.LogStdOut, true)
	setDefault(&conf.EtcdEndpoints, "127.0.0.1:2379")
	setDefault(&conf.RpcGroup, "gbs")
	setDefault(&conf.RpcPollerNum, max(runtime.NumCPU()/2, 1))
	setDefault(&conf.RpcReadTimeout, 5*1e3)
	setDefault(&conf.RpcWriteTimeout, 5*1e3)
	return nil
}

func replaceString(s *string, name string) {
	if v := GetEnvStr(name); v != "" {
		*s = v
	}
}

func replaceInt(n *int, name string) {
	if v := GetEnvInt(name); v != 0 {
		*n = v
	}
}

type baseType interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string | bool
}

func setDefault[T baseType](v *T, d T) {
	var zero T
	if *v == zero {
		*v = d
	}
}
