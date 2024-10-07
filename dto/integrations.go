package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserEmailDTO struct {
	UUID          uuid.UUID  `json:"uuid"`
	Email         string     `json:"email"`
	Password      string     `json:"password"`
	Domain        string     `json:"domain"`
	LastFetchedAt *time.Time `json:"last_fetched_at,omitempty"`
}

type CreateUserDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Domain   string `json:"domain" binding:"required"`
}

type EmailDTO struct {
	ID         uint      `json:"id"`
	UserUUID   uuid.UUID `json:"user_uuid"`
	Email      string    `json:"email"`
	Sender     string    `json:"sender"`
	Subject    string    `json:"subject"`
	BodyText   string    `json:"body_text"`
	ReceivedAt time.Time `json:"received_at"`
	BodyHash   string    `json:"body_hash"`
}

type SendEmailDTO struct {
	SenderEmail    string `json:"sender_email" validate:"email,required"`
	Password       string `json:"password" validate:"required"`
	RecipientEmail string `json:"recipient_email" validate:"email,required"`
	Subject        string `json:"subject" validate:"required"`
	Body           string `json:"body" validate:"required"`
}

type UserEmailDTOSettings struct {
	Domain *string `json:"domain,omitempty"`
	Status *int    `json:"status,omitempty"`
}

type SentEmailDTO struct {
	SenderEmailUUID uuid.UUID `json:"sender_email_uuid"`
	SenderEmail     string    `json:"sender_email"`
	RecipientEmail  string    `json:"recipient_email"`
	Subject         string    `json:"subject"`
	Body            string    `json:"body"`
	CreatedAt       time.Time `json:"created_at"`
}
