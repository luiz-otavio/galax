package cmd

import (
	"github.com/go-redis/redis/v8"
	"github.com/luiz-otavio/galax/internal/connector"
	"github.com/luiz-otavio/galax/pkg/config"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func CreateConnectors(config *config.Config) (*gorm.DB, *redis.Client, error) {
	gorm, err := connector.NewConnector(config.GetMySQL())

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MySQL.")
		return nil, nil, err
	}

	log.Info().Msg("Connected to MySQL successfully.")

	redis, err := connector.NewRedis(config.GetRedis())

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis.")
		return nil, nil, err
	}

	log.Info().Msg("Connected to Redis successfully.")

	return gorm, redis, nil
}
