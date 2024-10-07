package catalogs

import (
	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
)

const (
	UsersReservedCatalogName     = "users"
	CompaniesReservedCatalogName = "companies"
	GoodsReservedCatalogName     = "goods"
)

func GetReservedNames() []string {
	return []string{
		UsersReservedCatalogName,
		CompaniesReservedCatalogName,
		GoodsReservedCatalogName,
	}
}

func (s *Service) CreateUsersCatalog(companyCatalogUUID, federationUUID, commpanyUUID, createdByUUID uuid.UUID, createdBy string) (uid uuid.UUID, err error) {
	catalog := domain.NewCatalog(UsersReservedCatalogName, federationUUID, commpanyUUID, createdBy, createdByUUID)
	uid = catalog.UUID

	err = s.CreateCatalog(catalog)
	if err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"name",
		"name",
		domain.String,
		nil,
		catalog.UUID,
		createdBy,
	)); err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"lname",
		"lname",
		domain.String,
		nil,
		catalog.UUID,
		createdBy,
	)); err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"pname",
		"pname",
		domain.String,
		nil,
		catalog.UUID,
		createdBy,
	)); err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"phone",
		"phone",
		domain.String,
		nil,
		catalog.UUID,
		createdBy,
	)); err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"favorite",
		"favorite",
		domain.Bool,
		nil,
		catalog.UUID,
		createdBy,
	)); err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"company",
		"company_uuid",
		domain.Data,
		&companyCatalogUUID,
		catalog.UUID,
		createdBy,
	)); err != nil {
		return uid, err
	}

	return uid, err
}

func (s *Service) CreateCompanyCatalog(federationUUID, commpanyUUID, createdByUUID uuid.UUID, createdBy string) (uid uuid.UUID, err error) {
	catalog := domain.NewCatalog(CompaniesReservedCatalogName, federationUUID, commpanyUUID, createdBy, createdByUUID)
	uid = catalog.UUID

	err = s.CreateCatalog(catalog)
	if err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"name",
		"name",
		domain.String,
		nil,
		catalog.UUID,
		"nightsong@oviovi.site",
	)); err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"inn",
		"inn",
		domain.Integer,
		nil,
		catalog.UUID,
		"nightsong@oviovi.site",
	)); err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"address",
		"address",
		domain.String,
		nil,
		catalog.UUID,
		"nightsong@oviovi.site",
	)); err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"favorite",
		"favorite",
		domain.Bool,
		nil,
		catalog.UUID,
		"nightsong@oviovi.site",
	)); err != nil {
		return uid, err
	}

	return uid, err
}

func (s *Service) CreateGoodsCatalog(federationUUID, commpanyUUID, createdByUUID uuid.UUID, createdBy string) (uid uuid.UUID, err error) {
	catalog := domain.NewCatalog(GoodsReservedCatalogName, federationUUID, commpanyUUID, createdBy, createdByUUID)
	uid = catalog.UUID

	err = s.CreateCatalog(catalog)
	if err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"name",
		"name",
		domain.String,
		nil,
		catalog.UUID,
		"nightsong@oviovi.site",
	)); err != nil {
		return uid, err
	}

	if _, err = s.CreateCatalogField(domain.NewCatalogFiled(
		"favorite",
		"favorite",
		domain.Bool,
		nil,
		catalog.UUID,
		"nightsong@oviovi.site",
	)); err != nil {
		return uid, err
	}

	return uid, err
}
