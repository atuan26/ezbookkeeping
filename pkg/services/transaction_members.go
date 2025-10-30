package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// TransactionMemberService represents transaction member service
type TransactionMemberService struct {
	ServiceUsingDB
}

// Initialize a transaction member service singleton instance
var (
	TransactionMembers = &TransactionMemberService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
	}
)

// GetTransactionMembersByTransactionId returns all members linked to a transaction
func (s *TransactionMemberService) GetTransactionMembersByTransactionId(c core.Context, uid int64, transactionId int64) ([]*models.TransactionMember, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if transactionId <= 0 {
		return nil, errs.ErrTransactionIdInvalid
	}

	var transactionMembers []*models.TransactionMember
	err := s.UserDataDB(uid).NewSession(c).Where("transaction_id=?", transactionId).Find(&transactionMembers)

	return transactionMembers, err
}

// GetMemberIdsByTransactionId returns all member IDs linked to a transaction
func (s *TransactionMemberService) GetMemberIdsByTransactionId(c core.Context, uid int64, transactionId int64) ([]int64, error) {
	transactionMembers, err := s.GetTransactionMembersByTransactionId(c, uid, transactionId)
	if err != nil {
		return nil, err
	}

	memberIds := make([]int64, len(transactionMembers))
	for i, tm := range transactionMembers {
		memberIds[i] = tm.MemberId
	}

	return memberIds, nil
}

// GetTransactionsByMemberId returns all transactions linked to a specific member
func (s *TransactionMemberService) GetTransactionsByMemberId(c core.Context, uid int64, fundId int64, memberId int64) ([]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if fundId <= 0 {
		return nil, errs.ErrFundIdInvalid
	}

	if memberId <= 0 {
		return nil, errs.ErrMemberIdInvalid
	}

	// Check if user has access to this fund
	if !s.canUserAccessFund(c, uid, fundId) {
		return nil, errs.ErrFundAccessDenied
	}

	var transactions []*models.Transaction
	err := s.UserDataDB(uid).NewSession(c).
		Join("INNER", "transaction_members", "transactions.transaction_id = transaction_members.transaction_id").
		Where("transaction_members.member_id=? AND transactions.fund_id=? AND transactions.uid=? AND transactions.deleted=?", memberId, fundId, uid, false).
		OrderBy("transactions.transaction_time desc").
		Find(&transactions)

	return transactions, err
}

// LinkTransactionMembers links a transaction to multiple members
func (s *TransactionMemberService) LinkTransactionMembers(c core.Context, uid int64, transactionId int64, memberIds []int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if transactionId <= 0 {
		return errs.ErrTransactionIdInvalid
	}

	if len(memberIds) == 0 {
		// If no member IDs provided, link to all members of the fund
		return s.linkTransactionToAllMembers(c, uid, transactionId)
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// First, remove existing links
		_, err := sess.Where("transaction_id=?", transactionId).Delete(&models.TransactionMember{})
		if err != nil {
			log.Errorf(c, "[transaction_members.LinkTransactionMembers] failed to delete existing transaction members for transaction \"transaction_id:%d\", because %s", transactionId, err.Error())
			return err
		}

		// Then, create new links
		now := time.Now().Unix()
		for _, memberId := range memberIds {
			transactionMember := &models.TransactionMember{
				TransactionId:   transactionId,
				MemberId:        memberId,
				CreatedUnixTime: now,
			}

			_, err := sess.Insert(transactionMember)
			if err != nil {
				log.Errorf(c, "[transaction_members.LinkTransactionMembers] failed to insert transaction member \"transaction_id:%d, member_id:%d\", because %s", transactionId, memberId, err.Error())
				return err
			}
		}

		return nil
	})
}

// UnlinkTransactionMembers removes all member links from a transaction
func (s *TransactionMemberService) UnlinkTransactionMembers(c core.Context, uid int64, transactionId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if transactionId <= 0 {
		return errs.ErrTransactionIdInvalid
	}

	_, err := s.UserDataDB(uid).NewSession(c).Where("transaction_id=?", transactionId).Delete(&models.TransactionMember{})
	if err != nil {
		log.Errorf(c, "[transaction_members.UnlinkTransactionMembers] failed to delete transaction members for transaction \"transaction_id:%d\", because %s", transactionId, err.Error())
		return err
	}

	return nil
}

