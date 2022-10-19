package data

import (
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
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

type authenticationImpl struct {
	UUIDData

	Username string `json:"username" gorm:"type:varchar(16);not null;column:username"`
	Password string `json:"password" gorm:"type:char(60);not null;column:password"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (authentication *authenticationImpl) GetUniqueId() string {
	return authentication.UUID
}

func (authentication *authenticationImpl) GetUsername() string {
	return authentication.Username
}

func (authentication *authenticationImpl) GetPassword() string {
	return authentication.Password
}

func (authentication *authenticationImpl) GetCreatedAt() time.Time {
	return authentication.CreatedAt
}

func (authentication *authenticationImpl) GetUpdatedAt() time.Time {
	return authentication.UpdatedAt
}

func (authentication *authenticationImpl) CheckPassword(password string) bool {
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt password")
		return false
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(authentication.Password),
		encrypted,
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to compare password")
		return false
	}

	return err == nil
}

func (authentication *authenticationImpl) UpdatePassword(password string) {
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt password")
		return
	}

	authentication.Password = string(encrypted)
}

type loginRequestImpl struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (loginRequest *loginRequestImpl) GetUsername() string {
	return loginRequest.Username
}

func (loginRequest *loginRequestImpl) GetPassword() string {
	return loginRequest.Password
}

func CreateAuthentication(username string, password string) Authentication {
	authentication := &authenticationImpl{
		Username: username,
	}

	authentication.UpdatePassword(password)

	return authentication
}

func CreateEmptyAuthentication() Authentication {
	return &authenticationImpl{}
}

func CreateLoginRequest(username string, password string) LoginRequest {
	return &loginRequestImpl{
		Username: username,
		Password: password,
	}
}

func CreateEmptyLoginRequest() LoginRequest {
	return &loginRequestImpl{}
}
