package env

import (
	"os"
	"strconv"
)

const (
	APP_GateListenAddr = "GATE_LISTEN_ADDR" //gate监听地址
	APP_RpcListenAddr  = "RPC_LISTEN_ADDR"  //用于服务端监听的地址
	APP_RpcAddr        = "RPC_ADDR"         //用于rpc客户端连接服务的地址
	APP_EtcdEndpoints  = "ETC_ENDPOINTS"    //etcd服务地址，用','隔开
)

func GetEnvInt(key string) int {
	strVal := os.Getenv(key)
	if strVal == "" {
		return 0
	}
	val, err := strconv.Atoi(strVal)
	if err != nil {
		return 0
	}
	return val
}

func GetEnvStr(key string) string {
	return os.Getenv(key)
}
