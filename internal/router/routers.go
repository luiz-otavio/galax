package router

import "github.com/gofiber/fiber/v2"

type WebRouter interface {
	TakeEndpoints(router fiber.Router)
}
