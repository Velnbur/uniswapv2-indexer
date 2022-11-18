package config

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
)

func (c *config) Redis() *redis.Client {
	return c.redis.Do(func() interface{} {
		config := struct {
			Addr     string `fig:"address"`
			Password string `fig:"password"`
			DB       int    `fig:"db"`
		}{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		}

		err := figure.Out(&config).From(kv.MustGetStringMap(c.getter, "redis")).Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to get data redis from config"))
		}

		clientRedis := redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       config.DB,
		})

		if err := clientRedis.Ping(context.TODO()).Err(); err != nil {
			panic(errors.Wrap(err, "failed to connect to redis"))
		}

		return clientRedis
	}).(*redis.Client)
}
