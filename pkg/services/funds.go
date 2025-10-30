package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// FundService represents fund service
type FundService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a fund service singleton instance
var (
	Funds = &FundService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetFundByFundId returns fund model according to fund id
func (s *FundService) GetFundByFundId(c core.Context, uid int64, fundId int64) (*models.Fund, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if fundId <= 0 {
		return nil, errs.ErrFundIdInvalid
	}

	fund := &models.Fund{}
	has, err := s.UserDataDB(uid).NewSession(c).Where("fund_id=? AND deleted=?", fundId, false).Get(fund)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrFundNotFound
	}

	// Check if user has access to this fund
	if !s.canUserAccessFund(c, uid, fundId) {
		return nil, errs.ErrFundAccessDenied
	}

	return fund, nil
}

// GetAllFundsByUid returns all fund models that user has access to
func (s *FundService) GetAllFundsByUid(c core.Context, uid int64) ([]*models.Fund, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	// Get funds where user is a member
	var funds []*models.Fund
	err := s.UserDataDB(uid).NewSession(c).
		Join("INNER", "fund_member", "fund.fund_id = fund_member.fund_id").
		Where("fund_member.linked_uid=? AND fund.deleted=?", uid, false).
		OrderBy("fund.created_unix_time asc").
		Find(&funds)

	return funds, err
}

// GetFundMembersByFundId returns all members of a fund
func (s *FundService) GetFundMembersByFundId(c core.Context, uid int64, fundId int64) ([]*models.FundMember, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if fundId <= 0 {
		return nil, errs.ErrFundIdInvalid
	}

	// Check if user has access to this fund
	if !s.canUserAccessFund(c, uid, fundId) {
		return nil, errs.ErrFundAccessDenied
	}

	var members []*models.FundMember
	err := s.UserDataDB(uid).NewSession(c).Where("fund_id=?", fundId).OrderBy("role asc, member_id asc").Find(&members)

	return members, err
}

// GetFundMemberByMemberId returns fund member model according to member id
func (s *FundService) GetFundMemberByMemberId(c core.Context, uid int64, memberId int64) (*models.FundMember, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if memberId <= 0 {
		return nil, errs.ErrMemberIdInvalid
	}

	member := &models.FundMember{}
	has, err := s.UserDataDB(uid).NewSession(c).Where("member_id=?", memberId).Get(member)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrMemberNotFound
	}

	// Check if user has access to this fund
	if !s.canUserAccessFund(c, uid, member.FundId) {
		return nil, errs.ErrFundAccessDenied
	}

	return member, nil
}

// createFundInternal saves a new fund model to database
func (s *FundService) createFundInternal(c core.Context, fund *models.Fund) error {
	if fund.OwnerUid <= 0 {
		return errs.ErrUserIdInvalid
	}

	fundId := s.GenerateUuid(uuid.UUID_TYPE_FUND)
	if fundId < 1 {
		return errs.ErrSystemIsBusy
	}

	memberId := s.GenerateUuid(uuid.UUID_TYPE_FUND_MEMBER)
	if memberId < 1 {
		return errs.ErrSystemIsBusy
	}

	now := time.Now().Unix()

	fund.FundId = fundId
	fund.Deleted = false
	fund.CreatedUnixTime = now
	fund.UpdatedUnixTime = now

	// Create fund member for owner
	ownerMember := &models.FundMember{
		MemberId:        memberId,
		FundId:          fundId,
		Name:            "", // Will be set from user info
		Email:           "", // Will be set from user info
		Role:            models.FUND_ROLE_OWNER,
		LinkedUid:       fund.OwnerUid,
		CreatedBy:       fund.OwnerUid,
		CreatedUnixTime: now,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(fund.OwnerUid).DoTransaction(c, func(sess *xorm.Session) error {
		// Insert fund
		_, err := sess.Insert(fund)
		if err != nil {
			log.Errorf(c, "[funds.CreateFund] failed to insert fund \"fund_id:%d\", because %s", fund.FundId, err.Error())
			return err
		}

		// Get user info to populate member details
		user := &models.User{}
		has, err := sess.Where("uid=?", fund.OwnerUid).Get(user)
		if err != nil {
			return err
		}
		if has {
			ownerMember.Name = user.Nickname
			ownerMember.Email = user.Email
		}

		// Insert owner as fund member
		_, err = sess.Insert(ownerMember)
		if err != nil {
			log.Errorf(c, "[funds.CreateFund] failed to insert fund member \"member_id:%d\", because %s", ownerMember.MemberId, err.Error())
			return err
		}

		return nil
	})
}

// ModifyFund updates an existing fund
func (s *FundService) ModifyFund(c core.Context, uid int64, fund *models.Fund) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if fund.FundId <= 0 {
		return errs.ErrFundIdInvalid
	}

	// Check if user is owner of this fund
	if !s.canUserModifyFund(c, uid, fund.FundId) {
		return errs.ErrFundAccessDenied
	}

	now := time.Now().Unix()
	fund.UpdatedUnixTime = now

	updatedRows, err := s.UserDataDB(uid).NewSession(c).Where("fund_id=? AND deleted=?", fund.FundId, false).
		Cols("name", "default_currency", "updated_unix_time").Update(fund)

	if err != nil {
		log.Errorf(c, "[funds.ModifyFund] failed to update fund \"fund_id:%d\", because %s", fund.FundId, err.Error())
		return err
	} else if updatedRows < 1 {
		return errs.ErrFundNotFound
	}

	return nil
}

