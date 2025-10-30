package services

import (
	"fmt"
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// MigrationService represents migration service
type MigrationService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a migration service singleton instance
var (
	Migration = &MigrationService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// MigrationStatus represents the status of migration
type MigrationStatus struct {
	TotalUsers      int64    `json:"totalUsers"`
	ProcessedUsers  int64    `json:"processedUsers"`
	SuccessfulUsers int64    `json:"successfulUsers"`
	FailedUsers     int64    `json:"failedUsers"`
	IsCompleted     bool     `json:"isCompleted"`
	LastError       string   `json:"lastError,omitempty"`
	FailedUsernames []string `json:"failedUsernames,omitempty"`
	StartTime       int64    `json:"startTime"`
	EndTime         int64    `json:"endTime,omitempty"`
}

// ValidationResult represents the result of migration validation
type ValidationResult struct {
	TotalUsers       int64    `json:"totalUsers"`
	ValidatedUsers   int64    `json:"validatedUsers"`
	SuccessfulUsers  int64    `json:"successfulUsers"`
	FailedUsers      int64    `json:"failedUsers"`
	IsValid          bool     `json:"isValid"`
	ValidationErrors []string `json:"validationErrors,omitempty"`
	FailedUsernames  []string `json:"failedUsernames,omitempty"`
}

// MigrateToFunds migrates existing users to the multi-fund system
func (s *MigrationService) MigrateToFunds(c core.Context) (*MigrationStatus, error) {
	log.BootInfof(c, "[migration.MigrateToFunds] starting migration to multi-fund system")

	// Get all users
	users, err := s.getAllUsers(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	status := &MigrationStatus{
		TotalUsers:      int64(len(users)),
		ProcessedUsers:  0,
		SuccessfulUsers: 0,
		FailedUsers:     0,
		IsCompleted:     false,
		FailedUsernames: make([]string, 0),
		StartTime:       time.Now().Unix(),
	}

	log.BootInfof(c, "[migration.MigrateToFunds] found %d users to migrate", len(users))

	// Process each user
	for _, user := range users {
		status.ProcessedUsers++

		log.BootInfof(c, "[migration.MigrateToFunds] migrating user %s (uid: %d) [%d/%d]",
			user.Username, user.Uid, status.ProcessedUsers, status.TotalUsers)

		err := s.migrateUserToFunds(c, user)
		if err != nil {
			status.FailedUsers++
			status.LastError = fmt.Sprintf("failed to migrate user %s: %v", user.Username, err)
			status.FailedUsernames = append(status.FailedUsernames, user.Username)
			log.BootErrorf(c, "[migration.MigrateToFunds] %s", status.LastError)
			continue
		}

		status.SuccessfulUsers++
		log.BootInfof(c, "[migration.MigrateToFunds] successfully migrated user %s", user.Username)
	}

	status.IsCompleted = true
	status.EndTime = time.Now().Unix()

	log.BootInfof(c, "[migration.MigrateToFunds] migration completed: %d successful, %d failed out of %d total users",
		status.SuccessfulUsers, status.FailedUsers, status.TotalUsers)

	return status, nil
}

// ValidateMigration validates the migration results
func (s *MigrationService) ValidateMigration(c core.Context) (*ValidationResult, error) {
	log.BootInfof(c, "[migration.ValidateMigration] starting migration validation")

	// Get all users
	users, err := s.getAllUsers(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for validation: %w", err)
	}

	result := &ValidationResult{
		TotalUsers:       int64(len(users)),
		ValidatedUsers:   0,
		SuccessfulUsers:  0,
		FailedUsers:      0,
		IsValid:          true,
		ValidationErrors: make([]string, 0),
		FailedUsernames:  make([]string, 0),
	}

	log.BootInfof(c, "[migration.ValidateMigration] validating %d users", len(users))

	for _, user := range users {
		result.ValidatedUsers++

		err := s.validateUserMigration(c, user)
		if err != nil {
			result.FailedUsers++
			result.IsValid = false
			errorMsg := fmt.Sprintf("validation failed for user %s: %v", user.Username, err)
			result.ValidationErrors = append(result.ValidationErrors, errorMsg)
			result.FailedUsernames = append(result.FailedUsernames, user.Username)
			log.BootErrorf(c, "[migration.ValidateMigration] %s", errorMsg)
			continue
		}

		result.SuccessfulUsers++
		log.BootInfof(c, "[migration.ValidateMigration] user %s validation passed", user.Username)
	}

	if result.IsValid {
		log.BootInfof(c, "[migration.ValidateMigration] migration validation completed successfully")
	} else {
		log.BootErrorf(c, "[migration.ValidateMigration] migration validation failed: %d users failed validation", result.FailedUsers)
	}

	return result, nil
}

// RollbackMigration rolls back the migration (for testing purposes)
func (s *MigrationService) RollbackMigration(c core.Context) (*MigrationStatus, error) {
	log.BootInfof(c, "[migration.RollbackMigration] starting migration rollback")

	// Get all users
	users, err := s.getAllUsers(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for rollback: %w", err)
	}

	status := &MigrationStatus{
		TotalUsers:      int64(len(users)),
		ProcessedUsers:  0,
		SuccessfulUsers: 0,
		FailedUsers:     0,
		IsCompleted:     false,
		FailedUsernames: make([]string, 0),
		StartTime:       time.Now().Unix(),
	}

	log.BootInfof(c, "[migration.RollbackMigration] rolling back %d users", len(users))

	for _, user := range users {
		status.ProcessedUsers++

		log.BootInfof(c, "[migration.RollbackMigration] rolling back user %s (uid: %d) [%d/%d]",
			user.Username, user.Uid, status.ProcessedUsers, status.TotalUsers)

		err := s.rollbackUserMigration(c, user)
		if err != nil {
			status.FailedUsers++
			status.LastError = fmt.Sprintf("rollback failed for user %s: %v", user.Username, err)
			status.FailedUsernames = append(status.FailedUsernames, user.Username)
			log.BootErrorf(c, "[migration.RollbackMigration] %s", status.LastError)
			continue
		}

		status.SuccessfulUsers++
		log.BootInfof(c, "[migration.RollbackMigration] successfully rolled back user %s", user.Username)
	}

	status.IsCompleted = true
	status.EndTime = time.Now().Unix()

	log.BootInfof(c, "[migration.RollbackMigration] migration rollback completed: %d successful, %d failed out of %d total users",
		status.SuccessfulUsers, status.FailedUsers, status.TotalUsers)

	return status, nil
}

// CheckMigrationStatus checks if migration has already been performed
func (s *MigrationService) CheckMigrationStatus(c core.Context) (bool, error) {
	// Check if any user has funds - if so, migration has been performed
	users, err := s.getAllUsers(c)
	if err != nil {
		return false, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return false, nil // No users, no migration needed
	}

	// Check first user to see if they have funds
	firstUser := users[0]
	fundCount, err := s.UserDataDB(firstUser.Uid).NewSession(c).Where("owner_uid=? AND deleted=?", firstUser.Uid, false).Count(&models.Fund{})
	if err != nil {
		return false, fmt.Errorf("failed to check funds for user %s: %w", firstUser.Username, err)
	}

	return fundCount > 0, nil
}

// getAllUsers returns all non-deleted users
func (s *MigrationService) getAllUsers(c core.Context) ([]*models.User, error) {
	var users []*models.User
	err := s.UserDB().NewSession(c).Where("deleted=?", false).Find(&users)
	return users, err
}

// migrateUserToFunds migrates a single user to the multi-fund system
func (s *MigrationService) migrateUserToFunds(c core.Context, user *models.User) error {
	return s.UserDataDB(user.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		// 1. Check if user already has funds (migration already done)
		fundCount, err := sess.Where("owner_uid=? AND deleted=?", user.Uid, false).Count(&models.Fund{})
		if err != nil {
			return fmt.Errorf("failed to check existing funds: %w", err)
		}

		if fundCount > 0 {
			log.BootInfof(c, "[migration.migrateUserToFunds] user %s already has funds, skipping", user.Username)
			return nil
		}

		// 2. Create personal fund
		fundId := s.GenerateUuid(uuid.UUID_TYPE_FUND)
		if fundId < 1 {
			return errs.ErrSystemIsBusy
		}

		now := time.Now().Unix()
		personalFund := &models.Fund{
			FundId:          fundId,
			Name:            "Personal",
			OwnerUid:        user.Uid,
			DefaultCurrency: user.DefaultCurrency,
			Deleted:         false,
			CreatedUnixTime: now,
			UpdatedUnixTime: now,
		}

		_, err = sess.Insert(personalFund)
		if err != nil {
			return fmt.Errorf("failed to create personal fund: %w", err)
		}

		// 3. Create fund member for owner
		memberId := s.GenerateUuid(uuid.UUID_TYPE_FUND_MEMBER)
		if memberId < 1 {
			return errs.ErrSystemIsBusy
		}

		ownerMember := &models.FundMember{
			MemberId:        memberId,
			FundId:          fundId,
			Name:            user.Nickname,
			Email:           user.Email,
			Role:            models.FUND_ROLE_OWNER,
			LinkedUid:       user.Uid,
			CreatedBy:       user.Uid,
			CreatedUnixTime: now,
			UpdatedUnixTime: now,
		}

		_, err = sess.Insert(ownerMember)
		if err != nil {
			return fmt.Errorf("failed to create fund member: %w", err)
		}

		// 4. Update all financial data with fund_id
		err = s.updateFinancialDataWithFundId(c, sess, user.Uid, fundId)
		if err != nil {
			return fmt.Errorf("failed to update financial data: %w", err)
		}

		return nil
	})
}

// updateFinancialDataWithFundId updates all financial data to include the fund_id
func (s *MigrationService) updateFinancialDataWithFundId(c core.Context, sess *xorm.Session, uid int64, fundId int64) error {
	// Update accounts
	_, err := sess.Where("uid=?", uid).Update(&models.Account{FundId: fundId})
	if err != nil {
		return fmt.Errorf("failed to update accounts: %w", err)
	}

	// Update transactions
	_, err = sess.Where("uid=?", uid).Update(&models.Transaction{FundId: fundId})
	if err != nil {
		return fmt.Errorf("failed to update transactions: %w", err)
	}

	// Update transaction categories
	_, err = sess.Where("uid=?", uid).Update(&models.TransactionCategory{FundId: fundId})
	if err != nil {
		return fmt.Errorf("failed to update transaction categories: %w", err)
	}

	// Update transaction tags
	_, err = sess.Where("uid=?", uid).Update(&models.TransactionTag{FundId: fundId})
	if err != nil {
		return fmt.Errorf("failed to update transaction tags: %w", err)
	}

	// Update transaction templates
	_, err = sess.Where("uid=?", uid).Update(&models.TransactionTemplate{FundId: fundId})
	if err != nil {
		return fmt.Errorf("failed to update transaction templates: %w", err)
	}

	// Update transaction picture info
	_, err = sess.Where("uid=?", uid).Update(&models.TransactionPictureInfo{FundId: fundId})
	if err != nil {
		return fmt.Errorf("failed to update transaction picture info: %w", err)
	}

	return nil
}

// validateUserMigration validates that a user's migration was successful
func (s *MigrationService) validateUserMigration(c core.Context, user *models.User) error {
	sess := s.UserDataDB(user.Uid).NewSession(c)

	// 1. Check that user has exactly one personal fund
	fundCount, err := sess.Where("owner_uid=? AND deleted=?", user.Uid, false).Count(&models.Fund{})
	if err != nil {
		return fmt.Errorf("failed to count funds: %w", err)
	}
	if fundCount != 1 {
		return fmt.Errorf("expected 1 fund, found %d", fundCount)
	}

	// Get the fund
	fund := &models.Fund{}
	has, err := sess.Where("owner_uid=? AND deleted=?", user.Uid, false).Get(fund)
	if err != nil {
		return fmt.Errorf("failed to get fund: %w", err)
	}
	if !has {
		return fmt.Errorf("fund not found")
	}

	// 2. Check that user has exactly one fund member (owner)
	memberCount, err := sess.Where("fund_id=? AND linked_uid=?", fund.FundId, user.Uid).Count(&models.FundMember{})
	if err != nil {
		return fmt.Errorf("failed to count fund members: %w", err)
	}
	if memberCount != 1 {
		return fmt.Errorf("expected 1 fund member, found %d", memberCount)
	}

	// 3. Check that all financial data has the correct fund_id
	err = s.validateFinancialDataFundId(c, sess, user.Uid, fund.FundId)
	if err != nil {
		return fmt.Errorf("financial data validation failed: %w", err)
	}

	return nil
}

// validateFinancialDataFundId validates that all financial data has the correct fund_id
func (s *MigrationService) validateFinancialDataFundId(c core.Context, sess *xorm.Session, uid int64, fundId int64) error {
	// Check accounts
	accountCount, err := sess.Where("uid=? AND fund_id!=?", uid, fundId).Count(&models.Account{})
	if err != nil {
		return fmt.Errorf("failed to check accounts: %w", err)
	}
	if accountCount > 0 {
		return fmt.Errorf("found %d accounts with incorrect fund_id", accountCount)
	}

	// Check transactions
	transactionCount, err := sess.Where("uid=? AND fund_id!=?", uid, fundId).Count(&models.Transaction{})
	if err != nil {
		return fmt.Errorf("failed to check transactions: %w", err)
	}
	if transactionCount > 0 {
		return fmt.Errorf("found %d transactions with incorrect fund_id", transactionCount)
	}

	// Check transaction categories
	categoryCount, err := sess.Where("uid=? AND fund_id!=?", uid, fundId).Count(&models.TransactionCategory{})
	if err != nil {
		return fmt.Errorf("failed to check transaction categories: %w", err)
	}
	if categoryCount > 0 {
		return fmt.Errorf("found %d transaction categories with incorrect fund_id", categoryCount)
	}

	// Check transaction tags
	tagCount, err := sess.Where("uid=? AND fund_id!=?", uid, fundId).Count(&models.TransactionTag{})
	if err != nil {
		return fmt.Errorf("failed to check transaction tags: %w", err)
	}
	if tagCount > 0 {
		return fmt.Errorf("found %d transaction tags with incorrect fund_id", tagCount)
	}

	return nil
}

// rollbackUserMigration rolls back a user's migration (for testing)
func (s *MigrationService) rollbackUserMigration(c core.Context, user *models.User) error {
	return s.UserDataDB(user.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		// 1. Remove fund_id from all financial data
		_, err := sess.Where("uid=?", user.Uid).Update(&models.Account{FundId: 0})
		if err != nil {
			return fmt.Errorf("failed to rollback accounts: %w", err)
		}

		_, err = sess.Where("uid=?", user.Uid).Update(&models.Transaction{FundId: 0})
		if err != nil {
			return fmt.Errorf("failed to rollback transactions: %w", err)
		}

		_, err = sess.Where("uid=?", user.Uid).Update(&models.TransactionCategory{FundId: 0})
		if err != nil {
			return fmt.Errorf("failed to rollback transaction categories: %w", err)
		}

		_, err = sess.Where("uid=?", user.Uid).Update(&models.TransactionTag{FundId: 0})
		if err != nil {
			return fmt.Errorf("failed to rollback transaction tags: %w", err)
		}

		_, err = sess.Where("uid=?", user.Uid).Update(&models.TransactionTemplate{FundId: 0})
		if err != nil {
			return fmt.Errorf("failed to rollback transaction templates: %w", err)
		}

		_, err = sess.Where("uid=?", user.Uid).Update(&models.TransactionPictureInfo{FundId: 0})
		if err != nil {
			return fmt.Errorf("failed to rollback transaction picture info: %w", err)
		}

		// 2. Delete fund members
		_, err = sess.Where("linked_uid=?", user.Uid).Delete(&models.FundMember{})
		if err != nil {
			return fmt.Errorf("failed to delete fund members: %w", err)
		}

		// 3. Delete funds
		_, err = sess.Where("owner_uid=?", user.Uid).Delete(&models.Fund{})
		if err != nil {
			return fmt.Errorf("failed to delete funds: %w", err)
		}

		return nil
	})
}
