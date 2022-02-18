package connector

import (
	"github.com/Rede-Legit/galax/pkg/util"
	"github.com/go-redis/redis/v8"
)

func NewRedis(url string) *redis.Client {
	options, err := redis.ParseURL(url)

	if err != nil {
		util.Log(err)
	}

	return redis.NewClient(options)

}