// DeleteFund deletes an existing fund
func (s *FundService) DeleteFund(c core.Context, uid int64, fundId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if fundId <= 0 {
		return errs.ErrFundIdInvalid
	}

	// Check if user is owner of this fund
	if !s.canUserModifyFund(c, uid, fundId) {
		return errs.ErrFundAccessDenied
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// Soft delete fund
		updatedRows, err := sess.Where("fund_id=? AND deleted=?", fundId, false).
			Cols("deleted", "deleted_unix_time", "updated_unix_time").
			Update(&models.Fund{
				Deleted:         true,
				DeletedUnixTime: now,
				UpdatedUnixTime: now,
			})

		if err != nil {
			log.Errorf(c, "[funds.DeleteFund] failed to delete fund \"fund_id:%d\", because %s", fundId, err.Error())
			return err
		} else if updatedRows < 1 {
			return errs.ErrFundNotFound
		}

		// Note: We don't delete fund members or financial data here
		// They should be handled separately if needed

		return nil
	})
}

// addMemberInternal adds a new member to a fund
func (s *FundService) addMemberInternal(c core.Context, uid int64, fundId int64, member *models.FundMember) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if fundId <= 0 {
		return errs.ErrFundIdInvalid
	}

	// Check if user can modify this fund
	if !s.canUserModifyFund(c, uid, fundId) {
		return errs.ErrFundAccessDenied
	}

	memberId := s.GenerateUuid(uuid.UUID_TYPE_FUND_MEMBER)
	if memberId < 1 {
		return errs.ErrSystemIsBusy
	}

	now := time.Now().Unix()

	member.MemberId = memberId
	member.FundId = fundId
	member.Role = models.FUND_ROLE_MEMBER // New members are always regular members
	member.LinkedUid = 0                  // Not linked initially
	member.CreatedBy = uid
	member.CreatedUnixTime = now
	member.UpdatedUnixTime = now

	_, err := s.UserDataDB(uid).NewSession(c).Insert(member)
	if err != nil {
		log.Errorf(c, "[funds.AddMember] failed to insert fund member \"member_id:%d\", because %s", member.MemberId, err.Error())
		return err
	}

	return nil
}

// RemoveMember removes a member from a fund
func (s *FundService) RemoveMember(c core.Context, uid int64, fundId int64, memberId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if fundId <= 0 {
		return errs.ErrFundIdInvalid
	}

	if memberId <= 0 {
		return errs.ErrMemberIdInvalid
	}

	// Check if user can modify this fund
	if !s.canUserModifyFund(c, uid, fundId) {
		return errs.ErrFundAccessDenied
	}

	// Get member to check if it's the owner
	member, err := s.GetFundMemberByMemberId(c, uid, memberId)
	if err != nil {
		return err
	}

	if member.Role == models.FUND_ROLE_OWNER {
		return errs.ErrCannotRemoveOwner
	}

	deletedRows, err := s.UserDataDB(uid).NewSession(c).Where("member_id=? AND fund_id=?", memberId, fundId).Delete(&models.FundMember{})
	if err != nil {
		log.Errorf(c, "[funds.RemoveMember] failed to delete fund member \"member_id:%d\", because %s", memberId, err.Error())
		return err
	} else if deletedRows < 1 {
		return errs.ErrMemberNotFound
	}

	return nil
}

// linkMemberToUserInternal links a fund member to an existing user
func (s *FundService) linkMemberToUserInternal(c core.Context, uid int64, memberId int64, linkedUid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if memberId <= 0 {
		return errs.ErrMemberIdInvalid
	}

	if linkedUid <= 0 {
		return errs.ErrUserIdInvalid
	}

	// Get member to check fund access
	member, err := s.GetFundMemberByMemberId(c, uid, memberId)
	if err != nil {
		return err
	}

	// Check if user can modify this fund
	if !s.canUserModifyFund(c, uid, member.FundId) {
		return errs.ErrFundAccessDenied
	}

	now := time.Now().Unix()

	updatedRows, err := s.UserDataDB(uid).NewSession(c).Where("member_id=?", memberId).
		Cols("linked_uid", "updated_unix_time").
		Update(&models.FundMember{
			LinkedUid:       linkedUid,
			UpdatedUnixTime: now,
		})

	if err != nil {
		log.Errorf(c, "[funds.LinkMemberToUser] failed to link member \"member_id:%d\" to user \"uid:%d\", because %s", memberId, linkedUid, err.Error())
		return err
	} else if updatedRows < 1 {
		return errs.ErrMemberNotFound
	}

	return nil
}

