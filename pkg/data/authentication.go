package data

import (
	"time"
)

type Authentication interface {
	GetUniqueId() string

	GetUsername() string
	GetPassword() string

	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time

	CheckPassword(password string) bool
	UpdatePassword(password string)
}

type LoginRequest interface {
	GetUsername() string
	GetPassword() string
}
