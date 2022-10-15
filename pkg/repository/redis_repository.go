package repository

import (
	"context"
	"strconv"
	"time"

	galax "github.com/Rede-Legit/galax/pkg"
	config "github.com/Rede-Legit/galax/pkg/config"
	"github.com/Rede-Legit/galax/pkg/util"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type RedisCache struct {
	redis  *redis.Client
	config *config.Config
}

func (cache *RedisCache) GetRedis() *redis.Client {
	return cache.redis
}

func (cache *RedisCache) GetAccountKey() string {
	return cache.config.GetAccountKey()
}

func (cache *RedisCache) GetExpireInterval() time.Duration {
	return cache.config.GetExpireInterval()
}

func (cache *RedisCache) GetConfig() *config.Config {
	return cache.config
}

func NewCache(redis *redis.Client, config *config.Config) *RedisCache {
	return &RedisCache{
		redis:  redis,
		config: config,
	}
}

func (cache *RedisCache) DeleteAccount(account *galax.Account) {
	context := context.Background()

	_, err := cache.GetRedis().TxPipelined(context, func(p redis.Pipeliner) error {
		var err error

		if _, err = p.Del(context, cache.GetAccountKey()+"-"+account.GetUniqueId()).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot delete key for account: " + account.GetUniqueId())
			return err
		}

		if _, err = p.Del(context, cache.GetAccountKey()+"-"+account.GetUniqueId()+"-metadatas").Result(); err != nil {
			log.Error().Err(err).Msg("Cannot delete key for account metadatas: " + account.GetUniqueId())
			return err
		}

		if _, err := p.Del(context, cache.GetAccountKey()+"-"+account.GetUniqueId()+"-groups").Result(); err != nil {
			log.Error().Err(err).Msg("Cannot delete key for account: " + account.GetUniqueId())
			return err
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Cannot reproduce delete context for account: " + account.GetUniqueId())
	}
}

func (cache *RedisCache) SaveAccount(id string, account *galax.Account) {
	context := context.Background()

	_, err := cache.GetRedis().TxPipelined(context, func(p redis.Pipeliner) error {
		var err error

		_, err = p.HMSet(context, cache.GetAccountKey()+"-"+id, map[string]interface{}{
			"name":      account.Name,
			"cash":      account.Cash,
			"createdAt": account.Timestamp.CreatedAt.Unix(),
			"updatedAt": account.Timestamp.UpdatedAt.Unix(),
		}).Result()

		if err != nil {
			log.Error().Err(err).Msg("Cannot execute first step for saving account for: " + id)
			return err
		}

		metadataSet := account.MetadataSet

		_, err = p.HMSet(context, cache.GetAccountKey()+"-"+id+"-metadatas", map[string]interface{}{
			"skin":               metadataSet.Skin,
			"enable_public_tell": metadataSet.EnablePublicTell,
			"name":               metadataSet.Name,
			"vanish":             metadataSet.Vanish,
			"see_all_players":    metadataSet.SeeAllPlayers,
			"flying":             metadataSet.Flying,
			"current_group":      metadataSet.CurrentGroup,
			"staff_chat":         metadataSet.StaffChat,
		}).Result()

		if err != nil {
			log.Error().Err(err).Msg("Cannot execute second step for saving account for: " + id)
			return err
		}

		groupKey := cache.GetAccountKey() + "-" + id + "-groups"

		for _, group := range account.GroupSet {
			if _, err = p.SAdd(context, groupKey, group.Group).Result(); err != nil {
				log.Error().Err(err).Msg("Cannot save group info for account: " + id)
			}

			_, err = p.HMSet(context, groupKey+"-"+group.Group, map[string]interface{}{
				"author":    group.Author.String(),
				"createdAt": group.CreatedAt.Unix(),
				"expireAt":  group.ExpireAt.Unix(),
			}).Result()

			if err != nil {
				log.Error().Err(err).Msg("Cannot execute save step for saving group info: " + id)
				continue
			}

			if _, err = p.Expire(context, groupKey+"-"+group.Group, cache.GetExpireInterval()).Result(); err != nil {
				log.Error().Err(err).Msg("Cannot execute expire key step for saving group info: " + id)
			}
		}

		if _, err = p.Expire(context, cache.GetAccountKey()+"-"+id, cache.GetExpireInterval()).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot execute expire key step for account: " + id)
			return err
		}

		if _, err = p.Expire(context, cache.GetAccountKey()+"-"+id+"-metadatas", cache.GetExpireInterval()).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot execute expire key step for account metadatas: " + id)
			return err
		}

		if _, err = p.Expire(context, groupKey, cache.GetExpireInterval()).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot execute expire key step for account groups: " + id)
			return err
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Cannot execute save account for: " + account.GetUniqueId())
	}
}

func (cache *RedisCache) LoadAccount(uuid uuid.UUID) *galax.Account {
	client := cache.redis

	uniqueId := uuid.String()
	context := context.Background()

	result, err := client.HGetAll(context, cache.GetAccountKey()+"-"+uniqueId).Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot load account from: " + uniqueId)
		return nil
	}

	if len(result) == 0 {
		return nil
	}

	cash, err := strconv.Atoi(result["cash"])

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse cash entry for account: " + uniqueId)
		return nil
	}

	metadatas, err := client.HGetAll(context, cache.GetAccountKey()+"-"+uniqueId+"-metadatas").Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse metadata entry for account: " + uniqueId)
		return nil
	}

	metadataSet, err := util.ParseMetadataSet(metadatas)

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse metadatas for account: " + uniqueId)
		return nil
	}

	groups, err := client.SMembers(context, cache.GetAccountKey()+"-"+uniqueId+"-groups").Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse groups for account: " + uniqueId)
		return nil
	}

	groupSet := []galax.GroupInfo{}

	for _, key := range groups {
		hash, err := client.HGetAll(context, cache.GetAccountKey()+"-"+uniqueId+"-groups-"+key).Result()

		if err != nil {
			log.Error().Err(err).Msg("Cannot parse group info for account: " + uniqueId)
			continue
		}

		groupInfo, err := util.ParseInfo(uuid, key, hash)

		if err != nil {
			log.Error().Err(err).Msg("Cannot parse group info for account: " + uniqueId)
			continue
		}

		groupSet = append(groupSet, groupInfo)
	}

	createdAt, err := util.ParseUnix(result["createdAt"], -1)

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse created at from account: " + uniqueId)
		return nil
	}

	updatedAt, err := util.ParseUnix(result["updatedAt"], -1)

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse updated at from account: " + uniqueId)
		return nil
	}

	return &galax.Account{
		UniqueId: uuid,
		Name:     result["name"],

		Cash: int32(cash),

		MetadataSet: metadataSet,
		GroupSet:    groupSet,

		Timestamp: galax.Timestamp{
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
	}
}

