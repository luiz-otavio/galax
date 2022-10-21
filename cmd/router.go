package cmd

import (
	"os"
	"os/signal"

	"github.com/Rede-Legit/galax/pkg/config"
	"github.com/Rede-Legit/galax/pkg/repository"
	"github.com/Rede-Legit/galax/pkg/router"
	"github.com/Rede-Legit/galax/pkg/worker"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/keyauth/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func Listen(config *config.Config, db *gorm.DB, redis *redis.Client) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		DisableKeepalive:      true,
	})

	v1 := app.Group("/v1")

	worker := worker.CreateWorker(db)

	accountRouter := router.CreateAccountRouter(
		db,
		repository.CreateRedisRepository(
			redis,
			config,
		),
		worker,
	)

	// Listen to Ctrl + C
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)

	go func() {
		<-ch

		log.Info().Msg("Shutting down server...")

		worker.Shutdown()

		database, err := db.DB()

		if err != nil {
			log.Error().Err(err).Msg("Failed to close database connection.")
		} else {
			database.Close()
			log.Info().Msg("Closed database connection successfully.")
		}

		if err = redis.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Redis connection.")
		} else {
			log.Info().Msg("Closed Redis connection successfully.")
		}
	}()

	authRouter := router.CreateAuthRouter(db)

	accountRouter.TakeEndpoints(v1.Group("/account"))
	authRouter.TakeEndpoints(v1.Group("/auth"))

	// Add middleware to check if the key is valid
	app.Use(keyauth.New(keyauth.Config{
		Validator: func(ctx *fiber.Ctx, key string) (bool, error) {
			if config.GetDebug() {
				log.Debug().Msg("Income request from " + ctx.IP())
			}

			return key == config.GetKey(), nil
		},
	}))

	return app
}
