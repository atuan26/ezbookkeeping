package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func TestFundService_GetUserRoleInFund_InvalidUserId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.GetUserRoleInFund(nil, 0, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = fundService.GetUserRoleInFund(nil, -1, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestFundService_GetUserRoleInFund_InvalidFundId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.GetUserRoleInFund(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")

	_, err = fundService.GetUserRoleInFund(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")
}

func TestFundService_GetFundByFundId_InvalidUserId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.GetFundByFundId(nil, 0, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = fundService.GetFundByFundId(nil, -1, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestFundService_GetFundByFundId_InvalidFundId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.GetFundByFundId(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")

	_, err = fundService.GetFundByFundId(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")
}

func TestFundService_GetAllFundsByUid_InvalidUserId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.GetAllFundsByUid(nil, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = fundService.GetAllFundsByUid(nil, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestFundService_GetFundMembersByFundId_InvalidUserId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.GetFundMembersByFundId(nil, 0, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = fundService.GetFundMembersByFundId(nil, -1, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestFundService_GetFundMembersByFundId_InvalidFundId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.GetFundMembersByFundId(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")

	_, err = fundService.GetFundMembersByFundId(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")
}

func TestFundService_GetFundMemberByMemberId_InvalidUserId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.GetFundMemberByMemberId(nil, 0, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = fundService.GetFundMemberByMemberId(nil, -1, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestFundService_GetFundMemberByMemberId_InvalidMemberId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.GetFundMemberByMemberId(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")

	_, err = fundService.GetFundMemberByMemberId(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")
}

func TestFundService_ModifyFund_InvalidUserId(t *testing.T) {
	fundService := &FundService{}
	fund := &models.Fund{FundId: 1001}

	err := fundService.ModifyFund(nil, 0, fund)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	err = fundService.ModifyFund(nil, -1, fund)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestFundService_ModifyFund_InvalidFundId(t *testing.T) {
	fundService := &FundService{}

	fund := &models.Fund{FundId: 0}
	err := fundService.ModifyFund(nil, 1001, fund)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")

	fund = &models.Fund{FundId: -1}
	err = fundService.ModifyFund(nil, 1001, fund)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")
}

func TestFundService_DeleteFund_InvalidUserId(t *testing.T) {
	fundService := &FundService{}

	err := fundService.DeleteFund(nil, 0, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	err = fundService.DeleteFund(nil, -1, 1001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestFundService_DeleteFund_InvalidFundId(t *testing.T) {
	fundService := &FundService{}

	err := fundService.DeleteFund(nil, 1001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")

	err = fundService.DeleteFund(nil, 1001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")
}

func TestFundService_RemoveMember_InvalidUserId(t *testing.T) {
	fundService := &FundService{}

	err := fundService.RemoveMember(nil, 0, 1001, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	err = fundService.RemoveMember(nil, -1, 1001, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestFundService_RemoveMember_InvalidFundId(t *testing.T) {
	fundService := &FundService{}

	err := fundService.RemoveMember(nil, 1001, 0, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")

	err = fundService.RemoveMember(nil, 1001, -1, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "fund id is invalid")
}

func TestFundService_RemoveMember_InvalidMemberId(t *testing.T) {
	fundService := &FundService{}

	err := fundService.RemoveMember(nil, 1001, 2001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")

	err = fundService.RemoveMember(nil, 1001, 2001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")
}

func TestFundService_LinkMemberToUser_InvalidUserId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.LinkMemberToUser(nil, 0, 1001, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = fundService.LinkMemberToUser(nil, -1, 1001, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}

func TestFundService_LinkMemberToUser_InvalidMemberId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.LinkMemberToUser(nil, 1001, 0, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")

	_, err = fundService.LinkMemberToUser(nil, 1001, -1, 2001)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "member id is invalid")
}

func TestFundService_LinkMemberToUser_InvalidLinkedUserId(t *testing.T) {
	fundService := &FundService{}

	_, err := fundService.LinkMemberToUser(nil, 1001, 2001, 0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")

	_, err = fundService.LinkMemberToUser(nil, 1001, 2001, -1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "user id is invalid")
}
