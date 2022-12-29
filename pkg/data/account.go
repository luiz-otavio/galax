package data

import (
	"time"
)

type GroupInfo struct {
	User string `json:"-" gorm:"column:user;type:char(36);not null"`

	Group  GroupType `json:"group" gorm:"column:role;type:varchar(18);not null"`
	Author string    `json:"author" gorm:"column:author;type:char(36);not null"`

	ExpireAt  time.Time `json:"expire_at" gorm:"column:expire_at;not null;default:CURRENT_TIMESTAMP();"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP();"`
}

type MetadataSet struct {
	User string `json:"-" gorm:"column:user;type:char(36);not null"`

	Skin         string    `json:"skin" gorm:"column:skin;type:varchar(16);not null"`
	Name         string    `json:"name" gorm:"column:name;type:varchar(16);not null"`
	Vanish       bool      `json:"vanish" gorm:"column:vanish;type:boolean;not null"`
	Flying       bool      `json:"flying" gorm:"column:flying;type:boolean;not null"`
	CurrentGroup GroupType `json:"current_group" gorm:"column:current_group;type:varchar(18);not null"`

	SeeAllReports    bool `json:"see_all_reports" gorm:"column:see_all_reports;type:boolean;not null"`
	SeeAllStaffChat  bool `json:"see_all_staff_chat" gorm:"column:see_all_staff_chat;type:boolean;not null"`
	SeeAllPlayers    bool `json:"see_all_players" gorm:"column:see_all_players;type:boolean;not null"`
	EnablePublicTell bool `json:"enable_public_tell" gorm:"column:enable_public_tell;type:boolean;not null"`
}

type Account interface {
	GetUniqueId() string
	GetName() string

	GetAccountType() AccountType

	GetCash() int32
	SetCash(cash int32)
	AddCash(amount int32)
	TakeCash(amount int32)

	GetMetadataSet() MetadataSet
	GetGroupSet() []GroupInfo

	AddGroup(group GroupInfo)
	RemoveGroup(group GroupType)
	HasGroupSet(group GroupType) bool

	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}
