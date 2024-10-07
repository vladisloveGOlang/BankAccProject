package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
)

type CatalogDTO struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`

	Federation FederationDTOs `json:"federation"`
	Company    CompanyDTOs    `json:"company"`

	Fields      []CatalogFieldDTO `json:"fields"`
	FieldsTotal int               `json:"fields_total"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AllowSort []string `json:"allow_sort"`
}

type CatalogDTOs struct {
	UUID           uuid.UUID `json:"uuid"`
	Name           string    `json:"name"`
	FederationUUID uuid.UUID `json:"federation_uuid"`
	CompanyUUID    uuid.UUID `json:"company_uuid"`
}

type CatalogFieldDTO struct {
	UUID            uuid.UUID  `json:"uuid"`
	Name            string     `json:"name"`
	Hash            string     `json:"hash"`
	DataType        int        `json:"data_type"`
	DataCatalogUUID *uuid.UUID `json:"data_catalog_uuid,omitempty"`
	DataDesc        string     `json:"data_desc"`
}

type CatalogDataDTO struct {
	UUID uuid.UUID `json:"uuid"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	Fields []CatalogDataFieldDTO `json:"fields"`
}

type CatalogDataFieldDTO struct {
	Hash     string      `json:"hash"`
	Name     string      `json:"name"`
	DataType int         `json:"data_type"`
	Value    interface{} `json:"value"`
}

func NewCatalogDataDTO(dm domain.CatalogData, dict IDict) CatalogDataDTO {
	catalogFields, _ := dict.FindCatalogFields(dm.CatalogUUID)

	taskFields := []CatalogDataFieldDTO{}

	if dm.Fields != nil {
		fi := dm.Fields
		for _, pf := range catalogFields {
			taskFields = append(taskFields, CatalogDataFieldDTO{
				Hash:     pf.Hash,
				Name:     pf.Name,
				DataType: pf.DataType,
				Value:    fi[pf.Hash],
			})
		}
	}

	return CatalogDataDTO{
		UUID: dm.UUID,

		Fields: taskFields,

		CreatedAt: dm.CreatedAt,
		UpdatedAt: dm.UpdatedAt,
		DeletedAt: dm.DeletedAt,
	}
}

type CatalogSearchDTO struct {
	FederationUUID uuid.UUID `json:"federation_uuid"`
	CompanyUUID    uuid.UUID `json:"company_uuid"`
	CatalogUUID    uuid.UUID `json:"catalog_uuid"`
	Offset         *int      `json:"offset"`
	Limit          *int      `json:"limit"`

	Fields []FilterDTO `json:"fields"`

	Order *string `json:"order"`
	By    *string `json:"by"`
}

func (d *CatalogSearchDTO) Validate() error {
	return nil
}