func (cache *RedisCache) UpdateCash(id string, cash int32) {
	cache.redis.HSet(context.Background(), cache.GetAccountKey()+"-"+id, "cash", cash).Result()
}

func (cache *RedisCache) UpdateMetadata(id string, key string, value string) {
	cache.redis.HSet(context.Background(), cache.GetAccountKey()+"-"+id+"-metadatas", key, value).Result()
}

func (cache *RedisCache) InsertGroup(id string, value galax.GroupInfo) {
	context := context.Background()

	_, err := cache.GetRedis().TxPipelined(context, func(p redis.Pipeliner) error {
		var err error

		if _, err = p.SAdd(context, cache.GetAccountKey()+"-"+id+"-groups", value.Group).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot insert group to acccount: " + id)
			return nil
		}

		_, err = p.HMSet(context, cache.GetAccountKey()+"-"+id+"-groups-"+value.Group, map[string]interface{}{
			"author":    value.Author.String(),
			"createdAt": value.CreatedAt.Unix(),
			"expireAt":  value.ExpireAt.Unix(),
		}).Result()

		if err != nil {
			log.Error().Err(err).Msg("Cannot insert group info to acccount: " + id)
			return nil
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Cannot update group to acccount: " + id)
	}
}

func (cache *RedisCache) DeleteGroup(id string, key string) {
	context := context.Background()

	_, err := cache.GetRedis().TxPipelined(context, func(p redis.Pipeliner) error {
		var err error

		if _, err = p.SRem(context, cache.GetAccountKey()+"-"+id+"-groups", key).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot delete group to acccount: " + id)
			return nil
		}

		if _, err = p.Del(context, cache.GetAccountKey()+"-"+id+"-groups-"+key).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot delete group info to acccount: " + id)
			return nil
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Cannot delete group from acccount: " + id)
	}
}

func (cache *RedisCache) UpdateName(id string, name string) {
	cache.redis.HSet(context.Background(), cache.GetAccountKey()+"-"+id, "name", name).Result()
}
