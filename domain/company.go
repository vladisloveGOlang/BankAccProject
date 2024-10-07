package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
)

type Company struct {
	UUID           uuid.UUID
	FederationUUID uuid.UUID
	Name           string `validate:"gte=3,lte=100"  ru:"название"`

	CreatedBy     string
	CreatedByUUID uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	Meta          datatypes.JSON

	UserTotal int
	Users     []CompanyUser

	Fields        []CompanyField
	FieldLastName string

	ProjectsTotal int
	Projects      []Project
}

func NewCompany(name string, federationUUID uuid.UUID, createdBy string, ceatedByUUID uuid.UUID) *Company {
	return &Company{
		UUID:           uuid.New(),
		Name:           name,
		FederationUUID: federationUUID,
		CreatedBy:      createdBy,
		CreatedByUUID:  ceatedByUUID,
	}
}

func NewCompanyByUUID(uid uuid.UUID) *Company {
	return &Company{
		UUID: uid,
	}
}

func (c *Company) ChangeName(name string) error {
	if len(name) < 3 || len(name) > 100 {
		return errors.New("название от 3 до 100 символов")
	}

	c.Name = name

	return nil
}

type CompanyUser struct {
	UUID           uuid.UUID `validate:"uuid"`
	User           User
	FederationUUID uuid.UUID `validate:"uuid"`
	CompanyUUID    uuid.UUID `validate:"uuid"`
	AddedAt        time.Time `json:"added_at"`
}

func NewCompanyUser(federationUUID, companyUUID, userUUID uuid.UUID) *CompanyUser {
	fu := &CompanyUser{
		UUID:           uuid.New(),
		User:           User{UUID: userUUID},
		FederationUUID: federationUUID,
		CompanyUUID:    companyUUID,
		AddedAt:        time.Now(),
	}

	errs, ok := helpers.ValidationStruct(fu, "UUID", "FederationUUID", "User.UUID")
	if !ok {
		logrus.Error(errors.New(helpers.Join(errs, ", ")))
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return fu
}

type CompanyPriority struct {
	CompanyUUID uuid.UUID `validate:"uuid"`
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Number      int       `json:"priority"`
	Color       string    `json:"color"`
}
