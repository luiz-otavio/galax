package router

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"time"

	. "github.com/luiz-otavio/galax/internal/impl"
	"github.com/rs/zerolog/log"

	"github.com/luiz-otavio/galax/internal/repository"
	"github.com/luiz-otavio/galax/internal/util"
	"github.com/luiz-otavio/galax/internal/worker"
	"github.com/luiz-otavio/galax/pkg/data"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountRouter interface {
	WebRouter

	CreateAccount(ctx *fiber.Ctx) error
	SearchAccount(ctx *fiber.Ctx) error
	UpdateCash(ctx *fiber.Ctx) error
	AddCash(ctx *fiber.Ctx) error
	TakeCash(ctx *fiber.Ctx) error
	UpdateMetadata(ctx *fiber.Ctx) error
	AddGroup(ctx *fiber.Ctx) error
	RemoveGroup(ctx *fiber.Ctx) error
}

type accountRouterImpl struct {
	db     *gorm.DB
	cache  repository.RedisRepository
	worker worker.Worker
}

func (r *accountRouterImpl) TakeEndpoints(router fiber.Router) {
	router.Put("/create", r.CreateAccount)
	router.Get("/search", r.SearchAccount)
	router.Patch("/metadata", r.UpdateMetadata)
	router.Delete("/group", r.RemoveGroup)
	router.Post("/group", r.AddGroup)
	router.Patch("/cash/update", r.UpdateCash)
	router.Patch("/cash/sum", r.AddCash)
	router.Patch("/cash/take", r.TakeCash)
}

func (r *accountRouterImpl) CreateAccount(ctx *fiber.Ctx) error {
	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	name, ok := body["name"].(string)

	if !ok || len(name) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Name is required.",
		})
	}

	util.DebugOutput("Income request for creating account with name '%s'", name)

	if err := r.db.Where("username = ?", name).First(AccountImpl{}).Error; err == nil {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Account already exists.",
		})
	}

	uniqueId, ok := body["unique_id"].(string)

	if !ok || len(uniqueId) == 0 {
		uniqueId = util.OfflinePlayerUUID(name).
			String()
	}

	if len(uniqueId) < 32 || len(uniqueId) > 36 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Unique id.",
		})
	}

	var account AccountImpl

	if err := r.db.Where("unique_id = ?", uniqueId).First(&account).Error; err == nil {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Account already exists.",
		})
	}

	if util.EnsureUUID(uniqueId) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unique id is not valid.",
		})
	}

	accountType, ok := body["accountType"].(string)

	if !ok || len(accountType) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Account type is required.",
		})
	}

	targetType, err := util.ParseAccountType(accountType)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Account type is invalid.",
		})
	}

	account = AccountImpl{
		UUIDData: data.UUIDData{
			UUID: uniqueId,
		},

		AccountType: targetType,

		Name: name,
		Cash: 0,

		GroupSet:    []data.GroupInfo{},
		MetadataSet: data.MetadataSet{},

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.db.Create(&account).Error; err != nil {
		log.Error().Err(err).Msg("Could not create account.")

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not create account.",
		})
	}

	r.cache.SaveAccount(account)

	util.DebugOutput("Account created for %s", name)
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Account created.",
	})
}

func (r *accountRouterImpl) SearchAccount(ctx *fiber.Ctx) error {
	uniqueId, err := r.FilterUUIDByQuery(ctx)

	if err != nil {
		return err
	}

	util.DebugOutput("Income request for searching account with '%s'", uniqueId)
	account := r.cache.LoadAccount(uniqueId)

	if account == nil {
		account = r.RetrieveByDatabase(uniqueId)

		if account == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Account not found.",
			})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(account)
}

