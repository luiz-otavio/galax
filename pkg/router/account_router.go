package router

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"time"

	galax "github.com/Rede-Legit/galax/pkg"
	"github.com/Rede-Legit/galax/pkg/config"
	"github.com/Rede-Legit/galax/pkg/repository"
	"github.com/Rede-Legit/galax/pkg/util"
	worker "github.com/Rede-Legit/galax/pkg/worker"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRouter struct {
	db     *gorm.DB
	cache  *repository.RedisCache
	config *config.Config
}

func NewRouter(db *gorm.DB, cache *repository.RedisCache, config *config.Config) *UserRouter {
	return &UserRouter{
		db:     db,
		cache:  cache,
		config: config,
	}
}

func (r *UserRouter) CreateAccount(ctx *fiber.Ctx) error {
	body := map[string]interface{}{}

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

	r.Debug("Income request to create account for " + name)

	if err := r.db.Where("username = ?", name).First(&galax.Account{}).Error; err == nil {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Account already exists.",
		})
	}

	uniqueId, ok := body["unique_id"].(string)

	if !ok || len(uniqueId) == 0 {
		uniqueId = OfflinePlayerUUID(name).String()
	}

	if len(uniqueId) < 32 || len(uniqueId) > 36 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Unique id.",
		})
	}

	var account galax.Account

	if err := r.db.Where("unique_id = ?", uniqueId).First(&account).Error; err == nil {
		return ctx.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
			"message": "Account already exists.",
		})
	}

	uuid, err := util.ParseUUID(uniqueId)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unique id is not valid.",
		})
	}

	account = galax.New(uuid, name)

	if err := r.db.Create(&account).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not create account.",
		})
	}

	r.cache.SaveAccount(uuid.String(), &account)

	r.Debug("Created account for " + name + " with unique id " + account.GetUniqueId())

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Account created.",
	})
}

func (r *UserRouter) SearchAccount(ctx *fiber.Ctx) error {
	uniqueId, err := r.RetrieveUUID(ctx)

	if err != nil {
		return err
	}

	r.Debug("Income request for searching account for '" + uniqueId.String() + "' user.")

	account := r.LoadAccount(uniqueId)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(account)
}

func (r *UserRouter) UpdateName(ctx *fiber.Ctx) error {
	uniqueId, err := r.RetrieveUUID(ctx)

	if err != nil {
		return err
	}

	r.Debug("Income request for updating name for '" + uniqueId.String() + "' account.")

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	username := fmt.Sprint(body["name"])

	if len(username) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Name is required.",
		})
	}

	account := r.LoadAccount(uniqueId)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	account.Name = username

	worker.Do(func(d *gorm.DB) {
		d.Model(&galax.Account{}).Where("unique_id = ?", uniqueId).
			Update("username", account.Name)
	})

	r.cache.UpdateName(uniqueId.String(), username)

	r.Debug("Updated name to user '" + account.GetUniqueId() + "'.")
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) UpdateCash(ctx *fiber.Ctx) error {
	uniqueId, err := r.RetrieveUUID(ctx)

	if err != nil {
		return err
	}

	r.Debug("Income request for updating cash for user '" + uniqueId.String() + "'.")

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	cash := body["cash"].(float64)

	if cash < 0 {
		cash = 0
	}

	account := r.LoadAccount(uniqueId)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	account.Cash = int32(cash)

	worker.Do(func(d *gorm.DB) {
		d.Model(&galax.Account{}).Where("unique_id = ?", uniqueId).Update("cash", account.Cash)
	})

	r.cache.UpdateCash(uniqueId.String(), account.Cash)

	r.Debug("Updated cash for user '" + account.GetUniqueId() + "'.")
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) UpdateMetadata(ctx *fiber.Ctx) error {
	uniqueId, err := r.RetrieveUUID(ctx)

	if err != nil {
		return err
	}

	r.Debug("Income request to update metadata for user '" + uniqueId.String() + "'.")

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

	account := r.LoadAccount(uniqueId)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
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

		worker.Do(func(d *gorm.DB) {
			d.Model(&galax.MetadataSet{}).Where("user = ?", uniqueId).Update(key, target)
		})

		r.Debug("Updated metadata entry with '" + key + "' key and '" + fmt.Sprint(value) + "' value.")

		r.cache.UpdateMetadata(uniqueId.String(), key, fmt.Sprint(value))
	}

	r.Debug("Updated metadata for account '" + account.GetUniqueId() + "'.")
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) InsertGroup(ctx *fiber.Ctx) error {
	uniqueId, err := r.RetrieveUUID(ctx)

	if err != nil {
		return err
	}

	r.Debug("Income request to insert group for account '" + uniqueId.String() + "'.")

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

	account := r.LoadAccount(uniqueId)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	for key, group := range groupSet {
		if !account.HasGroupSet(key) {
			continue
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

		target, err := util.ParseUUID(author)

		if err != nil {
			target = r.EnsureUUID(author)

			if target == uuid.Nil {
				target = OfflinePlayerUUID(author)
			}
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

		groupInfo := galax.GroupInfo{
			Group:  key,
			Author: target,

			User: account.UniqueId,

			ExpiredTimestamp: galax.ExpiredTimestamp{
				ExpireAt:  expireUnix,
				CreatedAt: createdUnix,
			},
		}

		worker.Do(func(d *gorm.DB) {
			d.Create(&groupInfo)
		})

		r.Debug("Inserted group with '" + key + "' name with '" + author + "' author.")

		r.cache.InsertGroup(uniqueId.String(), groupInfo)
	}

	r.Debug("Group set updated for account '" + account.GetUniqueId() + "'.")
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) DeleteAccount(ctx *fiber.Ctx) error {
	uniqueId, err := r.RetrieveUUID(ctx)

	if err != nil {
		return err
	}

	r.Debug("Income request for deleting account from '" + uniqueId.String() + "'.")

	account := r.LoadAccount(uniqueId)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	worker.Do(func(d *gorm.DB) {
		d.Where("user = ?", uniqueId).Delete(&galax.GroupInfo{})

		d.Where("user = ?", uniqueId).Delete(&galax.MetadataSet{})

		d.Delete(&account)
	})

	r.cache.DeleteAccount(account)

	r.Debug("Deleted account from user '" + account.GetUniqueId() + "'.")
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account deleted.",
	})
}

