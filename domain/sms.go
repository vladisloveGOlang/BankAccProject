package domain

import (
	"time"

	"github.com/google/uuid"
)

type Sms struct {
	UUID           uuid.UUID
	FederationUUID uuid.UUID
	CompanyUUID    uuid.UUID
	CreatedBy      string
	CreatedByUUID  uuid.UUID

	To        string            `json:"to"`
	Text      string            `json:"text"`
	Translit  bool              `json:"translit"`
	Multi     map[string]string `json:"multi"`
	From      string            `json:"from"`
	Time      time.Time         `json:"time"`
	Test      bool              `json:"test"`
	PartnerID int               `json:"partner_id"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewSms(to, text, from string) *Sms {
	return &Sms{
		UUID: uuid.New(),
		To:   to,
		Text: text,
		From: from,
	}
}
