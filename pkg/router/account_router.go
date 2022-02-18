package galax

import (
	"fmt"
	"time"

	. "github.com/Rede-Legit/galax/pkg"
	"github.com/Rede-Legit/galax/pkg/repository"
	. "github.com/Rede-Legit/galax/pkg/worker"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRouter struct {
	db    *gorm.DB
	cache *repository.RedisCache
}

func NewRouter(db *gorm.DB, cache *repository.RedisCache) *UserRouter {
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

	name := body["name"].(string)

	if name == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Name is required.",
		})
	}

	if err := router.db.Where("username = ?", name).First(&Account{}).Error; err == nil {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Account already exists.",
		})
	}

	uniqueId, _ := body["unique_id"].(string)

	if len(uniqueId) == 0 {
		uniqueId = OfflinePlayerUUID(name).
			String()
	}

	if len(uniqueId) < 32 || len(uniqueId) > 36 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid Unique id.",
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

	account = New(uuid, name)

	if err := router.db.Create(&account).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Could not create account.",
		})
	}

	router.cache.SaveAccount(uuid.String(), &account)

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Account created.",
	})
}

func (r *UserRouter) SearchAccount(ctx *fiber.Ctx) error {
	uniqueId, err := r.queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	account := loadAccount(uniqueId, r)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(account)
}

func (r *UserRouter) UpdateName(ctx *fiber.Ctx) error {
	uniqueId, err := r.queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

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

	account := loadAccount(uniqueId, r)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	account.Name = username

	Do(func(d *gorm.DB) {
		d.Model(&Account{}).Where("unique_id = ?", uniqueId).
			Update("username", account.Name)
	})

	r.cache.UpdateName(uniqueId.String(), username)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) UpdateCash(ctx *fiber.Ctx) error {
	uniqueId, err := r.queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

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

	account := loadAccount(uniqueId, r)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	account.Cash = int32(cash)

	Do(func(d *gorm.DB) {
		d.Model(&Account{}).Where("unique_id = ?", uniqueId).Update("cash", account.Cash)
	})

	r.cache.UpdateCash(uniqueId.String(), account.Cash)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) UpdateMetadata(ctx *fiber.Ctx) error {
	uniqueId, err := r.queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

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

	account := loadAccount(uniqueId, r)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

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

		Do(func(d *gorm.DB) {
			d.Model(&MetadataSet{}).Where("user = ?", uniqueId).Update(key, value)
		})

		r.cache.UpdateMetadata(uniqueId.String(), key, fmt.Sprint(value))
	}

	account.MetadataSet = metadata

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) InsertGroup(ctx *fiber.Ctx) error {
	uniqueId, err := r.queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

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

	account := loadAccount(uniqueId, r)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	groups := account.GroupSet

	for key, group := range groupSet {
		// if !IsGroup(key) {
		// 	return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 		"message": "Group set is not valid.",
		// 	})
		// }

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

		if author == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Author is required.",
			})
		}

		if len(author) <= 16 {
			if alternative := r.checkUUID(author); alternative != uuid.Nil {
				author = alternative.String()
			} else {
				author = OfflinePlayerUUID(author).String()
			}
		}

		uuid, err := uuid.Parse(author)

		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Author is not valid.",
			})
		}

		expireAt, _ := info["expire_at"].(float64)
		createdAt, _ := info["created_at"].(float64)

		groupInfo := GroupInfo{
			ExpiredTimestamp: ExpiredTimestamp{
				ExpireAt:  time.Unix(int64(expireAt), 0),
				CreatedAt: time.Unix(int64(createdAt), 0),
			},

			User:   account.UniqueId,
			Group:  key,
			Author: uuid,
		}

		groups = append(groups, groupInfo)

		Do(func(d *gorm.DB) {
			d.Create(&groupInfo)
		})

		r.cache.InsertGroup(uniqueId.String(), groupInfo)
	}

	account.GroupSet = groups

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) DeleteAccount(ctx *fiber.Ctx) error {
	uniqueId, err := r.queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

	account := loadAccount(uniqueId, r)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	Do(func(d *gorm.DB) {
		d.Where("user = ?", uniqueId).Delete(&GroupInfo{})

		d.Where("user = ?", uniqueId).Delete(&MetadataSet{})

		d.Delete(&account)
	})

	r.cache.DeleteAccount(account)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account deleted.",
	})
}

