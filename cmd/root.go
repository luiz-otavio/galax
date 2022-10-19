package cmd

import (
	"fmt"
	"os"

	"github.com/Rede-Legit/galax/pkg/config"
	"github.com/rs/zerolog/log"
)

func Execute() {
	config, err := config.Load("./config.toml")

	log.Info().Msg("Galax is starting...")

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot decode TOML file for configuration.")
		return
	}

	os.Setenv("GALAX_DEBUGGING", fmt.Sprint(config.GetDebug()))

	log.Info().Msg("Loaded configuration successfully.")
	log.Info().Msg("Starting connectors...")

	db, redis, err := CreateConnectors(config)

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create connectors.")
		return
	}

	log.Info().Msg("Connectors started successfully.")

	if err := Migrate(db); err != nil {
		log.Fatal().Err(err).Msg("Cannot migrate sources to database.")
		return
	}

	log.Info().Msg("Migrated sources to database successfully.")
	log.Info().Msg("Starting routers...")

	fiberApp := Listen(config, db, redis)

	if fiberApp == nil {
		log.Fatal().Msg("Cannot start routers.")
		return
	}

	log.Info().Msg("Routers started successfully.")

	fiberApp.Listen(config.GetBinding())
}
