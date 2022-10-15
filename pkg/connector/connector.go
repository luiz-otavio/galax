package connector

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewConnector(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{})

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to Gorm with the URL")
	}

	back, _ := db.DB()

	back.SetConnMaxLifetime(time.Duration(5) * time.Minute)
	back.SetConnMaxIdleTime(time.Duration(2) * time.Minute)

	return db
}

func NewRedis(url string) *redis.Client {
	options, err := redis.ParseURL(url)

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to redis with the URL")
	}

	return redis.NewClient(options)
}
