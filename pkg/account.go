package galax

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

func New(uniqueId uuid.UUID, name string) Account {
	now := time.Now().
		In(time.UTC)

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
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type ExpiredTimestamp struct {
	ExpireAt  time.Time `json:"expire_at" gorm:"column:expire_at;not null;default:CURRENT_TIMESTAMP();"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP();"`
}

type Account struct {
	UniqueId uuid.UUID `json:"uniqueId" gorm:"primary_key;type:char(36);not null"`

	Name string `json:"name" gorm:"type:varchar(16);not null;column:username"`
	Cash int32  `json:"cash" gorm:"type:bigint;not null;default:0"`

	GroupSet    []GroupInfo `json:"group_set" gorm:"foreignkey:User;references:UniqueId"`
	MetadataSet MetadataSet `json:"metadata_set" gorm:"foreignkey:User;references:UniqueId"`

	Timestamp
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
	StaffChat        bool `json:"staff_chat" gorm:"column:staff_chat;type:boolean;not null"`
	// StaffScoreboard  bool `json:"staff_scoreboard" gorm:"column:staff_scoreboard;type:boolean;not null"`
}

type GroupInfo struct {
	ExpiredTimestamp

	User uuid.UUID `json:"-" gorm:"column:user;type:char(36);not null"`

	Group  string    `json:"group" gorm:"column:role;type:varchar(18);not null"`
	Author uuid.UUID `json:"author" gorm:"column:author;type:char(36);not null"`
}

func ParseType(key string, value interface{}) interface{} {
	switch key {
	case "skin":
		return value.(string)
	case "name":
		return value.(string)
	case "vanish":
		target, _ := strconv.ParseBool(value.(string))

		return target
	case "flying":
		target, _ := strconv.ParseBool(value.(string))

		return target
	case "current_group":
		return value.(string)
	case "see_all_players":
		target, _ := strconv.ParseBool(value.(string))

		return target
	case "enable_public_tell":
		target, _ := strconv.ParseBool(value.(string))

		return target
	case "staff_chat":
		target, _ := strconv.ParseBool(value.(string))

		return target
	}

	return value
}

func ReadInfo(id uuid.UUID, group string, data map[string]string) GroupInfo {
	createdAt, _ := strconv.ParseInt(data["createdAt"], 10, 64)
	expireAt, _ := strconv.ParseInt(data["expireAt"], 10, 64)

	return GroupInfo{
		ExpiredTimestamp: ExpiredTimestamp{
			CreatedAt: time.Unix(int64(createdAt), 0).In(time.UTC),
			ExpireAt:  time.Unix(int64(expireAt), 0).In(time.UTC),
		},

		User: id,

		Author: uuid.MustParse(data["author"]),

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
	info.StaffChat, _ = strconv.ParseBool(data["staff_chat"])
}
