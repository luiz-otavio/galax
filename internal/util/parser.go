package util

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	. "github.com/luiz-otavio/galax/internal/impl"
	"github.com/luiz-otavio/galax/pkg/data"

	"github.com/google/uuid"
)

func ParseInfo(uuid string, group string, source map[string]string) (data.GroupInfo, error) {
	createdAt, err := ParseUnix(source["createdAt"], -1)

	if err != nil {
		return data.GroupInfo{}, errors.New("cannot parse created_at string to unix time")
	}

	expireAt, err := ParseUnix(source["expireAt"], -1)

	if err != nil {
		return data.GroupInfo{}, errors.New("cannot parse expire_at string to unix time")
	}

	author, err := EnsureType(source["author"], reflect.String, "cannot parse author type")

	if err != nil {
		return data.GroupInfo{}, errors.New("cannot parse author ar string to uuid")
	}

	return CreateGroupInfo(uuid, author.(string), data.GroupType(group), expireAt, createdAt), nil
}

func ParseMetadataSet(source map[string]string) (data.MetadataSet, error) {
	metadata := data.MetadataSet{}

	if skin, err := EnsureType(source["skin"], reflect.String, "cannot parse skin type"); err != nil {
		return metadata, err
	} else {
		metadata.Skin = skin.(string)
	}

	if name, err := EnsureType(source["name"], reflect.String, "cannot parse name type"); err != nil {
		return metadata, err
	} else {
		metadata.Name = name.(string)
	}

	if vanish, err := EnsureType(source["vanish"], reflect.Bool, "cannot parse vanish type"); err != nil {
		return metadata, err
	} else {
		metadata.Vanish = vanish.(bool)
	}

	if flying, err := EnsureType(source["flying"], reflect.Bool, "cannot parse flying type"); err != nil {
		return metadata, err
	} else {
		metadata.Flying = flying.(bool)
	}

	if seeAllPlayers, err := EnsureType(source["see_all_players"], reflect.Bool, "cannot parse see all players type"); err != nil {
		return metadata, err
	} else {
		metadata.SeeAllPlayers = seeAllPlayers.(bool)
	}

	if publicTell, err := EnsureType(source["public_tell"], reflect.Bool, "cannot parse public tell type"); err != nil {
		return metadata, err
	} else {
		metadata.EnablePublicTell = publicTell.(bool)
	}

	if staffChat, err := EnsureType(source["staff_chat"], reflect.Bool, "cannot parse staff chat type"); err != nil {
		return metadata, err
	} else {
		metadata.SeeAllStaffChat = staffChat.(bool)
	}

	if seeAllReports, err := EnsureType(source["see_all_reports"], reflect.Bool, "cannot parse see all reports type"); err != nil {
		return metadata, err
	} else {
		metadata.SeeAllReports = seeAllReports.(bool)
	}

	return metadata, nil
}

func ParseMetadataEntry(key string, value interface{}) (interface{}, error) {
	switch key {
	case "skin":
		return EnsureType(value, reflect.String, "unknown data for skin: "+fmt.Sprint(value))
	case "name":
		return EnsureType(value, reflect.String, "unknown data for name: "+fmt.Sprint(value))
	case "vanish":
		return EnsureType(value, reflect.Bool, "unknown data for vanish: "+fmt.Sprint(value))
	case "flying":
		return EnsureType(value, reflect.Bool, "unknown data for flying: "+fmt.Sprint(value))
	case "current_group":
		return EnsureType(value, reflect.String, "unknown data for current_group: "+fmt.Sprint(value))
	case "see_all_players":
		return EnsureType(value, reflect.Bool, "unknown data for flying: "+fmt.Sprint(value))
	case "enable_public_tell":
		return EnsureType(value, reflect.Bool, "unknown data for flying: "+fmt.Sprint(value))
	case "staff_chat":
		return EnsureType(value, reflect.Bool, "unknown data for flying: "+fmt.Sprint(value))
	}

	return value, nil
}

func EnsureType(target interface{}, condition reflect.Kind, err string) (interface{}, error) {
	targetType := reflect.TypeOf(target)

	if targetType.Kind() != condition {
		return nil, errors.New(err)
	} else {
		return target, nil
	}
}

func EnsureUUID(unique string) bool {
	_, err := uuid.Parse(unique)

	return err == nil
}

func ParseUnix(value string, def int64) (time.Time, error) {
	target, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		if def != -1 {
			return time.Unix(def, 0), nil
		}

		return time.Unix(-1, 0), err
	}

	return time.Unix(target, 0), nil
}

func ParseGroupType(group string) (data.GroupType, error) {
	switch strings.ToLower(group) {
	case "owner":
		return data.OWNER, nil
	case "admin":
		return data.ADMIN, nil
	case "moderator":
		return data.MODERATOR, nil
	case "helper":
		return data.HELPER, nil
	case "default":
		return data.DEFAULT, nil
	case "youtuber":
		return data.YOUTUBER, nil
	case "streamer":
		return data.STREAMER, nil
	case "vip":
		return data.VIP, nil
	case "mvp":
		return data.MVP, nil
	case "patron":
		return data.PATRON, nil
	case "elite":
		return data.ELITE, nil
	}

	return data.UNKNOWN, errors.New("unknown group type: " + group)
}

func ParseAccountType(account string) (data.AccountType, error) {
	switch strings.ToLower(account) {
	case "premium":
		return data.PREMIUM, nil
	case "cracked":
		return data.CRACKED, nil
	}

	return "", errors.New("unknown account type: " + account)
}
