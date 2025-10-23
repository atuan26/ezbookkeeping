package cli

import (
	"encoding/json"
	"fmt"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// MigrationData represents migration data cli
type MigrationData struct {
	CliUsingConfig
}

// Initialize a migration data cli singleton instance
var (
	Migration = &MigrationData{
		CliUsingConfig: CliUsingConfig{
			container: settings.Container,
		},
	}
)

// MigrateToFunds migrates all existing users to the multi-fund system
func (s *MigrationData) MigrateToFunds(c core.Context) error {
	log.CliInfof(c, "[migration.MigrateToFunds] starting migration to multi-fund system")

	status, err := services.Migration.MigrateToFunds(c)
	if err != nil {
		return err
	}

	// Print migration status
	statusJson, _ := json.MarshalIndent(status, "", "  ")
	fmt.Printf("Migration Status:\n%s\n", string(statusJson))

	if status.FailedUsers > 0 {
		log.CliWarnf(c, "[migration.MigrateToFunds] migration completed with %d failures", status.FailedUsers)
		if status.LastError != "" {
			log.CliWarnf(c, "[migration.MigrateToFunds] last error: %s", status.LastError)
		}
	} else {
		log.CliInfof(c, "[migration.MigrateToFunds] migration completed successfully")
	}

	return nil
}

// ValidateMigration validates the migration results
func (s *MigrationData) ValidateMigration(c core.Context) error {
	log.CliInfof(c, "[migration.ValidateMigration] starting migration validation")

	result, err := services.Migration.ValidateMigration(c)
	if err != nil {
		log.CliErrorf(c, "[migration.ValidateMigration] validation failed: %s", err.Error())
		return err
	}

	// Print validation result
	resultJson, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Validation Result:\n%s\n", string(resultJson))

	if !result.IsValid {
		log.CliErrorf(c, "[migration.ValidateMigration] migration validation failed: %d users failed validation", result.FailedUsers)
		for _, errorMsg := range result.ValidationErrors {
			log.CliErrorf(c, "[migration.ValidateMigration] %s", errorMsg)
		}
		return fmt.Errorf("migration validation failed")
	}

	log.CliInfof(c, "[migration.ValidateMigration] migration validation completed successfully")
	return nil
}

// CheckMigrationStatus checks if migration has already been performed
func (s *MigrationData) CheckMigrationStatus(c core.Context) error {
	log.CliInfof(c, "[migration.CheckMigrationStatus] checking migration status")

	migrated, err := services.Migration.CheckMigrationStatus(c)
	if err != nil {
		log.CliErrorf(c, "[migration.CheckMigrationStatus] failed to check migration status: %s", err.Error())
		return err
	}

	if migrated {
		log.CliInfof(c, "[migration.CheckMigrationStatus] migration has already been performed")
		fmt.Printf("Migration Status: COMPLETED\n")
	} else {
		log.CliInfof(c, "[migration.CheckMigrationStatus] migration has not been performed yet")
		fmt.Printf("Migration Status: NOT PERFORMED\n")
	}

	return nil
}

// RollbackMigration rolls back the migration (for testing purposes)
func (s *MigrationData) RollbackMigration(c core.Context) error {
	log.CliInfof(c, "[migration.RollbackMigration] starting migration rollback")

	status, err := services.Migration.RollbackMigration(c)
	if err != nil {
		log.CliErrorf(c, "[migration.RollbackMigration] rollback failed: %s", err.Error())
		return err
	}

	// Print rollback status
	statusJson, _ := json.MarshalIndent(status, "", "  ")
	fmt.Printf("Rollback Status:\n%s\n", string(statusJson))

	if status.FailedUsers > 0 {
		log.CliWarnf(c, "[migration.RollbackMigration] rollback completed with %d failures", status.FailedUsers)
		if status.LastError != "" {
			log.CliWarnf(c, "[migration.RollbackMigration] last error: %s", status.LastError)
		}
	} else {
		log.CliInfof(c, "[migration.RollbackMigration] migration rollback completed successfully")
	}

	return nil
}
