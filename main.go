package galax

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

var (
	LEGIT_API_KEY = env("LEGIT_API_KEY")
	DSN_REDIS     = env("DSN_REDIS")
	DSN_MARIADB   = env("DSN_MARIADB")
)

func main() {
	connector := NewConnector(DSN_MARIADB)
	client := NewRedis(DSN_REDIS)

	connector.db.AutoMigrate(&Account{})

	app := fiber.New(fiber.Config{
		CaseSensitive:     true,
		ReduceMemoryUsage: true,
		AppName:           "galax",
	})

	cache := &RedisCache{redis: client}

	router := UserRouter{db: connector.db, cache: cache}

	app.Put("/account/create", router.CreateAccount)
	app.Get("/account/search", router.SearchAccount)
	app.Patch("/account/metadata/update", router.UpdateMetadata)
	app.Delete("/account/group/remove", router.DeleteAccount)
	app.Post("/account/group/insert", router.InsertGroup)
	app.Patch("/account/name", router.UpdateName)
	app.Patch("/account/cash/update", router.UpdateCash)

	app.Listen(":8080")
}

func env(key string) string {
	value := os.Getenv(key)

	if value == "" {
		panic(fmt.Sprintf("%s is not set", key))
	}

	return value
}
