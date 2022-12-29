package repository

import (
	"context"
	"strconv"

	. "github.com/luiz-otavio/galax/internal/impl"

	"github.com/luiz-otavio/galax/internal/util"
	"github.com/luiz-otavio/galax/pkg/config"
	"github.com/luiz-otavio/galax/pkg/data"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

type RedisRepository interface {
	GetRedis() *redis.Client
	GetConfig() *config.Config

	LoadAccount(uuid string) data.Account
	SaveAccount(account data.Account)

	RemoveGroup(account data.Account, groupInfo data.GroupInfo)
	AddGroup(account data.Account, groupInfo data.GroupInfo)

	UpdateCash(uuid string, cash int32)
	AddCash(uuid string, cash int32)
	TakeCash(uuid string, cash int32)

	UpdateMetadata(uuid string, key string, value string)
}

type repositoryImpl struct {
	redis  *redis.Client
	config *config.Config
}

func (cache repositoryImpl) LoadAccount(uuid string) data.Account {
	client := cache.redis

	context := context.Background()

	result, err := client.HGetAll(context, cache.config.GetAccountKey()+"-"+uuid).Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot load account from: " + uuid)
		return nil
	}

	if len(result) == 0 {
		return nil
	}

	cash, err := strconv.Atoi(result["cash"])

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse cash entry for account: " + uuid)
		return nil
	}

	metadatas, err := client.HGetAll(context, cache.config.GetAccountKey()+"-"+uuid+"-metadatas").Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse metadata entry for account: " + uuid)
		return nil
	}

	metadataSet, err := util.ParseMetadataSet(metadatas)

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse metadatas for account: " + uuid)
		return nil
	}

	groups, err := client.SMembers(context, cache.config.GetAccountKey()+"-"+uuid+"-groups").Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse groups for account: " + uuid)
		return nil
	}

	groupSet := []data.GroupInfo{}

	for _, key := range groups {
		hash, err := client.HGetAll(context, cache.config.GetAccountKey()+"-"+uuid+"-groups-"+key).Result()

		if err != nil {
			log.Error().Err(err).Msg("Cannot parse group info for account: " + uuid)
			continue
		}

		groupInfo, err := util.ParseInfo(uuid, key, hash)

		if err != nil {
			log.Error().Err(err).Msg("Cannot parse group info for account: " + uuid)
			continue
		}

		groupSet = append(groupSet, groupInfo)
	}

	createdAt, err := util.ParseUnix(result["createdAt"], -1)

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse created at from account: " + uuid)
		return nil
	}

	updatedAt, err := util.ParseUnix(result["updatedAt"], -1)

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse updated at from account: " + uuid)
		return nil
	}

	accountType, err := util.ParseAccountType(result["type"])

	if err != nil {
		log.Error().Err(err).Msg("Cannot parse account type from account: " + uuid)
		return nil
	}

	return AccountImpl{
		UUIDData: data.UUIDData{
			UUID: uuid,
		},

		AccountType: accountType,

		Name: result["name"],
		Cash: int32(cash),

		GroupSet:    groupSet,
		MetadataSet: metadataSet,

		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (cache repositoryImpl) SaveAccount(account data.Account) {
	context := context.Background()

	key := cache.config.GetAccountKey()
	_, err := cache.redis.TxPipelined(context, func(p redis.Pipeliner) error {
		var err error

		_, err = p.HMSet(context, key+"-"+account.GetUniqueId(), map[string]interface{}{
			"name":        account.GetName(),
			"cash":        account.GetCash(),
			"accountType": account.GetAccountType(),
			"createdAt":   account.GetCreatedAt().Unix(),
			"updatedAt":   account.GetUpdatedAt().Unix(),
		}).Result()

		if err != nil {
			log.Error().Err(err).Msg("Cannot execute first step for saving account for: " + account.GetUniqueId())
			return err
		}

		metadataSet := account.GetMetadataSet()

		_, err = p.HMSet(context, key+"-"+account.GetUniqueId()+"-metadatas", map[string]interface{}{
			"skin":            metadataSet.Skin,
			"public_tell":     metadataSet.EnablePublicTell,
			"name":            metadataSet.Name,
			"vanish":          metadataSet.Vanish,
			"see_all_players": metadataSet.SeeAllPlayers,
			"flying":          metadataSet.Flying,
			"current_group":   metadataSet.CurrentGroup,
			"staff_chat":      metadataSet.SeeAllStaffChat,
			"see_all_reports": metadataSet.SeeAllReports,
		}).Result()

		if err != nil {
			log.Error().Err(err).Msg("Cannot execute second step for saving account for: " + account.GetUniqueId())
			return err
		}

		groupKey := key + "-" + account.GetUniqueId() + "-groups"
		for _, group := range account.GetGroupSet() {
			if _, err = p.SAdd(context, groupKey, group.Group).Result(); err != nil {
				log.Error().Err(err).Msg("Cannot save group info for account: " + account.GetUniqueId())
			}

			_, err = p.HMSet(context, groupKey+"-"+string(group.Group), map[string]interface{}{
				"author":    group.Author,
				"createdAt": group.CreatedAt.Unix(),
				"expireAt":  group.ExpireAt.Unix(),
			}).Result()

			if err != nil {
				log.Error().Err(err).Msg("Cannot execute save step for saving group info: " + account.GetUniqueId())
				continue
			}

			if _, err = p.Expire(context, groupKey+"-"+string(group.Group), cache.config.GetExpireInterval()).Result(); err != nil {
				log.Error().Err(err).Msg("Cannot execute expire key step for saving group info: " + account.GetUniqueId())
			}
		}

		if _, err = p.Expire(context, key+"-"+account.GetUniqueId(), cache.config.GetExpireInterval()).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot execute expire key step for account: " + account.GetUniqueId())
			return err
		}

		if _, err = p.Expire(context, key+"-"+account.GetUniqueId()+"-metadatas", cache.config.GetExpireInterval()).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot execute expire key step for account metadatas: " + account.GetUniqueId())
			return err
		}

		if _, err = p.Expire(context, groupKey, cache.config.GetExpireInterval()).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot execute expire key step for account groups: " + account.GetUniqueId())
			return err
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Cannot execute save account for: " + account.GetUniqueId())
	}
}

