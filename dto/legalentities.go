package dto

import (
	"time"

	"github.com/google/uuid"
)

type LegalEntityDTO struct {
	UUID                         uuid.UUID  `json:"uuid"`
	FederationUUID               uuid.UUID  `json:"federation_uuid"`
	CompanyUUID                  *uuid.UUID `json:"company_uuid"`
	EntityType                   string     `json:"entity_type"`
	Logo                         string     `json:"logo"`
	Phone                        string     `json:"phone"`
	Fax                          string     `json:"fax"`
	Email                        string     `json:"email"`
	IsVatPayer                   bool       `json:"is_vat_payer"`
	FullName                     string     `json:"full_name"`
	ShortName                    string     `json:"short_name"`
	INN                          string     `json:"inn"`
	KPP                          string     `json:"kpp"`
	OGRN                         string     `json:"ogrn"`
	OKPO                         string     `json:"okpo"`
	LegalAddress                 string     `json:"legal_address"`
	LegalPostalCode              string     `json:"legal_postal_code"`
	LegalCountry                 string     `json:"legal_country"`
	LegalRegion                  string     `json:"legal_region"`
	LegalCity                    string     `json:"legal_city"`
	LegalStreet                  string     `json:"legal_street"`
	LegalHouse                   string     `json:"legal_house"`
	LegalApartment               string     `json:"legal_apartment"`
	LegalComments                string     `json:"legal_comments"`
	ActualAddressSameAsLegal     bool       `json:"actual_address_same_as_legal"`
	ActualPostalCode             string     `json:"actual_postal_code"`
	ActualCountry                string     `json:"actual_country"`
	ActualRegion                 string     `json:"actual_region"`
	ActualCity                   string     `json:"actual_city"`
	ActualStreet                 string     `json:"actual_street"`
	ActualHouse                  string     `json:"actual_house"`
	ActualApartment              string     `json:"actual_apartment"`
	ActualComments               string     `json:"actual_comments"`
	DirectorLastName             string     `json:"director_last_name"`
	DirectorFirstName            string     `json:"director_first_name"`
	DirectorMiddleName           string     `json:"director_middle_name"`
	DirectorLastNameGenitive     string     `json:"director_last_name_genitive"`
	DirectorFirstNameGenitive    string     `json:"director_first_name_genitive"`
	DirectorMiddleNameGenitive   string     `json:"director_middle_name_genitive"`
	DirectorPosition             string     `json:"director_position"`
	DirectorSignature            string     `json:"director_signature"`
	DirectorSeal                 string     `json:"director_seal"`
	AccountantLastName           string     `json:"accountant_last_name"`
	AccountantFirstName          string     `json:"accountant_first_name"`
	AccountantMiddleName         string     `json:"accountant_middle_name"`
	AccountantLastNameGenitive   string     `json:"accountant_last_name_genitive"`
	AccountantFirstNameGenitive  string     `json:"accountant_first_name_genitive"`
	AccountantMiddleNameGenitive string     `json:"accountant_middle_name_genitive"`
	Comments                     string     `json:"comments"`
	Status                       bool       `json:"status"`
	CreatedBy                    string     `json:"created_by"`
	CreatedByUUID                uuid.UUID  `json:"created_by_uuid"`
	CreatedAt                    time.Time  `json:"created_at"`
	UpdatedAt                    time.Time  `json:"updated_at"`
	DeletedAt                    *time.Time `json:"deleted_at"`
}
