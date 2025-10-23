package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionMemberService_GetTransactionMembersByTransactionId_InvalidUserId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionMembersByTransactionId(nil, 0, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = service.GetTransactionMembersByTransactionId(nil, -1, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestTransactionMemberService_GetTransactionMembersByTransactionId_InvalidTransactionId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionMembersByTransactionId(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")

	_, err = service.GetTransactionMembersByTransactionId(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")
}

func TestTransactionMemberService_GetMemberIdsByTransactionId_InvalidUserId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetMemberIdsByTransactionId(nil, 0, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = service.GetMemberIdsByTransactionId(nil, -1, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestTransactionMemberService_GetMemberIdsByTransactionId_InvalidTransactionId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetMemberIdsByTransactionId(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")

	_, err = service.GetMemberIdsByTransactionId(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")
}

func TestTransactionMemberService_GetTransactionsByMemberId_InvalidUserId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionsByMemberId(nil, 0, 1001, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = service.GetTransactionsByMemberId(nil, -1, 1001, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestTransactionMemberService_GetTransactionsByMemberId_InvalidFundId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionsByMemberId(nil, 1001, 0, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")

	_, err = service.GetTransactionsByMemberId(nil, 1001, -1, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")
}

func TestTransactionMemberService_GetTransactionsByMemberId_InvalidMemberId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionsByMemberId(nil, 1001, 2001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")

	_, err = service.GetTransactionsByMemberId(nil, 1001, 2001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")
}

func TestTransactionMemberService_LinkTransactionMembers_InvalidUserId(t *testing.T) {
	service := &TransactionMemberService{}
	memberIds := []int64{1001, 1002}

	err := service.LinkTransactionMembers(nil, 0, 2001, memberIds)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	err = service.LinkTransactionMembers(nil, -1, 2001, memberIds)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestTransactionMemberService_LinkTransactionMembers_InvalidTransactionId(t *testing.T) {
	service := &TransactionMemberService{}
	memberIds := []int64{1001, 1002}

	err := service.LinkTransactionMembers(nil, 1001, 0, memberIds)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")

	err = service.LinkTransactionMembers(nil, 1001, -1, memberIds)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")
}

func TestTransactionMemberService_LinkTransactionMembers_EmptyMemberIds(t *testing.T) {
	service := &TransactionMemberService{}
	memberIds := []int64{}

	// When empty member IDs are provided, it should attempt to link to all fund members
	// This will fail due to no database connection, but validates that the method
	// accepts empty member IDs and attempts the linkTransactionToAllMembers path
	err := service.LinkTransactionMembers(nil, 1001, 2001, memberIds)
	assert.NotNil(t, err) // Expected to fail due to no DB connection
	// The error should not be a validation error for user ID or transaction ID
	assert.NotContains(t, err.Error(), "user id is invalid")
	assert.NotContains(t, err.Error(), "transaction id is invalid")
}

func TestTransactionMemberService_UnlinkTransactionMembers_InvalidUserId(t *testing.T) {
	service := &TransactionMemberService{}

	err := service.UnlinkTransactionMembers(nil, 0, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	err = service.UnlinkTransactionMembers(nil, -1, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestTransactionMemberService_UnlinkTransactionMembers_InvalidTransactionId(t *testing.T) {
	service := &TransactionMemberService{}

	err := service.UnlinkTransactionMembers(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")

	err = service.UnlinkTransactionMembers(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")
}

func TestTransactionMemberService_GetTransactionMembersWithDetails_InvalidUserId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionMembersWithDetails(nil, 0, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = service.GetTransactionMembersWithDetails(nil, -1, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestTransactionMemberService_GetTransactionMembersWithDetails_InvalidTransactionId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionMembersWithDetails(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")

	_, err = service.GetTransactionMembersWithDetails(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")
}

func TestTransactionMemberService_IsTransactionLinkedToMember_InvalidUserId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.IsTransactionLinkedToMember(nil, 0, 2001, 3001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = service.IsTransactionLinkedToMember(nil, -1, 2001, 3001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestTransactionMemberService_IsTransactionLinkedToMember_InvalidTransactionId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.IsTransactionLinkedToMember(nil, 1001, 0, 3001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")

	_, err = service.IsTransactionLinkedToMember(nil, 1001, -1, 3001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "transaction id is invalid")
}

func TestTransactionMemberService_IsTransactionLinkedToMember_InvalidMemberId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.IsTransactionLinkedToMember(nil, 1001, 2001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")

	_, err = service.IsTransactionLinkedToMember(nil, 1001, 2001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")
}

func TestTransactionMemberService_GetTransactionCountByMemberId_InvalidUserId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionCountByMemberId(nil, 0, 1001, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = service.GetTransactionCountByMemberId(nil, -1, 1001, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestTransactionMemberService_GetTransactionCountByMemberId_InvalidFundId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionCountByMemberId(nil, 1001, 0, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")

	_, err = service.GetTransactionCountByMemberId(nil, 1001, -1, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")
}

func TestTransactionMemberService_GetTransactionCountByMemberId_InvalidMemberId(t *testing.T) {
	service := &TransactionMemberService{}

	_, err := service.GetTransactionCountByMemberId(nil, 1001, 2001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")

	_, err = service.GetTransactionCountByMemberId(nil, 1001, 2001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")
}

func TestTransactionMemberService_DeleteTransactionMembersByMemberId_InvalidUserId(t *testing.T) {
	service := &TransactionMemberService{}

	err := service.DeleteTransactionMembersByMemberId(nil, 0, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	err = service.DeleteTransactionMembersByMemberId(nil, -1, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestTransactionMemberService_DeleteTransactionMembersByMemberId_InvalidMemberId(t *testing.T) {
	service := &TransactionMemberService{}

	err := service.DeleteTransactionMembersByMemberId(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")

	err = service.DeleteTransactionMembersByMemberId(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")
}
