package repository

import (
	"context"
	"strconv"
	"time"

	. "github.com/Rede-Legit/galax/pkg"
	"github.com/Rede-Legit/galax/pkg/util"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisCache struct {
	redis *redis.Client
}

const (
	ACCOUNT_HASH_KEY    = "accounts"
	ACCOUNT_EXPIRE_TIME = time.Duration(5) * time.Minute
)

func NewCache(redis *redis.Client) *RedisCache {
	return &RedisCache{
		redis: redis,
	}
}

func (cache *RedisCache) DeleteAccount(account *Account) {
	client := cache.redis

	uuid := account.UniqueId.String()

	context := context.Background()

	client.Del(context, ACCOUNT_HASH_KEY+"-"+uuid)

	client.Del(context, ACCOUNT_HASH_KEY+"-"+uuid+"-metadatas")
	client.Del(context, ACCOUNT_HASH_KEY+"-"+uuid+"-groups")
}

func (cache *RedisCache) SaveAccount(id string, account *Account) {
	client := cache.redis

	context := context.Background()

	client.HMSet(context, ACCOUNT_HASH_KEY+"-"+id, map[string]interface{}{
		"name":      account.Name,
		"cash":      account.Cash,
		"createdAt": account.Timestamp.CreatedAt.Unix(),
		"updatedAt": account.Timestamp.UpdatedAt.Unix(),
	})

	metadataSet := account.MetadataSet

	client.HMSet(context, ACCOUNT_HASH_KEY+"-"+id+"-metadatas", map[string]interface{}{
		"skin":               metadataSet.Skin,
		"enable_public_tell": metadataSet.EnablePublicTell,
		"name":               metadataSet.Name,
		"vanish":             metadataSet.Vanish,
		"see_all_players":    metadataSet.SeeAllPlayers,
		"flying":             metadataSet.Flying,
		"current_group":      metadataSet.CurrentGroup,
		"staff_chat":         metadataSet.StaffChat,
	})

	groupKey := ACCOUNT_HASH_KEY + "-" + id + "-groups"

	for _, group := range account.GroupSet {
		client.SAdd(context, groupKey, group.Group)

		client.HMSet(context, groupKey+"-"+group.Group, map[string]interface{}{
			"author":    group.Author.String(),
			"createdAt": group.CreatedAt.Unix(),
			"expireAt":  group.ExpireAt.Unix(),
		})

		client.Expire(context, groupKey+"-"+group.Group, ACCOUNT_EXPIRE_TIME)
	}

	client.Expire(context, ACCOUNT_HASH_KEY+"-"+id, ACCOUNT_EXPIRE_TIME)
	client.Expire(context, ACCOUNT_HASH_KEY+"-"+id+"-metadatas", ACCOUNT_EXPIRE_TIME)
	client.Expire(context, groupKey, ACCOUNT_EXPIRE_TIME)
}

func (cache *RedisCache) LoadAccount(uuid uuid.UUID) *Account {
	client := cache.redis

	uniqueId := uuid.String()
	context := context.Background()

	result, err := client.HGetAll(context, ACCOUNT_HASH_KEY+"-"+uniqueId).Result()

	if err != nil {
		util.Log(err)
	}

	if len(result) == 0 {
		return nil
	}

	cash, _ := strconv.Atoi(result["cash"])

	metadatas, _ := client.HGetAll(context, ACCOUNT_HASH_KEY+"-"+uniqueId+"-metadatas").Result()
	groups, _ := client.SMembers(context, ACCOUNT_HASH_KEY+"-"+uniqueId+"-groups").Result()

	metadataSet := MetadataSet{}
	groupSet := []GroupInfo{}

	metadataSet.Read(metadatas)

	for _, key := range groups {
		hash, err := client.HGetAll(context, ACCOUNT_HASH_KEY+"-"+uniqueId+"-groups-"+key).Result()

		if err != nil {
			util.Log(err)
		}

		groupSet = append(groupSet, ReadInfo(uuid, key, hash))
	}

	createdAt, _ := strconv.ParseInt(result["createdAt"], 10, 64)
	updatedAt, _ := strconv.ParseInt(result["updatedAt"], 10, 64)

	return &Account{
		UniqueId: uuid,
		Name:     result["name"],

		Cash: int32(cash),

		MetadataSet: metadataSet,
		GroupSet:    groupSet,

		Timestamp: Timestamp{
			CreatedAt: time.Unix(createdAt, 0).In(time.UTC),
			UpdatedAt: time.Unix(updatedAt, 0).In(time.UTC),
		},
	}
}

func (cache *RedisCache) UpdateCash(id string, cash int32) {
	cache.redis.HSet(context.Background(), ACCOUNT_HASH_KEY+"-"+id, "cash", cash).Result()
}

func (cache *RedisCache) UpdateMetadata(id string, key string, value string) {
	cache.redis.HSet(context.Background(), ACCOUNT_HASH_KEY+"-"+id+"-metadatas", key, value).Result()
}

func (cache *RedisCache) InsertGroup(id string, value GroupInfo) {
	client := cache.redis

	context := context.Background()

	client.SAdd(context, ACCOUNT_HASH_KEY+"-"+id+"-groups", value.Group)

	client.HMSet(context, ACCOUNT_HASH_KEY+"-"+id+"-groups-"+value.Group, map[string]interface{}{
		"author":    value.Author.String(),
		"createdAt": value.CreatedAt.Unix(),
		"expireAt":  value.ExpireAt.Unix(),
	})
}

func (cache *RedisCache) DeleteGroup(id string, key string) {
	client := cache.redis

	context := context.Background()

	client.SRem(context, ACCOUNT_HASH_KEY+"-"+id+"-groups", key)
	client.Del(context, ACCOUNT_HASH_KEY+"-"+id+"-groups-"+key)
}

func (cache *RedisCache) UpdateName(id string, name string) {
	cache.redis.HSet(context.Background(), ACCOUNT_HASH_KEY+"-"+id, "name", name).Result()
}
