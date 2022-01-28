package galax

import (
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
		panic(err)
	}

	database, err := db.DB()

	if err != nil {
		panic(err)
	}

	if database.Ping() != nil {
		panic("Could not connect to database")
	}

	return &DBConnector{Database: db}
}
