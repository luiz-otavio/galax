package data

type AccountType string

const (
	PREMIUM AccountType = "PREMIUM"
	CRACKED AccountType = "CRACKED"
)

type GroupType string

const (
	OWNER     GroupType = "OWNER"
	ADMIN     GroupType = "ADMIN"
	MODERATOR GroupType = "MODERATOR"
	HELPER    GroupType = "HELPER"
	YOUTUBER  GroupType = "YOUTUBER"
	STREAMER  GroupType = "STREAMER"
	PATRON    GroupType = "PATRON"
	ELITE     GroupType = "ELITE"
	MVP       GroupType = "MVP"
	VIP       GroupType = "VIP"
	DEFAULT   GroupType = "DEFAULT"
	UNKNOWN   GroupType = "UNKNOWN"
)

type UUIDData struct {
	UUID string `gorm:"primaryKey;type:char(36)" json:"unique_id"`
}