func (cache repositoryImpl) UpdateCash(uuid string, cash int32) {
	_, err := cache.redis.HSet(context.Background(), cache.config.GetAccountKey()+"-"+uuid, "cash", cash).Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot update cash for account: " + uuid)
	}
}

func (cache repositoryImpl) AddCash(uuid string, cash int32) {
	_, err := cache.redis.HIncrBy(context.Background(), cache.config.GetAccountKey()+"-"+uuid, "cash", int64(cash)).Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot add cash for account: " + uuid)
	}
}

func (cache repositoryImpl) TakeCash(uuid string, cash int32) {
	_, err := cache.redis.HIncrBy(context.Background(), cache.config.GetAccountKey()+"-"+uuid, "cash", int64(-cash)).Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot take cash for account: " + uuid)
	}
}

func (cache repositoryImpl) UpdateMetadata(uuid string, key string, value string) {
	_, err := cache.redis.HSet(context.Background(), cache.config.GetAccountKey()+"-"+uuid+"-metadatas", key, value).Result()

	if err != nil {
		log.Error().Err(err).Msg("Cannot update metadata for account: " + uuid)
	}
}

func (cache repositoryImpl) AddGroup(account data.Account, group data.GroupInfo) {
	context := context.Background()

	key := cache.config.GetAccountKey()
	_, err := cache.redis.TxPipelined(context, func(p redis.Pipeliner) error {
		var err error

		if _, err = p.SAdd(context, key+"-"+account.GetUniqueId()+"-groups", group.Group).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot add group info for account: " + account.GetUniqueId())
		}

		_, err = p.HMSet(context, key+"-"+account.GetUniqueId()+"-groups-"+string(group.Group), map[string]interface{}{
			"author":    group.Author,
			"createdAt": group.CreatedAt.Unix(),
			"expireAt":  group.ExpireAt.Unix(),
		}).Result()

		if err != nil {
			log.Error().Err(err).Msg("Cannot execute save step for adding group info: " + account.GetUniqueId())
			return err
		}

		if _, err = p.Expire(context, key+"-"+account.GetUniqueId()+"-groups-"+string(group.Group), cache.config.GetExpireInterval()).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot execute expire key step for adding group info: " + account.GetUniqueId())
			return err
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Cannot execute add group for: " + account.GetUniqueId())
	}
}

func (cache repositoryImpl) RemoveGroup(account data.Account, group data.GroupInfo) {
	context := context.Background()

	key := cache.config.GetAccountKey()
	_, err := cache.redis.TxPipelined(context, func(p redis.Pipeliner) error {
		var err error

		if _, err = p.SRem(context, key+"-"+account.GetUniqueId()+"-groups", group).Result(); err != nil {
			log.Error().Err(err).Msg("Cannot remove group info for account: " + account.GetUniqueId())
		}

		_, err = p.Del(context, key+"-"+account.GetUniqueId()+"-groups-"+string(group.Group)).Result()

		if err != nil {
			log.Error().Err(err).Msg("Cannot execute remove step for removing group info: " + account.GetUniqueId())
			return err
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Cannot execute remove group for: " + account.GetUniqueId())
	}
}

func (cache repositoryImpl) GetConfig() *config.Config {
	return cache.config
}

func (cache repositoryImpl) GetRedis() *redis.Client {
	return cache.redis
}

func CreateRedisRepository(client *redis.Client, config *config.Config) RedisRepository {
	return repositoryImpl{
		redis:  client,
		config: config,
	}
}
