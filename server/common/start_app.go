package common

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/fixkme/gokit/framework/config"
	"github.com/fixkme/gokit/framework/core"
	"github.com/fixkme/gokit/mlog"
	"github.com/fixkme/gokit/rpc"
	"github.com/fixkme/othello/server/common/env"
)

func StartApp(ctx context.Context, wg *sync.WaitGroup, name string, rpcHandler rpc.RpcHandler) {
	fmt.Printf("start %s server\n", name)
	// load config
	// configPath := "config"
	// if v := env.GetEnvStr(env.APP_ConfigPath); len(v) > 0 {
	// 	configPath = v
	// }
	// configFile := filepath.Join(configPath, fmt.Sprintf("%s.json", name))
	err := config.LoadConfig("", env.LoadConfigFromEnv)
	if err != nil {
		fmt.Println("LoadConfig error:", err)
		panic(err)
	}
	fmt.Println("----------------- app config -----------------")
	fmt.Println(config.Config.JsonFormat())
	// log
	logConf := &config.Config.LogConfig
	if err := mlog.UseDefaultLogger(ctx, wg, logConf.LogPath, name, mlog.Level(logConf.LogLevel), logConf.LogStdOut); err != nil {
		fmt.Println("UseDefaultLogger error:", err)
		panic(err)
	}
	// pprof
	if port := config.Config.PprofPort; port != 0 {
		runtime.SetBlockProfileRate(1)
		go func() {
			logger.Infof("pprof listen on: %d", port)
			addr := fmt.Sprintf(":%d", port)
			if err := http.ListenAndServe(addr, nil); err != nil {
				logger.Fatalf("pprof listen failed: %v", err)
			}
		}()
	}
	// redis ?
	redisConf := &config.Config.RedisConfig
	if redisConf.RedisAddr != "" {
		if err = core.InitRedis(redisConf); err != nil {
			logger.Fatalf("init redis failed: %v", err)
		}
	}
	// mongodb ?
	if uri := config.Config.MongoUri; uri != "" {
		if err = core.InitMongo(uri); err != nil {
			logger.Fatalf("init mongodb failed: %v", err)
		}
	}
	// rpc
	rpcConf := &config.Config.RpcConfig
	if err = core.InitRpcModule(name, rpcHandler, rpcConf); err != nil {
		logger.Fatalf("init rpc failed: %v", err)
	}
}
