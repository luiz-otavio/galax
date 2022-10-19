package cmd

import (
	"github.com/Rede-Legit/galax/pkg/data"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	interfaces := []interface{}{
		data.CreateEmptyAccount(),
		data.CreateEmptyAuthentication(),
		data.CreateEmptyGroupInfo(),
		data.CreateEmptyMetadataSet(),
	}

	if err := db.AutoMigrate(interfaces...); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database.")
		return err
	}

	log.Info().Msg("Database migrated successfully.")
	return nil
}
