package sms

import "github.com/krisch/crm-backend/domain"

type ISMS interface {
	SmsSend(p *domain.Sms) (Response, error)
	SmsStatus(id string) (Response, error)
	SmsCost(p *domain.Sms) (Response, error)
	MyBalance() (Response, error)
}