func (r *UserRouter) DeleteGroup(ctx *fiber.Ctx) error {
	uniqueId, err := r.RetrieveUUID(ctx)

	if err != nil {
		return err
	}

	r.Debug("Income request to delete group info from '" + uniqueId.String() + "' account.")

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

	account := r.LoadAccount(uniqueId)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	groups := account.GroupSet

	for key := range groupSet {
		if !account.HasGroupSet(key) {
			continue
		}

		for _, group := range groups {
			if group.Group == key {
				r.Debug("Deleting group '" + key + "' from '" + uniqueId.String() + "' account.")

				worker.Do(func(d *gorm.DB) {
					d.Where("user = ? AND role = ?", uniqueId, key).Delete(&group)
				})

				r.cache.DeleteGroup(uniqueId.String(), key)

				break
			}
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) SumCash(ctx *fiber.Ctx) error {
	uniqueId, err := r.RetrieveUUID(ctx)

	if err != nil {
		return err
	}

	r.Debug("Income request to sum cash from '" + uniqueId.String() + "' account.")

	var body map[string]interface{}

	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Could not parse body.",
		})
	}

	cash, ok := body["cash"].(float64)

	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Cash is required.",
		})
	}

	account := r.LoadAccount(uniqueId)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	account.Cash += int32(cash)

	if account.Cash < 0 {
		account.Cash = 0
	}

	r.cache.UpdateCash(uniqueId.String(), account.Cash)

	worker.Do(func(d *gorm.DB) {
		d.Model(&galax.Account{}).Where("unique_id = ?", uniqueId).Update("cash", account.Cash)
	})

	r.Debug("Updated account cash from '" + fmt.Sprint((account.Cash - int32(cash))) + "' to '" + fmt.Sprint(account.Cash))
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) Query(ctx *fiber.Ctx) error {
	username := ctx.Query("id")

	if len(username) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Username is required.",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"UUID": OfflinePlayerUUID(username).String(),
	})
}

func (r *UserRouter) RetrieveUUID(ctx *fiber.Ctx) (uuid.UUID, error) {
	id := ctx.Query("id")

	if len(id) != 0 && len(id) <= 16 {
		alternative := r.EnsureUUID(id)

		if alternative != uuid.Nil {
			return alternative, nil
		}

		id = OfflinePlayerUUID(id).
			String()
	}

	if len(id) < 32 {
		return uuid.Nil, ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unique ID is required.",
		})
	}

	uniqueId, err := util.ParseUUID(id)

	if err != nil {
		return uuid.Nil, ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unique id is not valid.",
		})
	}

	return uniqueId, nil
}

func (router *UserRouter) LoadAccount(uuid uuid.UUID) *galax.Account {
	account := router.cache.LoadAccount(uuid)

	if account != nil {
		return router.EnsureGroups(account)
	}

	err := router.db.
		Preload(clause.Associations).
		Where("unique_id = ?", uuid).
		First(&account).Error

	if err != nil {
		return nil
	}

	router.cache.SaveAccount(uuid.String(), account)

	return router.EnsureGroups(account)
}

func (r *UserRouter) EnsureUUID(username string) uuid.UUID {
	var unique_id string

	if err := r.db.Model(&galax.Account{}).Select("unique_id").
		Where("username = ?", username).
		Row().
		Scan(&unique_id); err != nil {
		return uuid.Nil
	}

	return uuid.MustParse(unique_id)
}

func (r *UserRouter) EnsureGroups(account *galax.Account) *galax.Account {
	groupSet := account.GroupSet

	now := time.Now()

	for target := 0; target < len(groupSet); target++ {
		group := groupSet[target]

		if group.ExpireAt.IsZero() || group.ExpireAt.Second() == 0 || group.ExpireAt.After(now) {
			continue
		}

		groupSet = append(groupSet[:target], groupSet[target+1:]...)

		worker.Do(func(d *gorm.DB) {
			d.Where("user = ? AND role = ?", account.UniqueId, group.Group).Delete(&group)
		})

		r.cache.DeleteGroup(account.UniqueId.String(), group.Group)
	}

	account.GroupSet = groupSet

	return account
}

func (r *UserRouter) Debug(message string) {
	if !r.config.GetDebug() {
		return
	}

	log.Debug().Msg(message)
}

func OfflinePlayerUUID(username string) uuid.UUID {
	const version = 3 // UUID v3
	uuid := md5.Sum([]byte("OfflinePlayer:" + username))
	uuid[6] = (uuid[6] & 0x0f) | uint8((version&0xf)<<4)
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // RFC 4122 variant
	return uuid
}