// linkTransactionToAllMembers links a transaction to all members of its fund
func (s *TransactionMemberService) linkTransactionToAllMembers(c core.Context, uid int64, transactionId int64) error {
	// Guard against nil container (e.g., in unit tests)
	if s.container == nil {
		return errs.ErrSystemError
	}

	// First, get the transaction to find its fund
	transaction := &models.Transaction{}
	has, err := s.UserDataDB(uid).NewSession(c).Where("transaction_id=? AND uid=? AND deleted=?", transactionId, uid, false).Get(transaction)
	if err != nil {
		return err
	}
	if !has {
		return errs.ErrTransactionNotFound
	}

	// Get all members of the fund
	var fundMembers []*models.FundMember
	err = s.UserDataDB(uid).NewSession(c).Where("fund_id=?", transaction.FundId).Find(&fundMembers)
	if err != nil {
		return err
	}

	// Extract member IDs
	memberIds := make([]int64, len(fundMembers))
	for i, member := range fundMembers {
		memberIds[i] = member.MemberId
	}

	// Link to all members
	return s.LinkTransactionMembers(c, uid, transactionId, memberIds)
}

// GetTransactionMembersWithDetails returns transaction members with fund member details
func (s *TransactionMemberService) GetTransactionMembersWithDetails(c core.Context, uid int64, transactionId int64) ([]*models.FundMember, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if transactionId <= 0 {
		return nil, errs.ErrTransactionIdInvalid
	}

	var fundMembers []*models.FundMember
	err := s.UserDataDB(uid).NewSession(c).
		Join("INNER", "transaction_members", "fund_member.member_id = transaction_members.member_id").
		Where("transaction_members.transaction_id=?", transactionId).
		OrderBy("fund_member.role asc, fund_member.member_id asc").
		Find(&fundMembers)

	return fundMembers, err
}

// IsTransactionLinkedToMember checks if a transaction is linked to a specific member
func (s *TransactionMemberService) IsTransactionLinkedToMember(c core.Context, uid int64, transactionId int64, memberId int64) (bool, error) {
	if uid <= 0 {
		return false, errs.ErrUserIdInvalid
	}

	if transactionId <= 0 {
		return false, errs.ErrTransactionIdInvalid
	}

	if memberId <= 0 {
		return false, errs.ErrMemberIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("transaction_id=? AND member_id=?", transactionId, memberId).Count(&models.TransactionMember{})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetTransactionCountByMemberId returns the count of transactions linked to a specific member
func (s *TransactionMemberService) GetTransactionCountByMemberId(c core.Context, uid int64, fundId int64, memberId int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	if fundId <= 0 {
		return 0, errs.ErrFundIdInvalid
	}

	if memberId <= 0 {
		return 0, errs.ErrMemberIdInvalid
	}

	// Check if user has access to this fund
	if !s.canUserAccessFund(c, uid, fundId) {
		return 0, errs.ErrFundAccessDenied
	}

	count, err := s.UserDataDB(uid).NewSession(c).
		Join("INNER", "transaction_members", "transactions.transaction_id = transaction_members.transaction_id").
		Where("transaction_members.member_id=? AND transactions.fund_id=? AND transactions.uid=? AND transactions.deleted=?", memberId, fundId, uid, false).
		Count(&models.Transaction{})

	return count, err
}

// DeleteTransactionMembersByMemberId removes all transaction links for a specific member
// This is typically called when a member is removed from a fund
func (s *TransactionMemberService) DeleteTransactionMembersByMemberId(c core.Context, uid int64, memberId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if memberId <= 0 {
		return errs.ErrMemberIdInvalid
	}

	_, err := s.UserDataDB(uid).NewSession(c).Where("member_id=?", memberId).Delete(&models.TransactionMember{})
	if err != nil {
		log.Errorf(c, "[transaction_members.DeleteTransactionMembersByMemberId] failed to delete transaction members for member \"member_id:%d\", because %s", memberId, err.Error())
		return err
	}

	return nil
}

// canUserAccessFund checks if user has access to a fund (either as owner or member)
func (s *TransactionMemberService) canUserAccessFund(c core.Context, uid int64, fundId int64) bool {
	count, err := s.UserDataDB(uid).NewSession(c).Where("fund_id=? AND linked_uid=?", fundId, uid).Count(&models.FundMember{})
	if err != nil {
		log.Errorf(c, "[transaction_members.canUserAccessFund] failed to check fund access for user \"uid:%d\" and fund \"fund_id:%d\", because %s", uid, fundId, err.Error())
		return false
	}

	return count > 0
}
