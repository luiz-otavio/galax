package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	. "github.com/luiz-otavio/galax/internal/impl"
)

type AuthRouter interface {
	WebRouter

	Login(ctx *fiber.Ctx) error
	Register(ctx *fiber.Ctx) error
	ChangePassword(ctx *fiber.Ctx) error
}

type authRouterImpl struct {
	db *gorm.DB
}

func (r authRouterImpl) TakeEndpoints(router fiber.Router) {
	router.Post("/login", r.Login)
	router.Put("/register", r.Register)
	router.Patch("/update", r.ChangePassword)
}

func (r authRouterImpl) Login(ctx *fiber.Ctx) error {
	request := LoginImpl{}

	if err := ctx.BodyParser(&request); err != nil {
		return err
	}

	// Search for authentication
	var authentication AuthenticationImpl

	if err := r.db.Where("username = ?", request.GetUsername()).First(&authentication).Error; err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot find the authentication",
		})
	}

	// Check if the password is correct
	if !authentication.CheckPassword(request.GetPassword()) {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid password",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged in",
	})
}

func (r authRouterImpl) Register(ctx *fiber.Ctx) error {
	request := LoginImpl{}

	if err := ctx.BodyParser(&request); err != nil {
		return err
	}

	// Check if the username is already taken
	var authentication AuthenticationImpl

	if err := r.db.Where("username = ?", request.GetUsername()).First(&authentication).Error; err == nil {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Username is already taken",
		})
	}

	// Create the authentication
	authentication = AuthenticationImpl{
		Username: request.GetUsername(),
		Password: request.GetPassword(),

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	authentication.UpdatePassword(authentication.Password)

	if err := r.db.Create(&authentication).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot create the authentication",
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Successfully registered",
	})
}

func (r authRouterImpl) ChangePassword(ctx *fiber.Ctx) error {
	var body struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		NewPassword string `json:"new_password"`
	}

	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	// Search for authentication
	var authentication AuthenticationImpl

	if err := r.db.Where("username = ?", body.Username).First(&authentication).Error; err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Cannot find the authentication",
		})
	}

	// Check if the password is correct
	if !authentication.CheckPassword(body.Password) {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid password",
		})
	}

	// Update the password
	authentication.UpdatePassword(body.Password)

	if err := r.db.Save(&authentication).Error; err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Cannot update the authentication",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully changed password",
	})
}

func CreateAuthRouter(db *gorm.DB) AuthRouter {
	return authRouterImpl{db: db}
}
