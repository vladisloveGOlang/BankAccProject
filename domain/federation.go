package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
	"gorm.io/datatypes"
)

type Federation struct {
	UUID          uuid.UUID `validate:"uuid"`
	Name          string    `validate:"lte=100,gte=1"  ru:"название"`
	CreatedBy     string    `validate:"email"`
	CreatedByUUID uuid.UUID `validate:"uuid"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	Meta          datatypes.JSON

	Users      []FederationUser
	UsersTotal int

	Companies      []Company
	CompaniesTotal int
}

type FederationUser struct {
	UUID           uuid.UUID `validate:"uuid"`
	User           User
	FederationUUID uuid.UUID `validate:"uuid"`
	AddedAt        time.Time `json:"added_at"`
}

func NewFederationUser(federationUUID, userUUID uuid.UUID) *FederationUser {
	fu := &FederationUser{
		UUID:           uuid.New(),
		User:           User{UUID: userUUID},
		FederationUUID: federationUUID,
	}

	errs, ok := helpers.ValidationStruct(fu, "UUID", "FederationUUID", "User.UUID")
	if !ok {
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return fu
}

func NewFederation(name, createdBy string, createdByUUID uuid.UUID) *Federation {
	uid := uuid.New()

	federation := &Federation{
		UUID:          uid,
		Name:          name,
		CreatedBy:     createdBy,
		CreatedByUUID: createdByUUID,
		Meta:          datatypes.JSON("{}"),
		Users:         []FederationUser{*NewFederationUser(uid, createdByUUID)},
	}

	errs, ok := helpers.ValidationStruct(federation)
	if !ok {
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return federation
}

func NewFederationUUID(uid uuid.UUID) *Federation {
	federation := &Federation{
		UUID: uid,
	}

	return federation
}

func (f *Federation) ChangeName(name string) error {
	if len(name) < 1 || len(name) > 100 {
		return errors.New("название от 1 до 100 символов")
	}

	f.Name = name

	return nil
}
