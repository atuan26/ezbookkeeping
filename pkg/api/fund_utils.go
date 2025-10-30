package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// GetFundIdFromContext extracts fundId from URL parameter or returns user's default fund
// This is a common helper function that can be used across all API handlers
func GetFundIdFromContext(c *core.WebContext, uid int64) (int64, *errs.Error) {
	// First try to get fundId from URL parameter
	fundId, err := c.GetCurrentFundId()
	if err != nil {
		return 0, err
	}

	// If fundId is provided in URL, validate user has access to it
	if fundId > 0 {
		_, errFund := services.Funds.GetFundByFundId(c, uid, fundId)
		if errFund != nil {
			log.Errorf(c, "[GetFundIdFromContext] user uid:%d does not have access to fund id:%d, because %s", uid, fundId, errFund.Error())
			return 0, errs.Or(errFund, errs.ErrFundAccessDenied)
		}
		return fundId, nil
	}

	// If no fundId in URL, get user's default personal fund
	funds, errFunds := services.Funds.GetUserFunds(c, uid)
	if errFunds != nil {
		log.Errorf(c, "[GetFundIdFromContext] failed to get user funds for uid:%d, because %s", uid, errFunds.Error())
		return 0, errs.Or(errFunds, errs.ErrOperationFailed)
	}

	if len(funds) == 0 {
		log.Errorf(c, "[GetFundIdFromContext] no funds found for user uid:%d", uid)
		return 0, errs.ErrFundNotFound
	}

	// Return the first fund (should be personal fund)
	return funds[0].FundId, nil
}