func (r *UserRouter) DeleteGroup(ctx *fiber.Ctx) error {
	uniqueId, err := r.queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

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

	account := loadAccount(uniqueId, r)

	if account == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Account not found.",
		})
	}

	groups := account.GroupSet

	for key := range groupSet {
		// if !IsGroup(key) {
		// 	return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		// 		"message": "Group set is not valid.",
		// 	})
		// }

		if !hasGroup(account, key) {
			continue
		}

		for i, group := range groups {
			if group.Group == key {
				groups = append(groups[:i], groups[i+1:]...)

				Do(func(d *gorm.DB) {
					d.Where("user = ? AND role = ?", uniqueId, key).Delete(&group)
				})

				r.cache.DeleteGroup(uniqueId.String(), key)

				break
			}
		}
	}

	account.GroupSet = groups

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) SumCash(ctx *fiber.Ctx) error {
	uniqueId, err := r.queryUUID(ctx)

	if uniqueId == uuid.Nil {
		return err
	}

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

	account := loadAccount(uniqueId, r)

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

	Do(func(d *gorm.DB) {
		d.Model(&Account{}).Where("unique_id = ?", uniqueId).Update("cash", account.Cash)
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account updated.",
	})
}

func (r *UserRouter) Query(ctx *fiber.Ctx) error {
	username := ctx.Query("id")

	if username == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Username is required.",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"UUID": OfflinePlayerUUID(username).String(),
	})
}

func (r *UserRouter) queryUUID(ctx *fiber.Ctx) (uuid.UUID, error) {
	id := ctx.Query("id")

	if id != "" && len(id) <= 16 {
		alternative := r.checkUUID(id)

		if alternative != uuid.Nil {
			return alternative, nil
		}

		id = OfflinePlayerUUID(id).
			String()
	}

	if id == "" || len(id) < 32 {
		return uuid.Nil, ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unique ID is required.",
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

func loadAccount(uuid uuid.UUID, router *UserRouter) *Account {
	account := router.cache.LoadAccount(uuid)

	if account != nil {
		return router.ensureGroups(account)
	}

	err := router.db.
		Preload(clause.Associations).
		Where("unique_id = ?", uuid).
		First(&account).Error

	if err != nil {
		return nil
	}

	router.cache.SaveAccount(uuid.String(), account)

	return router.ensureGroups(account)
}

func (r *UserRouter) checkUUID(username string) uuid.UUID {
	var unique_id string

	if err := r.db.Model(&Account{}).Select("unique_id").
		Where("username = ?", username).
		Row().Scan(&unique_id); err != nil {
		return uuid.Nil
	}

	return uuid.MustParse(unique_id)
}

func hasGroup(account *Account, group string) bool {
	for _, g := range account.GroupSet {
		if g.Group == group {
			return true
		}
	}

	return false
}

func (r *UserRouter) ensureGroups(account *Account) *Account {
	now := time.Now().
		UnixMilli()

	for index, group := range account.GroupSet {
		millis := group.ExpireAt.
			UnixMilli()

		if millis == 0 || millis > now {
			continue
		}

		account.GroupSet = append(account.GroupSet[:index], account.GroupSet[index+1:]...)

		Do(func(d *gorm.DB) {
			d.Where("user = ? AND role = ?", account.UniqueId.String(), group.Group).Delete(&group)
		})

		r.cache.DeleteGroup(account.UniqueId.String(), group.Group)
	}

	return account
}
