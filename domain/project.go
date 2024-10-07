package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
)

var ErrProjectNotFound = errors.New("проект не найден")

type Project struct {
	UUID           uuid.UUID
	FederationUUID uuid.UUID
	CompanyUUID    uuid.UUID
	Name           string `validate:"lte=100,gte=3"  ru:"название"`
	Description    string `validate:"lte=5000"  ru:"описание"`
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
	Meta           datatypes.JSON
	Fields         []CompanyField

	ResponsibleBy string

	StatusGraph *StatusGraph

	Options ProjectOptions

	Users []ProjectUser

	StatusCode      int
	StatusUpdatedAt *time.Time
	StatusSort      []int
	FieldsSort      []string
}

type ProjectOptions struct {
	RequireCancelationComment *bool   `json:"require_cancelation_comment,omitempty"`
	RequireDoneComment        *bool   `json:"require_done_comment,omitempty"`
	StatusEnable              *bool   `json:"status_enable,omitempty"`
	Color                     *string `json:"color,omitempty"`
}

type ProjectParams struct {
	Status        *int      `json:"status,omitempty"`
	StatusSort    *[]int    `json:"status_sort,omitempty"`
	FieldsSort    *[]string `json:"fields_sort,omitempty"`
	ResponsibleBy *string   `json:"responsible_by,omitempty"`
}

func (j *ProjectOptions) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := ProjectOptions{
		Color:                     helpers.Ptr("#111111"),
		RequireCancelationComment: helpers.Ptr(false),
		RequireDoneComment:        helpers.Ptr(false),
		StatusEnable:              helpers.Ptr(false),
	}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

func (j ProjectOptions) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func NewProject(name, description string, federationUUID, companyUUID uuid.UUID, createdBy, responsibleBy string) *Project {
	p := &Project{
		UUID:           uuid.New(),
		Name:           name,
		Description:    description,
		FederationUUID: federationUUID,
		CompanyUUID:    companyUUID,
		CreatedBy:      createdBy,
		ResponsibleBy:  responsibleBy,
	}

	errs, ok := helpers.ValidationStruct(p)
	if !ok {
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return p
}

func NewProjectUUID(uid uuid.UUID) *Project {
	return &Project{
		UUID: uid,
	}
}

func (p *Project) ChangeName(name string) error {
	if len(name) < 3 || len(name) > 100 {
		return errors.New("название проекта от 3 до 100 символов")
	}

	p.Name = name

	return nil
}

func (p *Project) ChangeDescription(description string) error {
	if len(description) > 5000 {
		return errors.New("описание проекта до 5000 символов")
	}

	p.Description = description

	return nil
}

type FieldDataType int

const (
	Integer   FieldDataType = 0
	Float     FieldDataType = 1
	String    FieldDataType = 2
	Text      FieldDataType = 3
	Bool      FieldDataType = 4
	Switch    FieldDataType = 5
	Array     FieldDataType = 6
	Data      FieldDataType = 7
	DataArray FieldDataType = 8
	Phone     FieldDataType = 9
	Link      FieldDataType = 10
	Email     FieldDataType = 11
	Time      FieldDataType = 12
	DateTime  FieldDataType = 13
	People    FieldDataType = 14
)

type ProjectCatalogType string

const (
	Reasons ProjectCatalogType = "reasons"
)

type ProjectCatalogData struct {
	UUID           uuid.UUID
	FederationUUID uuid.UUID
	CompanyUUID    uuid.UUID
	ProjectUUID    uuid.UUID

	Name  ProjectCatalogType
	Value string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type CompanyField struct {
	UUID               uuid.UUID
	Hash               string
	Name               string        `validate:"lte=30,gte=1" ru:"название"`
	Description        string        `validate:"lte=5000" ru:"описание"`
	Icon               string        `validate:"lte=50" ru:"иконка"`
	DataType           FieldDataType `validate:"lte=10,gte=0" ru:"тип данных"`
	CompanyUUID        uuid.UUID     `validate:"uuid" ru:"компания uuid"`
	RequiredOnStatuses []int         `validate:"lte=50" ru:"необходимо на статусе"`
	Style              string        `validate:"lte=20" ru:"стиль"`
	CreatedBy          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
	Meta               datatypes.JSON
	ProjectUUID        []uuid.UUID `validate:"uuid" ru:"проект uuid"`

	TasksTotal        int
	TasksFilled       int
	TasksActiveTotal  int
	TasksActiveFilled int
}

func (p *Project) AddFiled(name, description string, dataType FieldDataType) {
	f := &CompanyField{
		Name:        name,
		DataType:    dataType,
		Description: description,
	}

	p.Fields = append(p.Fields, *f)
}

func (pf *CompanyField) FieldTypeDesc() string {
	switch pf.DataType {
	case Bool:
		return "bool"
	case Float:
		return "float"
	case Integer:
		return "integer"
	case String:
		return "string"
	case Text:
		return "text"
	case Switch:
		return "switch"
	case Array:
		return "array"
	case Data:
		return "data"
	case DataArray:
		return "data_array"
	case Phone:
		return "phone"
	case Link:
		return "link"
	case Email:
		return "email"
	case Time:
		return "time"
	case DateTime:
		return "datetime"
	case People:
		return "people"
	}

	return "unknown"
}

type ProjectUser struct {
	UUID           uuid.UUID `validate:"uuid"`
	User           User
	FederationUUID uuid.UUID `validate:"uuid"`
	CompanyUUID    uuid.UUID `validate:"uuid"`
	ProjectUUID    uuid.UUID `validate:"uuid"`
	AddedAt        time.Time `json:"added_at"`
}

func NewProjectUser(federationUUID, companyUUID, projectUUID, userUUID uuid.UUID) *ProjectUser {
	pu := &ProjectUser{
		UUID:           uuid.New(),
		User:           User{UUID: userUUID},
		FederationUUID: federationUUID,
		CompanyUUID:    companyUUID,
		ProjectUUID:    projectUUID,
		AddedAt:        time.Now(),
	}

	errs, ok := helpers.ValidationStruct(pu, "UUID", "FederationUUID", "ProjectUUID", "User.UUID")
	if !ok {
		logrus.Error(errors.New(helpers.Join(errs, ", ")))
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return pu
}

type ProjectStatus struct {
	CompanyUUID uuid.UUID  `validate:"uuid"`
	ProjectUUID uuid.UUID  `validate:"uuid"`
	UUID        *uuid.UUID `json:"uuid,omitempty"`
	Name        string     `json:"name"`
	Number      int        `json:"priority"`
	Color       string     `json:"color"`
	Description string     `json:"description"`
	Edit        []string   `json:"edit"`
}
