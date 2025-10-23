package models

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFundRole_String(t *testing.T) {
	assert.Equal(t, "Owner", FUND_ROLE_OWNER.String())
	assert.Equal(t, "Member", FUND_ROLE_MEMBER.String())
	assert.Equal(t, "Unknown", FundRole(99).String())
}

func TestFund_ToFundInfoResponse(t *testing.T) {
	fund := &Fund{
		FundId:          1001,
		Name:            "Test Fund",
		OwnerUid:        2001,
		DefaultCurrency: "USD",
		CreatedUnixTime: 1234567890,
	}

	response := fund.ToFundInfoResponse(FUND_ROLE_OWNER, 3)

	assert.Equal(t, int64(1001), response.Id)
	assert.Equal(t, "Test Fund", response.Name)
	assert.Equal(t, FUND_ROLE_OWNER, response.Role)
	assert.Equal(t, 3, response.MemberCount)
	assert.Equal(t, "USD", response.DefaultCurrency)
	assert.Equal(t, int64(1234567890), response.CreatedAt)
}

func TestFundMember_ToFundMemberResponse(t *testing.T) {
	// Test linked member
	linkedMember := &FundMember{
		MemberId:  1001,
		Name:      "John Doe",
		Email:     "john@example.com",
		Role:      FUND_ROLE_OWNER,
		LinkedUid: 2001,
	}

	response := linkedMember.ToFundMemberResponse()

	assert.Equal(t, int64(1001), response.MemberId)
	assert.Equal(t, "John Doe", response.Name)
	assert.Equal(t, "john@example.com", response.Email)
	assert.Equal(t, FUND_ROLE_OWNER, response.Role)
	assert.Equal(t, int64(2001), response.LinkedUid)
	assert.True(t, response.IsLinked)

	// Test unlinked member
	unlinkedMember := &FundMember{
		MemberId:  1002,
		Name:      "Jane Smith",
		Email:     "jane@example.com",
		Role:      FUND_ROLE_MEMBER,
		LinkedUid: 0,
	}

	response2 := unlinkedMember.ToFundMemberResponse()

	assert.Equal(t, int64(1002), response2.MemberId)
	assert.Equal(t, "Jane Smith", response2.Name)
	assert.Equal(t, "jane@example.com", response2.Email)
	assert.Equal(t, FUND_ROLE_MEMBER, response2.Role)
	assert.Equal(t, int64(0), response2.LinkedUid)
	assert.False(t, response2.IsLinked)
}

func TestFundInfoResponseSlice_Sort(t *testing.T) {
	var fundSlice FundInfoResponseSlice
	fundSlice = append(fundSlice, &FundInfoResponse{
		Id:        3,
		CreatedAt: 1234567892,
	})
	fundSlice = append(fundSlice, &FundInfoResponse{
		Id:        1,
		CreatedAt: 1234567890,
	})
	fundSlice = append(fundSlice, &FundInfoResponse{
		Id:        2,
		CreatedAt: 1234567891,
	})

	sort.Sort(fundSlice)

	assert.Equal(t, int64(1), fundSlice[0].Id)
	assert.Equal(t, int64(2), fundSlice[1].Id)
	assert.Equal(t, int64(3), fundSlice[2].Id)
}

func TestFundMemberResponseSlice_Sort(t *testing.T) {
	var memberSlice FundMemberResponseSlice

	// Add members with different roles and IDs
	memberSlice = append(memberSlice, &FundMemberResponse{
		MemberId: 3,
		Role:     FUND_ROLE_MEMBER,
	})
	memberSlice = append(memberSlice, &FundMemberResponse{
		MemberId: 1,
		Role:     FUND_ROLE_OWNER,
	})
	memberSlice = append(memberSlice, &FundMemberResponse{
		MemberId: 2,
		Role:     FUND_ROLE_OWNER,
	})
	memberSlice = append(memberSlice, &FundMemberResponse{
		MemberId: 4,
		Role:     FUND_ROLE_MEMBER,
	})

	sort.Sort(memberSlice)

	// Owners should come first (role 1), then members (role 2)
	// Within same role, sorted by member ID
	assert.Equal(t, int64(1), memberSlice[0].MemberId)
	assert.Equal(t, FUND_ROLE_OWNER, memberSlice[0].Role)
	assert.Equal(t, int64(2), memberSlice[1].MemberId)
	assert.Equal(t, FUND_ROLE_OWNER, memberSlice[1].Role)
	assert.Equal(t, int64(3), memberSlice[2].MemberId)
	assert.Equal(t, FUND_ROLE_MEMBER, memberSlice[2].Role)
	assert.Equal(t, int64(4), memberSlice[3].MemberId)
	assert.Equal(t, FUND_ROLE_MEMBER, memberSlice[3].Role)
}

func TestFundInfoResponseSlice_Len(t *testing.T) {
	var fundSlice FundInfoResponseSlice
	assert.Equal(t, 0, fundSlice.Len())

	fundSlice = append(fundSlice, &FundInfoResponse{})
	fundSlice = append(fundSlice, &FundInfoResponse{})
	assert.Equal(t, 2, fundSlice.Len())
}

func TestFundMemberResponseSlice_Len(t *testing.T) {
	var memberSlice FundMemberResponseSlice
	assert.Equal(t, 0, memberSlice.Len())

	memberSlice = append(memberSlice, &FundMemberResponse{})
	memberSlice = append(memberSlice, &FundMemberResponse{})
	memberSlice = append(memberSlice, &FundMemberResponse{})
	assert.Equal(t, 3, memberSlice.Len())
}

func TestFundInfoResponseSlice_Swap(t *testing.T) {
	var fundSlice FundInfoResponseSlice
	fundSlice = append(fundSlice, &FundInfoResponse{Id: 1})
	fundSlice = append(fundSlice, &FundInfoResponse{Id: 2})

	fundSlice.Swap(0, 1)

	assert.Equal(t, int64(2), fundSlice[0].Id)
	assert.Equal(t, int64(1), fundSlice[1].Id)
}

func TestFundMemberResponseSlice_Swap(t *testing.T) {
	var memberSlice FundMemberResponseSlice
	memberSlice = append(memberSlice, &FundMemberResponse{MemberId: 1})
	memberSlice = append(memberSlice, &FundMemberResponse{MemberId: 2})

	memberSlice.Swap(0, 1)

	assert.Equal(t, int64(2), memberSlice[0].MemberId)
	assert.Equal(t, int64(1), memberSlice[1].MemberId)
}
