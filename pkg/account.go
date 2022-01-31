package galax

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func New(uniqueId uuid.UUID, name string) Account {
	return Account{
		UniqueId: uniqueId,
		Name:     name,

		MetadataSet: MetadataSet{
			User:               uniqueId,
			Skin:               "",
			Name:               "",
			CurrentGroup:       "NORMAL",
			Vanish:             false,
			SEE_ALL_PLAYERS:    true,
			ENABLE_PUBLIC_TELL: true,
			STAFF_SCOREBOARD:   false,
		},

		GroupSet: []GroupInfo{
			{
				ExpiredTimestamp: ExpiredTimestamp{
					ExpireAt:  time.Now().Unix(),
					CreatedAt: time.Now().Unix(),
				},

				User:   uniqueId,
				Group:  "NORMAL",
				Author: uniqueId,
			},
		},
		Cash: 0,
	}
}

type Timestamp struct {
	UpdatedAt int64 `json:"updated_at" gorm:"column:updated_at;type:bigint;not null"`
	CreatedAt int64 `json:"created_at" gorm:"column:created_at;type:bigint;not null"`
}

type ExpiredTimestamp struct {
	ExpireAt  int64 `json:"expire_at" gorm:"column:expire_at;type:bigint;not null"`
	CreatedAt int64 `json:"created_at" gorm:"column:created_at;type:bigint;not null"`
}

type Account struct {
	UniqueId uuid.UUID `json:"uniqueId" gorm:"primary_key;type:char(36);not null"`

	Name string `json:"name" gorm:"type:varchar(16);not null;column:username"`
	Cash int32  `json:"cash" gorm:"type:bigint;not null;default:0"`

	GroupSet    []GroupInfo `json:"group_set" gorm:"foreignkey:User;references:UniqueId"`
	MetadataSet MetadataSet `json:"metadata_set" gorm:"foreignkey:User;references:UniqueId"`
}

var AVAILABLE_GROUPS = []string{
	"DIRECTOR",
	"SUB-DIRECTOR",
	"ADMIN",
	"MODERATOR",
	"HELPER",
	"DESIGNER",
	"BUILDER",
	"INFLUENCER_PLUS",
	"INFLUENCER",
	"MVP_PLUS_PLUS",
	"MVP_PLUS",
	"MVP",
	"VIP",
	"NORMAL",
}

type MetadataSet struct {
	User uuid.UUID `json:"-" gorm:"column:user;type:char(36);not null"`

	Skin         string `json:"skin" gorm:"column:skin;type:varchar(16);not null"`
	Name         string `json:"name" gorm:"column:name;type:varchar(16);not null"`
	Vanish       bool   `json:"vanish" gorm:"column:vanish;type:boolean;not null"`
	CurrentGroup string `json:"current_group" gorm:"column:current_group;type:varchar(18);not null"`

	SEE_ALL_PLAYERS    bool `json:"see_all_players" gorm:"column:see_all_players;type:boolean;not null"`
	ENABLE_PUBLIC_TELL bool `json:"enable_public_tell" gorm:"column:enable_public_tell;type:boolean;not null"`
	STAFF_SCOREBOARD   bool `json:"staff_scoreboard" gorm:"column:staff_scoreboard;type:boolean;not null"`
}

type GroupInfo struct {
	ExpiredTimestamp

	User uuid.UUID `json:"-" gorm:"column:user;type:char(36);not null"`

	Group  string    `json:"group" gorm:"column:role;type:varchar(18);not null"`
	Author uuid.UUID `json:"author" gorm:"column:author;type:char(36);not null"`
}

func IsGroup(target string) bool {
	for _, value := range AVAILABLE_GROUPS {
		if strings.EqualFold(value, target) {
			return true
		}
	}

	return false
}

func (metadataSet *MetadataSet) Write(target string, value interface{}) bool {
	switch target {
	case "skin":
		metadataSet.Skin = value.(string)
		return true
	case "name":
		metadataSet.Name = value.(string)
		return true
	case "vanish":
		metadataSet.Vanish, _ = value.(bool)
		return true
	case "see_all_players":
		metadataSet.SEE_ALL_PLAYERS, _ = value.(bool)
		return true
	case "enable_public_tell":
		metadataSet.ENABLE_PUBLIC_TELL, _ = value.(bool)
		return true
	case "staff_scoreboard":
		metadataSet.STAFF_SCOREBOARD, _ = value.(bool)
		return true
	case "current_group":
		metadataSet.CurrentGroup = value.(string)
		return true
	default:
		return false
	}
}

func ReadInfo(id uuid.UUID, group string, data map[string]string) GroupInfo {
	updatedAt, _ := strconv.ParseInt(data["createdAt"], 10, 64)
	expireAt, _ := strconv.ParseInt(data["expireAt"], 10, 64)

	return GroupInfo{
		ExpiredTimestamp: ExpiredTimestamp{
			CreatedAt: updatedAt,
			ExpireAt:  expireAt,
		},
		User:   id,
		Group:  group,
		Author: uuid.MustParse(data["author"]),
	}
}

func (metadataSet *MetadataSet) ReadFrom(target string, content string) {
	switch target {
	case "SKIN":
		metadataSet.Skin = content
	case "NAME":
		metadataSet.Name = content
	case "VANISH":
		metadataSet.Vanish, _ = strconv.ParseBool(content)
	case "SEE_ALL_PLAYERS":
		metadataSet.SEE_ALL_PLAYERS, _ = strconv.ParseBool(content)
	case "ENABLE_PUBLIC_TELL":
		metadataSet.ENABLE_PUBLIC_TELL, _ = strconv.ParseBool(content)
	case "STAFF_SCOREBOARD":
		metadataSet.STAFF_SCOREBOARD, _ = strconv.ParseBool(content)
	case "CURRENT_GROUP":
		metadataSet.CurrentGroup = content
	}
}

func (info *MetadataSet) Read(data map[string]string) {
	info.Skin = data["skin"]
	info.Name = data["name"]

	info.Vanish, _ = strconv.ParseBool(data["vanish"])

	info.SEE_ALL_PLAYERS, _ = strconv.ParseBool(data["see_all_players"])
	info.ENABLE_PUBLIC_TELL, _ = strconv.ParseBool(data["enable_public_tell"])
	info.STAFF_SCOREBOARD, _ = strconv.ParseBool(data["staff_scoreboard"])
	info.CurrentGroup = data["current_group"]
}
