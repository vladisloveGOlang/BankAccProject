package web

import (
	"context"
	"errors"
	"fmt"

	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/company"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/jwt"
	"github.com/krisch/crm-backend/internal/sms"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
	"github.com/samber/lo"
)

// PostCompanySmsCost implements oapi.StrictServerInterface.
func (a *Web) PostCompanyUUIDSmsCost(ctx context.Context, request oapi.PostCompanyUUIDSmsCostRequestObject) (oapi.PostCompanyUUIDSmsCostResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	smsOptions, err := a.app.CompanyService.GetSmsOptions(request.UUID)
	if err != nil {
		return nil, err
	}

	s := &domain.Sms{
		To:   fmt.Sprint(request.Body.Phone),
		Text: request.Body.Text,
		From: smsOptions.From,
	}

	rsp, err := a.app.SMSService.SmsCost(smsOptions.API, s)
	if err != nil {
		return nil, err
	}

	mp, err := helpers.StructToMap(rsp)
	if err != nil {
		return nil, err
	}

	return oapi.PostCompanyUUIDSmsCost200JSONResponse(mp), nil
}

func (a *Web) PostCompanyUUIDSmsOptions(ctx context.Context, request oapi.PostCompanyUUIDSmsOptionsRequestObject) (oapi.PostCompanyUUIDSmsOptionsResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CompanyService.CreateSmsOptions(request.UUID, company.SmsOptions{
		API:  request.Body.Api,
		From: request.Body.From,
	})
	if err != nil {
		return nil, err
	}

	return oapi.PostCompanyUUIDSmsOptions200Response{}, nil
}

var MockSms = "mock_sms"

func (a *Web) PostCompanyUUIDSmsSend(ctx context.Context, request oapi.PostCompanyUUIDSmsSendRequestObject) (oapi.PostCompanyUUIDSmsSendResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	mockSms := false
	if request.Params.MockSms != nil && *request.Params.MockSms == "true" {
		mockSms = true
	}

	cmpny, f := a.app.DictionaryService.FindCompany(request.UUID)
	if !f {
		return nil, errors.New("company not found")
	}

	smsOptions, err := a.app.CompanyService.GetSmsOptions(cmpny.UUID)
	if err != nil {
		return nil, err
	}

	s := sms.NewCompanySms(fmt.Sprint(request.Body.Phone), request.Body.Text, smsOptions.From, claims.UUID, claims.Email, cmpny)

	mp := make(map[string]interface{})

	if !mockSms {
		rsp, err := a.app.SMSService.SmsSend(smsOptions.API, s)
		if err != nil {
			return nil, err
		}
		mp, err = helpers.StructToMap(rsp)
		if err != nil {
			return nil, err
		}

		err = a.app.SMSService.StoreSms(s)
		if err != nil {
			return nil, err
		}
	} else {
		err := a.app.SMSService.StoreSms(s)
		if err != nil {
			return nil, err
		}

		mp["balance"] = "0"
		mp["callbacks"] = nil
		mp["cost"] = 0
		mp["count"] = 0
		mp["id"] = []string{"202419-1000003"}
		mp["limit"] = 0
		mp["limit_sent"] = 0
		mp["senders"] = nil
		mp["status"] = 100
		mp["stoplist"] = nil
	}

	return oapi.PostCompanyUUIDSmsSend200JSONResponse(mp), nil
}

func (a *Web) GetCompanyUUIDSms(ctx context.Context, request oapi.GetCompanyUUIDSmsRequestObject) (oapi.GetCompanyUUIDSmsResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	filter := dto.SmsFilterDTO{
		Offset:  request.Params.Offset,
		Limit:   request.Params.Limit,
		IsMy:    request.Params.IsMy,
		MyEmail: &claims.Email,
	}

	dms, total, err := a.app.SMSService.GetSms(ctx, filter)
	if err != nil {
		return nil, err
	}

	return oapi.GetCompanyUUIDSms200JSONResponse{
		Count: len(dms),
		Items: lo.Map(dms, func(item domain.Sms, i int) dto.SmsDTO {
			return dto.SmsDTO{
				UUID:           item.UUID,
				FederationUUID: item.FederationUUID,
				CompanyUUID:    item.CompanyUUID,
				UserUUID:       item.CreatedByUUID,
				Phone:          item.To,
				Text:           item.Text,
				Status:         "sent",
				CreatedAt:      item.CreatedAt,
				UpdatedAt:      item.UpdatedAt,
			}
		}),
		Total: total,
	}, nil
}
