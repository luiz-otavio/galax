package connector

import (
	"time"

	"github.com/Rede-Legit/galax/pkg/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBConnector struct {
	Database *gorm.DB
}

func NewConnector(dsn string) *DBConnector {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{})

	if err != nil {
		util.Log(err)
	}

	back, _ := db.DB()

	back.SetConnMaxLifetime(time.Duration(5) * time.Minute)
	back.SetConnMaxIdleTime(time.Duration(2) * time.Minute)

	return &DBConnector{Database: db}
}
