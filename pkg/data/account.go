package data

import "time"

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

type accountImpl struct {
	UUIDData

	AccountType AccountType `json:"accountType" gorm:"column:account_type;type:varchar(16);not null"`

	Name string `json:"name" gorm:"type:varchar(16);not null;column:username"`
	Cash int32  `json:"cash" gorm:"type:bigint;not null;default:0"`

	GroupSet    []GroupInfo `json:"group_set" gorm:"foreignkey:User;references:UUID"`
	MetadataSet MetadataSet `json:"metadata_set" gorm:"foreignkey:User;references:UUID"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (account *accountImpl) GetUniqueId() string {
	return account.UUID
}

func (account *accountImpl) GetCash() int32 {
	return account.Cash
}

func (account *accountImpl) AddCash(amount int32) {
	account.Cash += amount
}

func (account *accountImpl) SetCash(cash int32) {
	account.Cash = cash
}

func (account *accountImpl) TakeCash(amount int32) {
	account.Cash -= amount
}

func (account *accountImpl) GetMetadataSet() MetadataSet {
	return account.MetadataSet
}

func (account *accountImpl) GetAccountType() AccountType {
	return account.AccountType
}

func (account *accountImpl) GetGroupSet() []GroupInfo {
	return account.GroupSet
}

func (account *accountImpl) AddGroup(group GroupInfo) {
	account.GroupSet = append(account.GroupSet, group)
}

func (account *accountImpl) RemoveGroup(group GroupType) {
	for i, g := range account.GroupSet {
		if g.Group == group {
			account.GroupSet = append(account.GroupSet[:i], account.GroupSet[i+1:]...)
		}
	}
}

func (account *accountImpl) HasGroupSet(group GroupType) bool {
	for _, g := range account.GroupSet {
		if g.Group == group {
			return true
		}
	}

	return false
}

func (account *accountImpl) GetCreatedAt() time.Time {
	return account.CreatedAt
}

func (account *accountImpl) GetUpdatedAt() time.Time {
	return account.UpdatedAt
}

func (account *accountImpl) GetName() string {
	return account.Name
}

func CreateMetadata(unique, skin, name, currentGroup string, vanish, flying, seeAllPlayers, publicTell, seeAllReports, staffChat bool) *MetadataSet {
	return &MetadataSet{
		User: unique,
		Skin: skin,
		Name: name,

		Vanish:           vanish,
		Flying:           flying,
		SeeAllPlayers:    seeAllPlayers,
		EnablePublicTell: publicTell,
		SeeAllReports:    staffChat,
		SeeAllStaffChat:  seeAllReports,
	}
}

func CreateGroupInfo(unique, author string, group GroupType, expireAt, createdAt time.Time) GroupInfo {
	return GroupInfo{
		User:  unique,
		Group: group,

		Author: author,

		ExpireAt:  expireAt,
		CreatedAt: createdAt,
	}
}

func CreateAccount(unique, name string, cash int32, accountType AccountType, metadataSet MetadataSet, groupInfos []GroupInfo, createdAt, updatedAt time.Time) Account {
	return &accountImpl{
		UUIDData: UUIDData{
			UUID: unique,
		},

		AccountType: accountType,

		Name: name,
		Cash: cash,

		MetadataSet: metadataSet,
		GroupSet:    groupInfos,

		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func CreateEmptyMetadataSet() MetadataSet {
	return MetadataSet{}
}

func CreateEmptyGroupInfo() GroupInfo {
	return GroupInfo{}
}

func CreateEmptyAccount() Account {
	return &accountImpl{}
}
