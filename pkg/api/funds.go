package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// FundsApi represents fund api
type FundsApi struct {
	ApiUsingConfig
	ApiUsingDuplicateChecker
	funds *services.FundService
}

// Initialize a fund api singleton instance
var (
	Funds = &FundsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ApiUsingDuplicateChecker: ApiUsingDuplicateChecker{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			container: duplicatechecker.Container,
		},
		funds: services.Funds,
	}
)

// FundListHandler returns funds list of current user
func (a *FundsApi) FundListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	funds, err := a.funds.GetUserFunds(c, uid)

	if err != nil {
		log.Errorf(c, "[funds.FundListHandler] failed to get all funds for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	fundResps := make([]*models.FundInfoResponse, len(funds))

	for i := 0; i < len(funds); i++ {
		// Get user role in fund
		role, err := a.funds.GetUserRoleInFund(c, uid, funds[i].FundId)
		if err != nil {
			log.Errorf(c, "[funds.FundListHandler] failed to get user role for fund \"id:%d\" and user \"uid:%d\", because %s", funds[i].FundId, uid, err.Error())
			role = models.FUND_ROLE_MEMBER // Default to member if we can't determine role
		}

		// Get member count
		members, err := a.funds.GetFundMembers(c, uid, funds[i].FundId)
		memberCount := 0
		if err == nil {
			memberCount = len(members)
		}

		fundResps[i] = funds[i].ToFundInfoResponse(role, memberCount)
	}

	return fundResps, nil
}

// FundGetHandler returns one specific fund of current user
func (a *FundsApi) FundGetHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()

	// Get fundId from URL context parameter
	fundId, err := GetFundIdFromContext(c, uid)
	if err != nil {
		return nil, err
	}

	fund, fundErr := a.funds.GetFundByFundId(c, uid, fundId)

	if fundErr != nil {
		log.Errorf(c, "[funds.FundGetHandler] failed to get fund \"id:%d\" for user \"uid:%d\", because %s", fundId, uid, fundErr.Error())
		return nil, errs.Or(fundErr, errs.ErrOperationFailed)
	}

	// Get user role in fund
	role, roleErr := a.funds.GetUserRoleInFund(c, uid, fund.FundId)
	if roleErr != nil {
		log.Errorf(c, "[funds.FundGetHandler] failed to get user role for fund \"id:%d\" and user \"uid:%d\", because %s", fund.FundId, uid, roleErr.Error())
		role = models.FUND_ROLE_MEMBER // Default to member if we can't determine role
	}

	// Get member count
	members, membersErr := a.funds.GetFundMembers(c, uid, fund.FundId)
	memberCount := 0
	if membersErr == nil {
		memberCount = len(members)
	}

	return fund.ToFundInfoResponse(role, memberCount), nil
}

// FundCreateHandler saves a new fund by request parameters for current user
func (a *FundsApi) FundCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var fundCreateReq models.FundCreateRequest
	err := c.ShouldBindJSON(&fundCreateReq)

	if err != nil {
		log.Warnf(c, "[funds.FundCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && fundCreateReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_FUND, uid, fundCreateReq.ClientSessionId)

		if found {
			log.Infof(c, "[funds.FundCreateHandler] another fund \"id:%s\" has been created for user \"uid:%d\"", remark, uid)
			fundId, err := utils.StringToInt64(remark)

			if err == nil {
				fund, err := a.funds.GetFundByFundId(c, uid, fundId)

				if err != nil {
					log.Errorf(c, "[funds.FundCreateHandler] failed to get existed fund \"id:%d\" for user \"uid:%d\", because %s", fundId, uid, err.Error())
					return nil, errs.Or(err, errs.ErrOperationFailed)
				}

				// Get user role in fund
				role, err := a.funds.GetUserRoleInFund(c, uid, fund.FundId)
				if err != nil {
					log.Errorf(c, "[funds.FundCreateHandler] failed to get user role for fund \"id:%d\" and user \"uid:%d\", because %s", fund.FundId, uid, err.Error())
					role = models.FUND_ROLE_OWNER // Should be owner since they just created it
				}

				// Get member count (should be 1 for new fund)
				members, err := a.funds.GetFundMembers(c, uid, fund.FundId)
				memberCount := 1
				if err == nil {
					memberCount = len(members)
				}

				return fund.ToFundInfoResponse(role, memberCount), nil
			}
		}
	}

	fund, err := a.funds.CreateFund(c, uid, &fundCreateReq)

	if err != nil {
		log.Errorf(c, "[funds.FundCreateHandler] failed to create fund for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[funds.FundCreateHandler] user \"uid:%d\" has created a new fund \"id:%d\" successfully", uid, fund.FundId)

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_FUND, uid, fundCreateReq.ClientSessionId, utils.Int64ToString(fund.FundId))

	// Get user role in fund
	role, err := a.funds.GetUserRoleInFund(c, uid, fund.FundId)
	if err != nil {
		log.Errorf(c, "[funds.FundCreateHandler] failed to get user role for fund \"id:%d\" and user \"uid:%d\", because %s", fund.FundId, uid, err.Error())
		role = models.FUND_ROLE_OWNER // Should be owner since they just created it
	}

	// Get member count (should be 1 for new fund)
	members, err := a.funds.GetFundMembers(c, uid, fund.FundId)
	memberCount := 1
	if err == nil {
		memberCount = len(members)
	}

	return fund.ToFundInfoResponse(role, memberCount), nil
}