func (r *accountRouterImpl) UpdateCash(ctx *fiber.Ctx) error {
	uniqueId, err := r.FilterUUIDByQuery(ctx)

	if err != nil {
		return err
	}

	util.DebugOutput("Income request for updating cash for user %s.", uniqueId)

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	if _, err = util.EnsureType(body["cash"], reflect.Float64, "cash isn't a number."); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cash isn't a number.",
		})
	}

	// Parse cash to int32
	cash := int32(
		math.Round(
			body["cash"].(float64),
		),
	)

	if cash < 0 {
		cash = 0
	}

	account := r.cache.LoadAccount(uniqueId)

	if account == nil {
		account = r.RetrieveByDatabase(uniqueId)

		if account == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Account not found.",
			})
		}
	}

	account.SetCash(cash)

	r.worker.Do(func(d *gorm.DB) {
		d.Model(AccountImpl{}).
			Where("unique_id = ?", uniqueId).
			Update("cash", account.GetCash())
	})

	r.cache.UpdateCash(uniqueId, cash)

	util.DebugOutput("Updated cash for user %s with %d", account.GetUniqueId(), account.GetCash())
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *accountRouterImpl) AddCash(ctx *fiber.Ctx) error {
	uniqueId, err := r.FilterUUIDByQuery(ctx)

	if err != nil {
		return err
	}

	util.DebugOutput("Income request for adding cash for user %s.", uniqueId)

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	if _, err = util.EnsureType(body["cash"], reflect.Float64, "cash isn't a number."); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cash isn't a number.",
		})
	}

	// Parse cash to int32
	cash := int32(
		math.Round(
			body["cash"].(float64),
		),
	)

	account := r.cache.LoadAccount(uniqueId)

	if account == nil {
		account = r.RetrieveByDatabase(uniqueId)

		if account == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Account not found.",
			})
		}
	}

	if account.GetCash()+cash < 0 {
		cash = -account.GetCash()
	}

	account.AddCash(cash)

	r.worker.Do(func(d *gorm.DB) {
		d.Model(AccountImpl{}).
			Where("unique_id = ?", uniqueId).
			Update("cash", account.GetCash())
	})

	r.cache.AddCash(uniqueId, account.GetCash())

	util.DebugOutput("Added cash for user %s with %d", account.GetUniqueId(), account.GetCash())
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *accountRouterImpl) TakeCash(ctx *fiber.Ctx) error {
	uniqueId, err := r.FilterUUIDByQuery(ctx)

	if err != nil {
		return err
	}

	util.DebugOutput("Income request for taking cash for user %s.", uniqueId)

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	if _, err = util.EnsureType(body["cash"], reflect.Float64, "cash isn't a number."); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cash isn't a number.",
		})
	}

	// Parse cash to int32
	cash := int32(
		math.Round(
			body["cash"].(float64),
		),
	)

	account := r.cache.LoadAccount(uniqueId)

	if account == nil {
		account = r.RetrieveByDatabase(uniqueId)

		if account == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Account not found.",
			})
		}
	}

	if account.GetCash()-cash < 0 {
		cash = account.GetCash()
	}

	account.TakeCash(cash)

	r.worker.Do(func(d *gorm.DB) {
		d.Model(AccountImpl{}).
			Where("unique_id = ?", uniqueId).
			Update("cash", account.GetCash())
	})

	r.cache.TakeCash(uniqueId, account.GetCash())

	util.DebugOutput("Taken cash for user %s with %d", account.GetUniqueId(), account.GetCash())
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *accountRouterImpl) UpdateMetadata(ctx *fiber.Ctx) error {
	uniqueId, err := r.FilterUUIDByQuery(ctx)

	if err != nil {
		return err
	}

	util.DebugOutput("Income request to update metadata entry for user %s", uniqueId)

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	metadataSet, ok := body["metadata_set"].(map[string]interface{})

	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Metadata set is required.",
		})
	}

	account := r.cache.LoadAccount(uniqueId)

	if account == nil {
		account = r.RetrieveByDatabase(uniqueId)

		if account == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Account not found.",
			})
		}
	}

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

		target, err := util.ParseMetadataEntry(key, value)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Missed data type for: '" + key + "' property.",
			})
		}

		r.worker.Do(func(d *gorm.DB) {
			d.Model(data.MetadataSet{}).
				Where("user = ?", uniqueId).
				Update(key, target)
		})

		util.DebugOutput("Updated metadata entry with %s key and %s value.", key, fmt.Sprint(value))
		r.cache.UpdateMetadata(uniqueId, key, fmt.Sprint(value))
	}

	util.DebugOutput("Updated metadata set for account %s", account.GetUniqueId())
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *accountRouterImpl) AddGroup(ctx *fiber.Ctx) error {
	uniqueId, err := r.FilterUUIDByQuery(ctx)

	if err != nil {
		return err
	}

	util.DebugOutput("Income request to add group for user %s.", uniqueId)

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	groupSet, ok := body["group_set"].(map[string]interface{})

	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Group set is required.",
		})
	}

	account := r.cache.LoadAccount(uniqueId)

	if account == nil {
		account = r.RetrieveByDatabase(uniqueId)

		if account == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Account not found.",
			})
		}
	}

	for key, group := range groupSet {
		groupType, err := util.ParseGroupType(key)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid group type: '" + key + "'.",
			})
		}

		if !account.HasGroupSet(groupType) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Account does not have group set: '" + key + "'.",
			})
		}

		info, ok := group.(map[string]interface{})

		if !ok || info == nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Group set is not valid.",
			})
		}

		author := fmt.Sprint(info["author"])

		if len(author) == 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Author is required.",
			})
		}

		var target string
		if !util.EnsureUUID(author) {
			target, err = r.FilterByUsername(author)

			if err != nil {
				return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Author is not valid.",
				})
			}
		} else {
			target = author
		}

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Author is not valid.",
			})
		}

		var expireAt, createdAt string

		if _, err := util.EnsureType(info["expire_at"], reflect.Float64, "expire at isn't a number"); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Expire at is not valid.",
			})
		} else {
			expireAt = fmt.Sprint(info["expire_at"])
		}

		if _, err := util.EnsureType(info["created_at"], reflect.Float64, "created at isn't a number"); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Created at is not valid.",
			})
		} else {
			createdAt = fmt.Sprint(info["created_at"])
		}

		expireUnix, err := util.ParseUnix(expireAt, -1)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Expire at cannot be parsed to Unix.",
			})
		}

		createdUnix, err := util.ParseUnix(createdAt, -1)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Created at cannot be parsed to Unix.",
			})
		}

		groupInfo := data.GroupInfo{
			Group:  groupType,
			Author: target,

			User: account.GetUniqueId(),

			ExpireAt:  expireUnix,
			CreatedAt: createdUnix,
		}

		r.worker.Do(func(d *gorm.DB) {
			d.Create(&groupInfo)
		})

		util.DebugOutput(
			"Added group info for account %s, group %s, author %s, expire at %d, created at %d.",
			account.GetUniqueId(),
			groupType,
			target,
			expireAt,
			createdAt,
		)

		r.cache.AddGroup(account, groupInfo)
	}

	util.DebugOutput("Group set has updated for account %s.", account.GetUniqueId())
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *accountRouterImpl) RemoveGroup(ctx *fiber.Ctx) error {
	uniqueId, err := r.FilterUUIDByQuery(ctx)

	if err != nil {
		return err
	}

	util.DebugOutput("Income request to remove group from account %s.", uniqueId)

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	groups, ok := body["group_set"].([]string)

	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Group set is required.",
		})
	}

	account := r.cache.LoadAccount(uniqueId)

	if account == nil {
		account = r.RetrieveByDatabase(uniqueId)

		if account == nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Account not found.",
			})
		}
	}

	for _, key := range groups {
		groupType, err := util.ParseGroupType(key)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid group type: '" + key + "'.",
			})
		}

		if !account.HasGroupSet(groupType) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Account does not have group set: '" + key + "'.",
			})
		}

		util.DebugOutput("Removing group with %s name from %s account.", key, uniqueId)

		account.RemoveGroup(groupType)

		r.worker.Do(func(d *gorm.DB) {
			d.Where("user = ? AND role = ?", uniqueId, key).Delete(&data.GroupInfo{})
		})
	}

	util.DebugOutput("Group set has updated for account %s.", account.GetUniqueId())
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *accountRouterImpl) FilterUUIDByQuery(ctx *fiber.Ctx) (string, error) {
	id := ctx.Query("id")

	if !util.EnsureUUID(id) {
		new, err := r.FilterByUsername(id)

		if err != nil {
			return "", ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Id is not valid.",
			})
		}

		return new, nil
	}

	return id, nil
}

func (r *accountRouterImpl) RetrieveByDatabase(unique string) data.Account {
	var account AccountImpl

	if err := r.db.Preload(clause.Associations).Where("unique_id = ?", unique).First(&account).Error; err != nil {
		return nil
	}

	r.cache.SaveAccount(account)

	return account
}

func (r *accountRouterImpl) FilterByUsername(username string) (string, error) {
	var unique_id string

	if err := r.db.Model(AccountImpl{}).
		Select("unique_id").
		Where("username = ?", username).
		Row().
		Scan(&unique_id); err != nil {
		return unique_id, err
	}

	if !util.EnsureUUID(unique_id) {
		return "", errors.New("invalid uuid")
	}

	return unique_id, nil
}

func CreateAccountRouter(db *gorm.DB, repository repository.RedisRepository, worker worker.Worker) AccountRouter {
	return &accountRouterImpl{
		db:     db,
		cache:  repository,
		worker: worker,
	}
}
