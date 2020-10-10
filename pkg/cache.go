package pkg

import (
	"time"
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"stree-index/pkg/config"
	"github.com/spf13/viper"
)

func RegisterNodeToCache(cnf *config.RedisKeys, logger log.Logger) {
	rdb := GetRedis()
	//ctx := context.Background()
	key := cnf.IndexNodeKeypPrefix + LocalIp
	err := rdb.Set(key, LocalIp, RedisIndexNodeExpire).Err()
	if err != nil {
		level.Error(logger).Log("msg", "Set node key error", "error", err)
	}
	err = rdb.HSet(cnf.IndexNodeMapName, key, LocalIp+viper.GetString("http_listen_addr")).Err()

	if err != nil {
		level.Error(logger).Log("msg", "Set node map key error", "error", err)
	}
	level.Info(logger).Log("msg", "RegisterNodeToCache Successfully", "key", key)

}

func RegisterHeartBeatToCache(ctx context.Context, cnf *config.RedisKeys, logger log.Logger) error {
	ticker := time.NewTicker(AliveRegisterInterval)
	level.Info(logger).Log("msg", "RegisterHeartBeatToCache start....")

	RegisterNodeToCache(cnf, logger)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			level.Info(logger).Log("msg", "receive_quit_signal_and_quit")
			return nil
		case <-ticker.C:
			RegisterNodeToCache(cnf, logger)
		}

	}
	return nil

}