// FundModifyHandler saves an existed fund by request parameters for current user
func (a *FundsApi) FundModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var fundModifyReq models.FundModifyRequest
	err := c.ShouldBindJSON(&fundModifyReq)

	if err != nil {
		log.Warnf(c, "[funds.FundModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	fundId, fundErr := GetFundIdFromContext(c, uid)
	if fundErr != nil {
		return nil, fundErr
	}

	if fundId <= 0 {
		return nil, errs.ErrFundIdInvalid
	}

	fund, err := a.funds.ModifyFundWithRequest(c, uid, fundId, &fundModifyReq)

	if err != nil {
		log.Errorf(c, "[funds.FundModifyHandler] failed to update fund \"id:%d\" for user \"uid:%d\", because %s", fundId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[funds.FundModifyHandler] user \"uid:%d\" has updated fund \"id:%d\" successfully", uid, fundId)

	// Get user role in fund
	role, err := a.funds.GetUserRoleInFund(c, uid, fund.FundId)
	if err != nil {
		log.Errorf(c, "[funds.FundModifyHandler] failed to get user role for fund \"id:%d\" and user \"uid:%d\", because %s", fund.FundId, uid, err.Error())
		role = models.FUND_ROLE_OWNER // Should be owner since they can modify it
	}

	// Get member count
	members, err := a.funds.GetFundMembers(c, uid, fund.FundId)
	memberCount := 0
	if err == nil {
		memberCount = len(members)
	}

	return fund.ToFundInfoResponse(role, memberCount), nil
}

// FundDeleteHandler deletes an existed fund by request parameters for current user
func (a *FundsApi) FundDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()

	// Get fundId from URL context parameter
	fundId, err := GetFundIdFromContext(c, uid)
	if err != nil {
		return nil, err
	}

	deleteErr := a.funds.DeleteFund(c, uid, fundId)

	if deleteErr != nil {
		log.Errorf(c, "[funds.FundDeleteHandler] failed to delete fund \"id:%d\" for user \"uid:%d\", because %s", fundId, uid, deleteErr.Error())
		return nil, errs.Or(deleteErr, errs.ErrOperationFailed)
	}

	log.Infof(c, "[funds.FundDeleteHandler] user \"uid:%d\" has deleted fund \"id:%d\"", uid, fundId)
	return true, nil
}

// FundMemberListHandler returns fund members list for a specific fund
func (a *FundsApi) FundMemberListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()

	// Get fundId from URL context parameter
	fundId, err := GetFundIdFromContext(c, uid)
	if err != nil {
		return nil, err
	}

	members, membersErr := a.funds.GetFundMembers(c, uid, fundId)

	if membersErr != nil {
		log.Errorf(c, "[funds.FundMemberListHandler] failed to get fund members for fund \"id:%d\" and user \"uid:%d\", because %s", fundId, uid, membersErr.Error())
		return nil, errs.Or(membersErr, errs.ErrOperationFailed)
	}

	memberResps := make([]*models.FundMemberResponse, len(members))

	for i := 0; i < len(members); i++ {
		memberResps[i] = members[i].ToFundMemberResponse()
	}

	return memberResps, nil
}

