package impl

import (
	"time"

	. "github.com/luiz-otavio/galax/pkg/data"
)

type AccountImpl struct {
	UUIDData

	AccountType AccountType `json:"accountType" gorm:"column:account_type;type:varchar(16);not null"`

	Name string `json:"name" gorm:"type:varchar(16);not null;column:username"`
	Cash int32  `json:"cash" gorm:"type:bigint;not null;default:0"`

	GroupSet    []GroupInfo `json:"group_set" gorm:"foreignkey:User;references:UUID"`
	MetadataSet MetadataSet `json:"metadata_set" gorm:"foreignkey:User;references:UUID"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (account AccountImpl) GetUniqueId() string {
	return account.GetUniqueId()
}

func (account AccountImpl) GetCash() int32 {
	return account.Cash
}

func (account AccountImpl) AddCash(amount int32) {
	account.Cash += amount
}

func (account AccountImpl) SetCash(cash int32) {
	account.Cash = cash
}

func (account AccountImpl) TakeCash(amount int32) {
	account.Cash -= amount
}

func (account AccountImpl) GetMetadataSet() MetadataSet {
	return account.MetadataSet
}

func (account AccountImpl) GetAccountType() AccountType {
	return account.AccountType
}

func (account AccountImpl) GetGroupSet() []GroupInfo {
	return account.GroupSet
}

func (account AccountImpl) AddGroup(group GroupInfo) {
	account.GroupSet = append(account.GroupSet, group)
}

func (account AccountImpl) RemoveGroup(group GroupType) {
	for i, g := range account.GroupSet {
		if g.Group == group {
			account.GroupSet = append(account.GroupSet[:i], account.GroupSet[i+1:]...)
		}
	}
}

func (account AccountImpl) HasGroupSet(group GroupType) bool {
	for _, g := range account.GroupSet {
		if g.Group == group {
			return true
		}
	}

	return false
}

func (account AccountImpl) GetCreatedAt() time.Time {
	return account.CreatedAt
}

func (account AccountImpl) GetUpdatedAt() time.Time {
	return account.UpdatedAt
}

func (account AccountImpl) GetName() string {
	return account.Name
}

func CreateMetadata(unique, skin, name, currentGroup string, vanish, flying, seeAllPlayers, publicTell, seeAllReports, staffChat bool) MetadataSet {
	return MetadataSet{
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
	return AccountImpl{
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
