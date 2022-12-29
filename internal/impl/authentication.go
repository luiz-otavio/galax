package impl

import (
	"time"

	. "github.com/luiz-otavio/galax/pkg/data"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AuthenticationImpl struct {
	UUIDData

	Username string `json:"username" gorm:"type:varchar(16);not null;column:username"`
	Password string `json:"password" gorm:"type:char(60);not null;column:password"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (authentication AuthenticationImpl) GetUniqueId() string {
	return authentication.GetUniqueId()
}

func (authentication AuthenticationImpl) GetUsername() string {
	return authentication.Username
}

func (authentication AuthenticationImpl) GetPassword() string {
	return authentication.Password
}

func (authentication AuthenticationImpl) GetCreatedAt() time.Time {
	return authentication.CreatedAt
}

func (authentication AuthenticationImpl) GetUpdatedAt() time.Time {
	return authentication.UpdatedAt
}

func (authentication AuthenticationImpl) CheckPassword(password string) bool {
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

func (authentication AuthenticationImpl) UpdatePassword(password string) {
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt password")
		return
	}

	authentication.Password = string(encrypted)
}

type LoginImpl struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (loginRequest LoginImpl) GetUsername() string {
	return loginRequest.Username
}

func (loginRequest LoginImpl) GetPassword() string {
	return loginRequest.Password
}

func CreateAuthentication(username string, password string) Authentication {
	authentication := &AuthenticationImpl{
		Username: username,
	}

	authentication.UpdatePassword(password)

	return authentication
}

func CreateEmptyAuthentication() Authentication {
	return &AuthenticationImpl{}
}

func CreateLoginRequest(username string, password string) LoginRequest {
	return &LoginImpl{
		Username: username,
		Password: password,
	}
}

func CreateEmptyLoginRequest() LoginRequest {
	return &LoginImpl{}
}
