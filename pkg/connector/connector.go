package connector

import (
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewConnector(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	back, err := db.DB()

	if err != nil {
		return nil, err
	}

	back.SetConnMaxLifetime(time.Duration(5) * time.Minute)
	back.SetConnMaxIdleTime(time.Duration(2) * time.Minute)

	return db, nil
}

func NewRedis(url string) (*redis.Client, error) {
	options, err := redis.ParseURL(url)

	if err != nil {
		return nil, err
	}

	return redis.NewClient(options), nil
}
