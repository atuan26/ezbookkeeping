package models

// FundRole represents fund member role
type FundRole byte

// Fund roles
const (
	FUND_ROLE_OWNER  FundRole = 1 // Full access
	FUND_ROLE_MEMBER FundRole = 2 // Read-only
)

// String returns a textual representation of the fund role enum
func (r FundRole) String() string {
	switch r {
	case FUND_ROLE_OWNER:
		return "Owner"
	case FUND_ROLE_MEMBER:
		return "Member"
	default:
		return "Unknown"
	}
}

// Fund represents fund data stored in database
type Fund struct {
	FundId          int64  `xorm:"PK"`
	Name            string `xorm:"VARCHAR(64) NOT NULL"`
	OwnerUid        int64  `xorm:"INDEX(IDX_fund_owner_uid) NOT NULL"` // FK to users
	DefaultCurrency string `xorm:"VARCHAR(3) NOT NULL"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
	Deleted         bool `xorm:"INDEX(IDX_fund_deleted) NOT NULL"`
}

// FundMember represents fund member data stored in database
type FundMember struct {
	MemberId        int64    `xorm:"PK"`
	FundId          int64    `xorm:"INDEX(IDX_fund_member_fund_id) NOT NULL"`
	Name            string   `xorm:"VARCHAR(64) NOT NULL"`
	Email           string   `xorm:"VARCHAR(100)"`
	Role            FundRole `xorm:"TINYINT NOT NULL"`
	LinkedUid       int64    `xorm:"INDEX(IDX_fund_member_linked_uid) DEFAULT 0"` // 0 = not linked, >0 = linked to real user
	CreatedBy       int64    `xorm:"NOT NULL"`                                    // FK to creating user
	CreatedUnixTime int64
	UpdatedUnixTime int64
}

// TransactionMember represents transaction-member relationship stored in database
type TransactionMember struct {
	TransactionId   int64 `xorm:"PK"`
	MemberId        int64 `xorm:"PK INDEX(IDX_transaction_member_member_id)"` // FK to fund_member
	CreatedUnixTime int64
}

// FundCreateRequest represents all parameters of fund creation request
type FundCreateRequest struct {
	Name            string `json:"name" binding:"required,notBlank,max=64"`
	DefaultCurrency string `json:"defaultCurrency" binding:"required,len=3,validCurrency"`
	ClientSessionId string `json:"clientSessionId"`
}

// FundModifyRequest represents all parameters of fund modification request
type FundModifyRequest struct {
	// Id              int64  `json:"id,string" binding:"required,min=1"`
	Name            string `json:"name" binding:"required,notBlank,max=64"`
	DefaultCurrency string `json:"defaultCurrency" binding:"required,len=3,validCurrency"`
}

// FundDeleteRequest represents all parameters of fund deleting request
type FundDeleteRequest struct {
	// Id will be retrieved from URL context parameter
}

// FundGetRequest represents all parameters of fund getting request
type FundGetRequest struct {
	// Id will be retrieved from URL context parameter
}

// FundMemberListRequest represents all parameters of fund member list request
type FundMemberListRequest struct {
	// FundId will be retrieved from URL context parameter
}

// FundMemberCreateRequest represents all parameters of fund member creation request
type FundMemberCreateRequest struct {
	// FundId          int64  `json:"fundId,string" binding:"required,min=1"`
	Name            string `json:"name" binding:"required,notBlank,max=64"`
	Email           string `json:"email" binding:"omitempty,max=100,validEmail"`
	ClientSessionId string `json:"clientSessionId"`
}

// FundMemberLinkRequest represents all parameters of fund member linking request
type FundMemberLinkRequest struct {
	MemberId  int64 `json:"memberId,string" binding:"required,min=1"`
	LinkedUid int64 `json:"linkedUid,string" binding:"required,min=1"`
}

// FundMemberDeleteRequest represents all parameters of fund member deleting request
type FundMemberDeleteRequest struct {
	// FundId will be retrieved from URL context parameter
	MemberId int64 `json:"memberId,string" binding:"required,min=1"`
}

// FundInfoResponse represents a view-object of fund
type FundInfoResponse struct {
	Id              int64    `json:"id,string"`
	Name            string   `json:"name"`
	Role            FundRole `json:"role"`
	MemberCount     int      `json:"memberCount"`
	DefaultCurrency string   `json:"defaultCurrency"`
	CreatedAt       int64    `json:"createdAt"`
}

// FundMemberResponse represents a view-object of fund member
type FundMemberResponse struct {
	MemberId  int64    `json:"memberId,string"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Role      FundRole `json:"role"`
	LinkedUid int64    `json:"linkedUid,string"`
	IsLinked  bool     `json:"isLinked"`
}

// ToFundInfoResponse returns a view-object according to database model
func (f *Fund) ToFundInfoResponse(role FundRole, memberCount int) *FundInfoResponse {
	return &FundInfoResponse{
		Id:              f.FundId,
		Name:            f.Name,
		Role:            role,
		MemberCount:     memberCount,
		DefaultCurrency: f.DefaultCurrency,
		CreatedAt:       f.CreatedUnixTime,
	}
}

// ToFundMemberResponse returns a view-object according to database model
func (fm *FundMember) ToFundMemberResponse() *FundMemberResponse {
	return &FundMemberResponse{
		MemberId:  fm.MemberId,
		Name:      fm.Name,
		Email:     fm.Email,
		Role:      fm.Role,
		LinkedUid: fm.LinkedUid,
		IsLinked:  fm.LinkedUid > 0,
	}
}

// FundInfoResponseSlice represents the slice data structure of FundInfoResponse
type FundInfoResponseSlice []*FundInfoResponse

// Len returns the count of items
func (s FundInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s FundInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s FundInfoResponseSlice) Less(i, j int) bool {
	return s[i].CreatedAt < s[j].CreatedAt
}

// FundMemberResponseSlice represents the slice data structure of FundMemberResponse
type FundMemberResponseSlice []*FundMemberResponse

// Len returns the count of items
func (s FundMemberResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s FundMemberResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s FundMemberResponseSlice) Less(i, j int) bool {
	// Owners first, then by creation time
	if s[i].Role != s[j].Role {
		return s[i].Role < s[j].Role
	}
	return s[i].MemberId < s[j].MemberId
}
