package cmd

import (
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

	accountRouter := router.CreateAccountRouter(
		db,
		repository.CreateRedisRepository(
			redis,
			config,
		),
		worker.CreateWorker(db),
	)

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
