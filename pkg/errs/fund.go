package errs

import "net/http"

// Error codes related to funds
var (
	ErrFundIdInvalid        = NewNormalError(NormalSubcategoryFund, 0, http.StatusBadRequest, "fund id is invalid")
	ErrFundNotFound         = NewNormalError(NormalSubcategoryFund, 1, http.StatusBadRequest, "fund not found")
	ErrFundAccessDenied     = NewNormalError(NormalSubcategoryFund, 2, http.StatusForbidden, "fund access denied")
	ErrFundNameExists       = NewNormalError(NormalSubcategoryFund, 3, http.StatusBadRequest, "fund name already exists")
	ErrMemberIdInvalid      = NewNormalError(NormalSubcategoryFund, 4, http.StatusBadRequest, "member id is invalid")
	ErrMemberNotFound       = NewNormalError(NormalSubcategoryFund, 5, http.StatusBadRequest, "member not found")
	ErrCannotRemoveOwner    = NewNormalError(NormalSubcategoryFund, 6, http.StatusBadRequest, "cannot remove fund owner")
	ErrInvalidFundRole      = NewNormalError(NormalSubcategoryFund, 7, http.StatusBadRequest, "invalid fund role")
	ErrMemberAlreadyLinked  = NewNormalError(NormalSubcategoryFund, 8, http.StatusBadRequest, "member already linked to user")
	ErrCannotLinkToSelf     = NewNormalError(NormalSubcategoryFund, 9, http.StatusBadRequest, "cannot link member to self")
)