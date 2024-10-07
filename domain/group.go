package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Group struct {
	UUID           uuid.UUID
	FederationUUID uuid.UUID
	CompanyUUID    uuid.UUID
	Name           string `validate:"gte=1,lte=100"  ru:"название"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	UsersUUIDS []string
	Users      []User
}

func NewGroup(name string, federationUUID, companyUUID uuid.UUID) *Group {
	return &Group{
		UUID:           uuid.New(),
		Name:           name,
		FederationUUID: federationUUID,
		CompanyUUID:    companyUUID,
	}
}

func NewGroupByUUID(uid uuid.UUID) *Group {
	return &Group{
		UUID: uid,
	}
}

func (g *Group) ChangeName(name string) error {
	if len(name) < 1 || len(name) > 100 {
		return errors.New("название должно быть от 1 до 100 символов")
	}

	g.Name = name
	return nil
}
