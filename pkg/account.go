package galax

import (
	"time"

	"github.com/google/uuid"
)

func New(uniqueId uuid.UUID, name string) Account {
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
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
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

func (account *Account) GetUniqueId() string {
	return account.UniqueId.String()
}

func (account *Account) HasGroupSet(group string) bool {
	for _, g := range account.GroupSet {
		if g.Group == group {
			return true
		}
	}

	return false
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
}

type GroupInfo struct {
	ExpiredTimestamp

	User uuid.UUID `json:"-" gorm:"column:user;type:char(36);not null"`

	Group  string    `json:"group" gorm:"column:role;type:varchar(18);not null"`
	Author uuid.UUID `json:"author" gorm:"column:author;type:char(36);not null"`
}
