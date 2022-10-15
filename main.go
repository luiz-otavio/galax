package main

import (
	"os"

	galax "github.com/Rede-Legit/galax/pkg"
	"github.com/Rede-Legit/galax/pkg/config"
	connector "github.com/Rede-Legit/galax/pkg/connector"
	"github.com/Rede-Legit/galax/pkg/repository"
	"github.com/Rede-Legit/galax/pkg/router"
	"github.com/Rede-Legit/galax/pkg/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/keyauth/v2"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr,
	}).With().Timestamp().Logger()

	config, err := config.Load("./config.toml")

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot decode TOML file for configuration.")
		return
	}

	gorm := connector.NewConnector(config.GetMySQL())
	client := connector.NewRedis(config.GetRedis())

	if err = gorm.AutoMigrate(&galax.Account{}, &galax.GroupInfo{}, &galax.MetadataSet{}); err != nil {
		log.Fatal().Err(err).Msg("Cannot auto migrate sources to database")
		return
	}

	app := fiber.New(fiber.Config{
		CaseSensitive:     true,
		ReduceMemoryUsage: true,
		AppName:           "galax",
	})

	app.Use(keyauth.New(keyauth.Config{
		Validator: func(ctx *fiber.Ctx, key string) (bool, error) {
			if config.GetDebug() {
				log.Debug().Msg("Income request from " + ctx.IP())
			}

			return key == config.GetKey(), nil
		},
	}))

	cache := repository.NewCache(client, config)
	router := router.NewRouter(gorm, cache, config)

	worker.Initialize(gorm)

	app.Put("/account/create", router.CreateAccount)
	app.Get("/account/search", router.SearchAccount)
	app.Delete("/account/delete", router.DeleteAccount)
	app.Patch("/account/metadata/update", router.UpdateMetadata)
	app.Delete("/account/group/remove", router.DeleteGroup)
	app.Post("/account/group/insert", router.InsertGroup)
	app.Patch("/account/name", router.UpdateName)
	app.Patch("/account/cash/update", router.UpdateCash)
	app.Patch("/account/cash/sum", router.SumCash)
	app.Get("/query", router.Query)

	log.Info().Msg("Initializing the listening port...")

	app.Listen(config.GetBinding())

	log.Info().Msg("Started to listen in " + config.GetBinding() + " port!")
}
