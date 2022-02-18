package main

import (
	"fmt"
	"os"

	. "github.com/Rede-Legit/galax/pkg"
	. "github.com/Rede-Legit/galax/pkg/connector"
	. "github.com/Rede-Legit/galax/pkg/repository"
	. "github.com/Rede-Legit/galax/pkg/router"
	. "github.com/Rede-Legit/galax/pkg/worker"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/keyauth/v2"
)

var (
	LEGIT_API_KEY = env("LEGIT_API_KEY")
	DSN_REDIS     = env("DSN_REDIS")
	DSN_MARIADB   = env("DSN_MARIADB")
)

func main() {
	connector := NewConnector(DSN_MARIADB)
	client := NewRedis(DSN_REDIS)

	err := connector.Database.AutoMigrate(&Account{}, &GroupInfo{}, &MetadataSet{})

	if err != nil {
		panic(err)
	}

	app := fiber.New(fiber.Config{
		CaseSensitive:     true,
		ReduceMemoryUsage: true,
		AppName:           "galax",
	})

	app.Use(keyauth.New(keyauth.Config{
		Validator: func(ctx *fiber.Ctx, key string) (bool, error) {
			return key == LEGIT_API_KEY, nil
		},
	}))

	cache := NewCache(client)
	router := NewRouter(connector.Database, cache)

	Initialize(connector.Database)

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

	println("Listening on port 5896!")

	app.Listen(":5896")
}

func env(key string) string {
	value := os.Getenv(key)

	if value == "" {
		panic(fmt.Sprintf("%s is not set", key))
	}

	return value
}
