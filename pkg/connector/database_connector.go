package connector

import (
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

	return &DBConnector{Database: db}
}