// FundMemberCreateHandler adds a new member to a fund
func (a *FundsApi) FundMemberCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var fundMemberCreateReq models.FundMemberCreateRequest
	err := c.ShouldBindJSON(&fundMemberCreateReq)

	if err != nil {
		log.Warnf(c, "[funds.FundMemberCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	fundId, errFund := GetFundIdFromContext(c, uid)
	if errFund != nil {
		return nil, errFund
	}

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && fundMemberCreateReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_FUND_MEMBER, uid, fundMemberCreateReq.ClientSessionId)

		if found {
			log.Infof(c, "[funds.FundMemberCreateHandler] another fund member \"id:%s\" has been created for user \"uid:%d\"", remark, uid)
			memberId, err := utils.StringToInt64(remark)

			if err == nil {
				member, err := a.funds.GetFundMemberByMemberId(c, uid, memberId)

				if err != nil {
					log.Errorf(c, "[funds.FundMemberCreateHandler] failed to get existed fund member \"id:%d\" for user \"uid:%d\", because %s", memberId, uid, err.Error())
					return nil, errs.Or(err, errs.ErrOperationFailed)
				}

				return member.ToFundMemberResponse(), nil
			}
		}
	}

	member, err := a.funds.AddMember(c, uid, fundId, &fundMemberCreateReq)

	if err != nil {
		log.Errorf(c, "[funds.FundMemberCreateHandler] failed to add member to fund \"id:%d\" for user \"uid:%d\", because %s", fundId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[funds.FundMemberCreateHandler] user \"uid:%d\" has added a new member \"id:%d\" to fund \"id:%d\" successfully", uid, member.MemberId, fundId)

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_FUND_MEMBER, uid, fundMemberCreateReq.ClientSessionId, utils.Int64ToString(member.MemberId))

	return member.ToFundMemberResponse(), nil
}

// FundMemberDeleteHandler removes a member from a fund
func (a *FundsApi) FundMemberDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var fundMemberDeleteReq models.FundMemberDeleteRequest
	err := c.ShouldBindJSON(&fundMemberDeleteReq)

	if err != nil {
		log.Warnf(c, "[funds.FundMemberDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	// Get fundId from URL context parameter
	fundId, errFund := GetFundIdFromContext(c, uid)
	if errFund != nil {
		return nil, errFund
	}

	err = a.funds.RemoveMember(c, uid, fundId, fundMemberDeleteReq.MemberId)

	if err != nil {
		log.Errorf(c, "[funds.FundMemberDeleteHandler] failed to remove member \"id:%d\" from fund \"id:%d\" for user \"uid:%d\", because %s", fundMemberDeleteReq.MemberId, fundId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[funds.FundMemberDeleteHandler] user \"uid:%d\" has removed member \"id:%d\" from fund \"id:%d\"", uid, fundMemberDeleteReq.MemberId, fundId)
	return true, nil
}

// FundMemberLinkHandler links a fund member to an existing user
func (a *FundsApi) FundMemberLinkHandler(c *core.WebContext) (any, *errs.Error) {
	var fundMemberLinkReq models.FundMemberLinkRequest
	err := c.ShouldBindJSON(&fundMemberLinkReq)

	if err != nil {
		log.Warnf(c, "[funds.FundMemberLinkHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	member, err := a.funds.LinkMemberToUser(c, uid, fundMemberLinkReq.MemberId, fundMemberLinkReq.LinkedUid)

	if err != nil {
		log.Errorf(c, "[funds.FundMemberLinkHandler] failed to link member \"id:%d\" to user \"uid:%d\" for user \"uid:%d\", because %s", fundMemberLinkReq.MemberId, fundMemberLinkReq.LinkedUid, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[funds.FundMemberLinkHandler] user \"uid:%d\" has linked member \"id:%d\" to user \"uid:%d\"", uid, fundMemberLinkReq.MemberId, fundMemberLinkReq.LinkedUid)

	return member.ToFundMemberResponse(), nil
}
