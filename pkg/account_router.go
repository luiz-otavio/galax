package galax

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRouter struct {
	db    *gorm.DB
	cache *RedisCache
}

func NewRouter(db *gorm.DB, cache *RedisCache) *UserRouter {
	return &UserRouter{
		db:    db,
		cache: cache,
	}
}

func (router *UserRouter) CreateAccount(ctx *fiber.Ctx) error {
	body := map[string]interface{}{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	uniqueId := fmt.Sprint(body["uniqueId"])

	if len(uniqueId) == 0 || len(uniqueId) < 32 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unique id is required.",
		})
	}

	var account Account

	if err := router.db.Where("unique_id = ?", uniqueId).First(&account).Error; err == nil {
		return ctx.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
			"message": "Account already exists.",
		})
	}

	uuid, err := uuid.Parse(uniqueId)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unique id is not valid.",
		})
	}

	account = *New(uuid, fmt.Sprint(body["name"]))

	if err := router.db.Create(&account).Error; err != nil {
		println("Could not create account.", err.Error())

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not create account.",
		})
	}

	router.cache.SaveAccount(uuid.String(), &account)

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Account created.",
	})
}

func (router UserRouter) SearchAccount(ctx *fiber.Ctx) error {
	uniqueId, err := queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	account := loadAccount(uniqueId, router)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(account)
}

func (router UserRouter) UpdateName(ctx *fiber.Ctx) error {
	uniqueId, err := queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		panic(err)
	}

	username := fmt.Sprint(body["name"])

	if len(username) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Name is required.",
		})
	}

	account := loadAccount(uniqueId, router)

	account.Name = username

	Do(func(d *gorm.DB) {
		d.Save(&account)
	})

	router.cache.UpdateName(uniqueId.String(), username)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (router UserRouter) UpdateCash(ctx *fiber.Ctx) error {
	uniqueId, err := queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		panic(err)
	}

	cash := body["cash"].(float64)

	if cash < 0 {
		cash = 0
	}

	account := loadAccount(uniqueId, router)

	account.Cash = int32(cash)

	Do(func(d *gorm.DB) {
		d.Save(&account)
	})

	router.cache.UpdateCash(uniqueId.String(), account.Cash)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (router UserRouter) UpdateMetadata(ctx *fiber.Ctx) error {
	uniqueId, err := queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		panic(err)
	}

	metadataSet, ok := body["metadata_set"].(map[string]interface{})

	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Metadata set is required.",
		})
	}

	account := loadAccount(uniqueId, router)

	metadata := account.MetadataSet

	for key, value := range metadataSet {
		if len(key) == 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Metadata key is required.",
			})
		}

		if len(fmt.Sprint(value)) == 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Metadata value is required.",
			})
		}

		if !metadata.Write(key, value) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Metadata key is not valid.",
			})
		}

		router.cache.UpdateMetadata(uniqueId.String(), key, fmt.Sprint(value))
	}

	account.MetadataSet = metadata

	Do(func(d *gorm.DB) {
		d.Save(&account)
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (router UserRouter) InsertGroup(ctx *fiber.Ctx) error {
	uniqueId, err := queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Body is required.",
		})
	}

	groupSet, ok := body["group_set"].(map[string]interface{})

	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Group set is required.",
		})
	}

	account := loadAccount(uniqueId, router)

	groups := account.GroupSet

	for key, group := range groupSet {
		if !IsGroup(key) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Group set is not valid.",
			})
		}

		if hasGroup(account, key) {
			continue
		}

		info, ok := group.(map[string]interface{})

		if !ok || info == nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Group set is not valid.",
			})
		}

		author := fmt.Sprint(info["author"])

		if author == "" || len(author) < 32 {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Author is not valid.",
			})
		}

		expireAt, _ := strconv.ParseInt(fmt.Sprint(info["expire_at"]), 10, 64)
		createdAt, _ := strconv.ParseInt(fmt.Sprint(info["created_at"]), 10, 64)

		if expireAt < createdAt {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Expire at is not valid.",
			})
		}

		uuid, err := uuid.Parse(author)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Author is not valid.",
			})
		}

		groupInfo := GroupInfo{
			ExpiredTimestamp: ExpiredTimestamp{
				ExpireAt:  expireAt,
				CreatedAt: createdAt,
			},

			User:   account.UniqueId,
			Group:  key,
			Author: uuid,
		}

		groups = append(groups, groupInfo)

		router.cache.InsertGroup(uniqueId.String(), groupInfo)
	}

	account.GroupSet = groups

	Do(func(d *gorm.DB) {
		if err := router.db.Save(&account).Error; err != nil {
			println("Could not update account. ", err.Error())
		}
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (router UserRouter) DeleteAccount(ctx *fiber.Ctx) error {
	uniqueId, err := queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	account := loadAccount(uniqueId, router)

	if err := router.db.Delete(&account).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not delete account.",
		})
	}

	router.cache.DeleteAccount(account)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account deleted.",
	})
}

func (router UserRouter) DeleteGroup(ctx *fiber.Ctx) error {
	uniqueId, err := queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Body is required.",
		})
	}

	groupSet, ok := body["group_set"].(map[string]interface{})

	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Group set is required.",
		})
	}

	account := loadAccount(uniqueId, router)

	groups := account.GroupSet

	for key, _ := range groupSet {
		if !IsGroup(key) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Group set is not valid.",
			})
		}

		if !hasGroup(account, key) {
			continue
		}

		for i, group := range groups {
			if group.Group == key {
				groups = append(groups[:i], groups[i+1:]...)
				break
			}
		}
	}

	account.GroupSet = groups

	Do(func(d *gorm.DB) {
		d.Save(&account)
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (router UserRouter) SumCash(ctx *fiber.Ctx) error {
	uniqueId, err := queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Body is required.",
		})
	}

	cash, ok := body["cash"].(float64)

	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cash is required.",
		})
	}

	account := loadAccount(uniqueId, router)

	account.Cash += int32(cash)

	if account.Cash < 0 {
		account.Cash = 0
	}

	router.cache.UpdateCash(uniqueId.String(), account.Cash)

	Do(func(d *gorm.DB) {
		if err := d.Save(&account).Error; err != nil {
			println("Could not update account.", err.Error())
		}
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func queryUUID(ctx *fiber.Ctx) (uuid.UUID, error) {
	id := ctx.Query("id")

	if len(id) == 0 || len(id) < 32 {
		return uuid.Nil, ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unique id is required.",
		})
	}

	uniqueId, err := uuid.Parse(id)

	if err != nil {
		return uuid.Nil, ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unique id is not valid.",
		})
	}

	return uniqueId, nil
}

func loadAccount(uuid uuid.UUID, router UserRouter) *Account {
	account := router.cache.LoadAccount(uuid)

	if account != nil {
		return account
	}

	err := router.db.
		Preload(clause.Associations).
		Where("unique_id = ?", uuid).
		First(&account).Error

	if err != nil {
		println("Could not load account.", err.Error())
	}

	router.cache.SaveAccount(uuid.String(), account)

	return account
}

func hasGroup(account *Account, group string) bool {
	for _, g := range account.GroupSet {
		if g.Group == group {
			return true
		}
	}

	return false
}
