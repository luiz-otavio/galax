package cmd

import (
	"github.com/luiz-otavio/galax/internal/impl"
	"github.com/luiz-otavio/galax/pkg/data"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	interfaces := []interface{}{
		impl.AccountImpl{},
		impl.AuthenticationImpl{},
		data.GroupInfo{},
		data.MetadataSet{},
	}

	if err := db.AutoMigrate(interfaces...); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database.")
		return err
	}

	log.Info().Msg("Database migrated successfully.")
	return nil
}
