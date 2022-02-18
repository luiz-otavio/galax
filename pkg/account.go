package galax

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

func New(uniqueId uuid.UUID, name string) Account {
	now := time.Now().UnixMilli()

	return Account{
		UniqueId: uniqueId,
		Name:     name,

		MetadataSet: MetadataSet{
			User:             uniqueId,
			CurrentGroup:     "NORMAL",
			SeeAllPlayers:    true,
			EnablePublicTell: true,
		},

		Timestamp: Timestamp{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

type Timestamp struct {
	UpdatedAt int64 `json:"updated_at" gorm:"column:updated_at;type:bigint;not null;default:0"`
	CreatedAt int64 `json:"created_at" gorm:"column:created_at;type:bigint;not null;default:0"`
}

type ExpiredTimestamp struct {
	ExpireAt  int64 `json:"expire_at" gorm:"column:created_at;type:bigint;not null;default:0"`
	CreatedAt int64 `json:"created_at" gorm:"column:created_at;type:bigint;not null;default:0"`
}

type Account struct {
	UniqueId uuid.UUID `json:"uniqueId" gorm:"primary_key;type:char(36);not null"`

	Name string `json:"name" gorm:"type:varchar(16);not null;column:username"`
	Cash int32  `json:"cash" gorm:"type:bigint;not null;default:0"`

	GroupSet    []GroupInfo `json:"group_set" gorm:"foreingKey:User"`
	MetadataSet MetadataSet `json:"metadata_set" gorm:"foreingKey:User"`

	Timestamp Timestamp `gorm:"embedded"`
}

type MetadataSet struct {
	User uuid.UUID `json:"-" gorm:"column:user;type:char(36);not null"`

	Skin         string `json:"skin" gorm:"column:skin;type:varchar(16);not null"`
	Name         string `json:"name" gorm:"column:name;type:varchar(16);not null"`
	Vanish       bool   `json:"vanish" gorm:"column:vanish;type:boolean;not null"`
	Flying       bool   `json:"flying" gorm:"column:flying;type:boolean;not null"`
	CurrentGroup string `json:"current_group" gorm:"column:current_group;type:varchar(18);not null"`

	SeeAllPlayers    bool `json:"see_all_players" gorm:"column:see_all_players;type:boolean;not null"`
	EnablePublicTell bool `json:"enable_public_tell" gorm:"column:enable_public_tell;type:boolean;not null"`
	// StaffScoreboard  bool `json:"staff_scoreboard" gorm:"column:staff_scoreboard;type:boolean;not null"`
}

type GroupInfo struct {
	ExpiredTimestamp

	User uuid.UUID `json:"-" gorm:"column:user;type:char(36);not null"`

	Group  string    `json:"group" gorm:"column:role;type:varchar(18);not null"`
	Author uuid.UUID `json:"author" gorm:"column:author;type:char(36);not null"`
}

// func IsGroup(target string) bool {
// 	for _, value := range AVAILABLE_GROUPS {
// 		if strings.EqualFold(value, target) {
// 			return true
// 		}
// 	}

// 	return false
// }

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
		metadataSet.SeeAllPlayers, _ = value.(bool)
		return true
	case "enable_public_tell":
		metadataSet.EnablePublicTell, _ = value.(bool)
		return true
	case "current_group":
		metadataSet.CurrentGroup = value.(string)
		return true
	default:
		return false
	}
}

func ReadInfo(id uuid.UUID, group string, data map[string]string) GroupInfo {
	createdAt, _ := strconv.ParseInt(data["createdAt"], 10, 64)
	expireAt, _ := strconv.ParseInt(data["expireAt"], 10, 64)

	return GroupInfo{
		ExpiredTimestamp: ExpiredTimestamp{
			CreatedAt: int64(createdAt),
			ExpireAt:  int64(expireAt),
		},

		User:  id,
		Group: group,
	}
}

func (info *MetadataSet) Read(data map[string]string) {
	info.Skin = data["skin"]
	info.Name = data["name"]
	info.CurrentGroup = data["current_group"]

	info.Vanish, _ = strconv.ParseBool(data["vanish"])
	info.Flying, _ = strconv.ParseBool(data["flying"])

	info.SeeAllPlayers, _ = strconv.ParseBool(data["see_all_players"])
	info.EnablePublicTell, _ = strconv.ParseBool(data["enable_public_tell"])
}
