package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UUID    uuid.UUID `json:"uuid"`
	Email   string    `json:"email"`
	Name    string    `json:"name"`
	IsValid bool      `json:"is_valid"`
	jwt.RegisteredClaims
}

func (c *Claims) IsRefresh() bool {
	return c.Subject == "refresh"
}

func (c *Claims) GetEmail() string {
	return c.Email
}

func (c *Claims) GetUUID() uuid.UUID {
	return c.UUID
}
