package galax

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func NewRedis(url string) *redis.Client {
	options, err := redis.ParseURL(url)

	if err != nil {
		panic(err)
	}

	client := redis.NewClient(options)

	_, err = client.Ping(
		context.Background(),
	).Result()

	if err != nil {
		panic(err)
	}

	return client
}
