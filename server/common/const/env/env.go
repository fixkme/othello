package env

import (
	"os"
	"strconv"
)

const (
	APP_RpcListenAddr = "RPC_LISTEN_ADDR"
	APP_RpcAddr       = "RPC_ADDR"
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