// canUserAccessFund checks if user has access to a fund (either as owner or member)
func (s *FundService) canUserAccessFund(c core.Context, uid int64, fundId int64) bool {
	count, err := s.UserDataDB(uid).NewSession(c).Where("fund_id=? AND linked_uid=?", fundId, uid).Count(&models.FundMember{})
	if err != nil {
		log.Errorf(c, "[funds.canUserAccessFund] failed to check fund access for user \"uid:%d\" and fund \"fund_id:%d\", because %s", uid, fundId, err.Error())
		return false
	}

	return count > 0
}

// canUserModifyFund checks if user can modify a fund (must be owner)
func (s *FundService) canUserModifyFund(c core.Context, uid int64, fundId int64) bool {
	count, err := s.UserDataDB(uid).NewSession(c).Where("fund_id=? AND linked_uid=? AND role=?", fundId, uid, models.FUND_ROLE_OWNER).Count(&models.FundMember{})
	if err != nil {
		log.Errorf(c, "[funds.canUserModifyFund] failed to check fund modify access for user \"uid:%d\" and fund \"fund_id:%d\", because %s", uid, fundId, err.Error())
		return false
	}

	return count > 0
}

// GetUserRoleInFund returns the user's role in a specific fund
func (s *FundService) GetUserRoleInFund(c core.Context, uid int64, fundId int64) (models.FundRole, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	if fundId <= 0 {
		return 0, errs.ErrFundIdInvalid
	}

	member := &models.FundMember{}
	has, err := s.UserDataDB(uid).NewSession(c).Where("fund_id=? AND linked_uid=?", fundId, uid).Get(member)

	if err != nil {
		return 0, err
	} else if !has {
		return 0, errs.ErrFundAccessDenied
	}

	return member.Role, nil
}

// API wrapper methods

// GetUserFunds returns all funds that user has access to with role and member count
func (s *FundService) GetUserFunds(c core.Context, uid int64) ([]*models.Fund, error) {
	return s.GetAllFundsByUid(c, uid)
}

// GetFundMembers returns all members of a fund
func (s *FundService) GetFundMembers(c core.Context, uid int64, fundId int64) ([]*models.FundMember, error) {
	return s.GetFundMembersByFundId(c, uid, fundId)
}

// CreateFund creates a new fund with the given request parameters
func (s *FundService) CreateFund(c core.Context, uid int64, req *models.FundCreateRequest) (*models.Fund, error) {
	fund := &models.Fund{
		Name:            req.Name,
		OwnerUid:        uid,
		DefaultCurrency: req.DefaultCurrency,
	}

	err := s.createFundInternal(c, fund)
	if err != nil {
		return nil, err
	}

	return fund, nil
}

// ModifyFundWithRequest modifies an existing fund with the given request parameters
func (s *FundService) ModifyFundWithRequest(c core.Context, uid int64, fundId int64, req *models.FundModifyRequest) (*models.Fund, error) {
	fund := &models.Fund{
		FundId:          fundId,
		Name:            req.Name,
		DefaultCurrency: req.DefaultCurrency,
	}

	err := s.ModifyFund(c, uid, fund)
	if err != nil {
		return nil, err
	}

	// Return the updated fund
	return s.GetFundByFundId(c, uid, fundId)
}

// AddMember adds a new member to a fund with the given request parameters
func (s *FundService) AddMember(c core.Context, uid int64, fundId int64, req *models.FundMemberCreateRequest) (*models.FundMember, error) {
	member := &models.FundMember{
		Name:  req.Name,
		Email: req.Email,
	}

	err := s.addMemberInternal(c, uid, fundId, member)
	if err != nil {
		return nil, err
	}

	return member, nil
}

// LinkMemberToUser links a fund member to an existing user and returns the updated member
func (s *FundService) LinkMemberToUser(c core.Context, uid int64, memberId int64, linkedUid int64) (*models.FundMember, error) {
	err := s.linkMemberToUserInternal(c, uid, memberId, linkedUid)
	if err != nil {
		return nil, err
	}

	// Return the updated member
	return s.GetFundMemberByMemberId(c, uid, memberId)
}
