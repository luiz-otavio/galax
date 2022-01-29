package main

import (
	"fmt"
	"os"

	galax "github.com/Rede-Legit/galax/pkg"
	"github.com/gofiber/fiber/v2"
)

var (
	LEGIT_API_KEY = env("LEGIT_API_KEY")
	DSN_REDIS     = env("DSN_REDIS")
	DSN_MARIADB   = env("DSN_MARIADB")
)

func main() {
	connector := galax.NewConnector(DSN_MARIADB)
	client := galax.NewRedis(DSN_REDIS)

	err := connector.Database.AutoMigrate(&galax.Account{}, &galax.GroupInfo{}, &galax.MetadataSet{})

	if err != nil {
		panic(err)
	}

	app := fiber.New(fiber.Config{
		CaseSensitive:     true,
		ReduceMemoryUsage: true,
		AppName:           "galax",
	})

	cache := galax.NewCache(client)
	router := galax.NewRouter(connector.Database, cache)

	galax.Initialize(connector.Database)

	app.Put("/account/create", router.CreateAccount)
	app.Get("/account/search", router.SearchAccount)
	app.Delete("/account/delete", router.DeleteAccount)
	app.Patch("/account/metadata/update", router.UpdateMetadata)
	app.Delete("/account/group/remove", router.DeleteGroup)
	app.Post("/account/group/insert", router.InsertGroup)
	app.Patch("/account/name", router.UpdateName)
	app.Patch("/account/cash/update", router.UpdateCash)
	app.Patch("/account/cash/sum", router.SumCash)

	app.Listen(":5896")
}

func env(key string) string {
	value := os.Getenv(key)

	if value == "" {
		panic(fmt.Sprintf("%s is not set", key))
	}

	return value
}
