package util

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	galax "github.com/Rede-Legit/galax/pkg/data"
	"github.com/google/uuid"
)

func ParseInfo(uuid string, group string, data map[string]string) (galax.GroupInfo, error) {
	createdAt, err := ParseUnix(data["createdAt"], -1)

	if err != nil {
		return galax.GroupInfo{}, errors.New("cannot parse created_at string to unix time")
	}

	expireAt, err := ParseUnix(data["expireAt"], -1)

	if err != nil {
		return galax.GroupInfo{}, errors.New("cannot parse expire_at string to unix time")
	}

	author, err := EnsureType(data["author"], reflect.String, "cannot parse author type: "+fmt.Sprint(data["author"]))

	if err != nil {
		return galax.GroupInfo{}, errors.New("cannot parse author ar string to uuid")
	}

	return galax.CreateGroupInfo(uuid, author.(string), galax.GroupType(group), expireAt, createdAt), nil
}

func ParseMetadataSet(data map[string]string) (galax.MetadataSet, error) {
	metadata := galax.CreateEmptyMetadataSet()

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

	if publicTell, err := EnsureType(data["public_tell"], reflect.Bool, "cannot parse public tell type: "+fmt.Sprint(data["enable_public_tell"])); err != nil {
		return metadata, err
	} else {
		metadata.EnablePublicTell = publicTell.(bool)
	}

	if staffChat, err := EnsureType(data["staff_chat"], reflect.Bool, "cannot parse staff chat type: "+fmt.Sprint(data["staff_chat"])); err != nil {
		return metadata, err
	} else {
		metadata.SeeAllStaffChat = staffChat.(bool)
	}

	if seeAllReports, err := EnsureType(data["see_all_reports"], reflect.Bool, "cannot parse see all reports type: "+fmt.Sprint(data["see_all_reports"])); err != nil {
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

func ParseGroupType(group string) (galax.GroupType, error) {
	switch strings.ToLower(group) {
	case "owner":
		return galax.OWNER, nil
	case "admin":
		return galax.ADMIN, nil
	case "moderator":
		return galax.MODERATOR, nil
	case "helper":
		return galax.HELPER, nil
	case "default":
		return galax.DEFAULT, nil
	case "youtuber":
		return galax.YOUTUBER, nil
	case "streamer":
		return galax.STREAMER, nil
	case "vip":
		return galax.VIP, nil
	case "mvp":
		return galax.MVP, nil
	case "patron":
		return galax.PATRON, nil
	case "elite":
		return galax.ELITE, nil
	}

	return galax.UNKNOWN, errors.New("unknown group type: " + group)
}

func ParseAccountType(account string) (galax.AccountType, error) {
	switch strings.ToLower(account) {
	case "premium":
		return galax.PREMIUM, nil
	case "cracked":
		return galax.CRACKED, nil
	}

	return "", errors.New("unknown account type: " + account)
}
