package data

type DataType int

const (
	LOBBY DataType = iota
	AUTHENTICATION
	PRISON
	SKYWARS
	BEDWARS
)

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

type GalaxData struct {
	Type DataType
}

type UUIDData struct {
	GalaxData

	UUID string `gorm:"primaryKey;type:char(36)" json:"unique_id"`
}
