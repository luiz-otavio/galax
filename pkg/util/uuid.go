package util

import (
	"crypto/md5"

	"github.com/google/uuid"
)

func OfflinePlayerUUID(username string) uuid.UUID {
	const version = 3 // UUID v3
	uuid := md5.Sum([]byte("OfflinePlayer:" + username))
	uuid[6] = (uuid[6] & 0x0f) | uint8((version&0xf)<<4)
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // RFC 4122 variant
	return uuid
}
