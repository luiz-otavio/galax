package util

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	galax "github.com/Rede-Legit/galax/pkg"
	"github.com/google/uuid"
)

func ParseInfo(uuid uuid.UUID, group string, data map[string]string) (galax.GroupInfo, error) {
	createdAt, err := ParseUnix(data["createdAt"], -1)

	if err != nil {
		return galax.GroupInfo{}, errors.New("cannot parse created_at string to unix time")
	}

	expireAt, err := ParseUnix(data["expireAt"], -1)

	if err != nil {
		return galax.GroupInfo{}, errors.New("cannot parse expire_at string to unix time")
	}

	author, err := ParseUUID(data["author"])

	if err != nil {
		return galax.GroupInfo{}, errors.New("cannot parse author ar string to uuid")
	}

	return galax.GroupInfo{
		User: uuid,

		ExpiredTimestamp: galax.ExpiredTimestamp{
			CreatedAt: createdAt,
			ExpireAt:  expireAt,
		},

		Author: author,
		Group:  group,
	}, nil
}

func ParseMetadataSet(data map[string]string) (galax.MetadataSet, error) {
	metadata := galax.MetadataSet{}

	if skin, err := EnsureType(data["skin"], reflect.String, "cannot parse skin type: "+fmt.Sprint(data["skin"])); err != nil {
		return metadata, err
	} else {
		metadata.Skin = skin.(string)
	}

	if name, err := EnsureType(data["name"], reflect.String, "cannot parse name type: "+fmt.Sprint(data["name"])); err != nil {
		return metadata, err
	} else {
		metadata.Name = name.(string)
	}

	if vanish, err := EnsureType(data["vanish"], reflect.Bool, "cannot parse vanish type: "+fmt.Sprint(data["vanish"])); err != nil {
		return metadata, err
	} else {
		metadata.Vanish = vanish.(bool)
	}

	if flying, err := EnsureType(data["flying"], reflect.Bool, "cannot parse flying type: "+fmt.Sprint(data["flying"])); err != nil {
		return metadata, err
	} else {
		metadata.Flying = flying.(bool)
	}

	if seeAllPlayers, err := EnsureType(data["see_all_players"], reflect.Bool, "cannot parse see all players type: "+fmt.Sprint(data["see_all_players"])); err != nil {
		return metadata, err
	} else {
		metadata.SeeAllPlayers = seeAllPlayers.(bool)
	}

	if publicTell, err := EnsureType(data["enable_public_tell"], reflect.Bool, "cannot parse public tell type: "+fmt.Sprint(data["enable_public_tell"])); err != nil {
		return metadata, err
	} else {
		metadata.EnablePublicTell = publicTell.(bool)
	}

	if staffChat, err := EnsureType(data["staff_chat"], reflect.Bool, "cannot parse staff chat type: "+fmt.Sprint(data["staff_chat"])); err != nil {
		return metadata, err
	} else {
		metadata.StaffChat = staffChat.(bool)
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

func ParseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}

func ParseUnix(value string, def int64) (time.Time, error) {
	target, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		if def != -1 {
			return time.Unix(def, 0), nil
		}

		return time.Unix(-1, 0), nil
	}

	return time.Unix(target, 0), err
}
