package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
	"gorm.io/datatypes"
)

var ErrCatalogNotFound = errors.New("справочник не найден")

type Catalog struct {
	UUID           uuid.UUID
	FederationUUID uuid.UUID
	CompanyUUID    uuid.UUID
	Name           string `validate:"lte=100,gte=3"  ru:"название"`

	CreatedBy     string
	CreatedByUUID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Meta datatypes.JSON

	Fields []CatalogFiled
}

func NewCatalog(name string, federationUUID, companyUUID uuid.UUID, createdBy string, createdByUUID uuid.UUID) *Catalog {
	p := &Catalog{
		UUID:           uuid.New(),
		Name:           name,
		FederationUUID: federationUUID,
		CompanyUUID:    companyUUID,
		CreatedBy:      createdBy,
		CreatedByUUID:  createdByUUID,
	}

	errs, ok := helpers.ValidationStruct(p)
	if !ok {
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return p
}

func NewcatalogUUID(uid uuid.UUID) *Catalog {
	return &Catalog{
		UUID: uid,
	}
}

func (p *Catalog) ChangeName(name string) error {
	if len(name) < 3 || len(name) > 100 {
		return errors.New("название справочника от 3 до 100 символов")
	}

	p.Name = name

	return nil
}

type CatalogFiled struct {
	UUID uuid.UUID
	Hash string
	Name string `validate:"lte=30,gte=1"  ru:"название"`

	DataType        FieldDataType `validate:"lte=10,gte=0"  ru:"тип данных"`
	DataCatalogUUID *uuid.UUID

	CatalogUUID uuid.UUID `validate:"uuid"  ru:"project uuid"`

	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Meta      datatypes.JSON
}

func NewCatalogFiled(name, hash string, dataType FieldDataType, dataCatalogUUID *uuid.UUID, catalogUUID uuid.UUID, createdBy string) *CatalogFiled {
	if hash != "" {
		rgxp := regexp.MustCompile(`^[a-z_]{3,20}$`)
		if !rgxp.MatchString(hash) {
			panic(errors.New("hash должен состоять из латинских букв, и(или) знака подчеркивания "))
		}
	}

	if (dataType == Data || dataType == DataArray) && dataCatalogUUID == nil {
		panic(errors.New("для типа данных data и data_array обязательно указывать data_uuid"))
	}

	if dataCatalogUUID != nil && *dataCatalogUUID == catalogUUID {
		panic(errors.New("data_uuid не может быть равен catalog_uuid"))
	}

	p := &CatalogFiled{
		UUID:            uuid.New(),
		Name:            name,
		Hash:            hash,
		DataType:        dataType,
		DataCatalogUUID: dataCatalogUUID,
		CatalogUUID:     catalogUUID,
		CreatedBy:       createdBy,
	}

	errs, ok := helpers.ValidationStruct(p)
	if !ok {
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return p
}

type CatalogData struct {
	UUID           uuid.UUID
	FederationUUID uuid.UUID
	CompanyUUID    uuid.UUID
	CatalogUUID    uuid.UUID

	CreatedBy     string
	CreatedByUUID uuid.UUID

	RawFields map[string]interface{}
	Fields    map[string]interface{}

	Entities     map[string]interface{}
	EntitiesRich []any

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (p *Catalog) AddFiled(name string, dataType FieldDataType) {
	f := &CatalogFiled{
		Name:     name,
		DataType: dataType,
	}

	p.Fields = append(p.Fields, *f)
}

func (pf *CatalogFiled) FieldTypeDesc() string {
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
	}

	return "unknown"
}

type UUID struct {
	UUID uuid.UUID `json:"uuid"`
}
