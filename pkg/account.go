package galax

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func New(uniqueId uuid.UUID, name string) *Account {
	return &Account{
		UniqueId: uniqueId,
		Name:     name,

		MetadataSet: MetadataSet{
			Skin:               "",
			Name:               "",
			CurrentGroup:       "NORMAL",
			Vanish:             false,
			SEE_ALL_PLAYERS:    true,
			ENABLE_PUBLIC_TELL: true,
			STAFF_SCOREBOARD:   false,
		},
		GroupSet: make(map[string]GroupInfo),
		Cash:     0,
	}
}

type Timestamp struct {
	UpdatedAt int64 `json:"updated_at"`
	CreatedAt int64 `json:"created_at"`
}

type ExpiredTimestamp struct {
	ExpireAt  int64 `json:"expire_at"`
	CreatedAt int64 `json:"created_at"`
}

type Account struct {
	UniqueId uuid.UUID `json:"uniqueId"`

	Name string `json:"name"`
	Cash int32  `json:"cash"`

	GroupSet    GroupSet    `json:"group_set"`
	MetadataSet MetadataSet `json:"metadata_set"`
}

var AVAILABLE_GROUPS = []string{
	"DIRECTOR",
	"SUB-DIRECTOR",
	"ADMIN",
	"MODERATOR",
	"HELPER",
	"DESIGNER",
	"BUILDER",
	"INFLUENCER+",
	"INFLUENCER",
	"MVP++",
	"MVP+",
	"MVP",
	"VIP",
	"NORMAL",
}

type MetadataSet struct {
	Skin         string `json:"skin"`
	Name         string `json:"name"`
	Vanish       bool   `json:"vanish"`
	CurrentGroup string `json:"currentGroup"`

	SEE_ALL_PLAYERS    bool `json:"seeAllPlayers"`
	ENABLE_PUBLIC_TELL bool `json:"enablePublicTell"`
	STAFF_SCOREBOARD   bool `json:"staffScoreboard"`
}

type GroupSet = map[string]GroupInfo

type GroupInfo struct {
	ExpiredTimestamp

	Author string `json:"author"`
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
	case "SKIN":
		metadataSet.Skin = value.(string)
		return true
	case "NAME":
		metadataSet.Name = value.(string)
		return true
	case "VANISH":
		metadataSet.Vanish, _ = value.(bool)
		return true
	case "SEE_ALL_PLAYERS":
		metadataSet.SEE_ALL_PLAYERS, _ = value.(bool)
		return true
	case "ENABLE_PUBLIC_TELL":
		metadataSet.ENABLE_PUBLIC_TELL, _ = value.(bool)
		return true
	case "STAFF_SCOREBOARD":
		metadataSet.STAFF_SCOREBOARD, _ = value.(bool)
		return true
	default:
		return false
	}
}

func Read(data map[string]string) GroupInfo {
	updatedAt, _ := strconv.ParseInt(data["createdAt"], 10, 64)
	expireAt, _ := strconv.ParseInt(data["expireAt"], 10, 64)

	return GroupInfo{
		ExpiredTimestamp: ExpiredTimestamp{
			CreatedAt: updatedAt,
			ExpireAt:  expireAt,
		},
		Author: data["author"],
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

	info.SEE_ALL_PLAYERS, _ = strconv.ParseBool(data["seeAllPlayers"])
	info.ENABLE_PUBLIC_TELL, _ = strconv.ParseBool(data["enablePublicTell"])
	info.STAFF_SCOREBOARD, _ = strconv.ParseBool(data["staffScoreboard"])
	info.CurrentGroup = data["currentGroup"]
}
