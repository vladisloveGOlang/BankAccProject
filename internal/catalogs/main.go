package catalogs

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/dictionary"
	"github.com/krisch/crm-backend/internal/helpers"
)

type Service struct {
	repo *Repository
	dict *dictionary.Service
}

func New(repo *Repository, dict *dictionary.Service) *Service {
	return &Service{
		repo: repo,
		dict: dict,
	}
}

func (s *Service) CreateCatalog(c *domain.Catalog) (err error) {
	err = s.repo.CreateCatalog(c)

	return err
}

func (s *Service) DeleteCatalog(uid uuid.UUID) (err error) {
	return s.repo.DeleteCatalog(uid)
}

func (s *Service) GetCatalog(uid uuid.UUID) (dir domain.Catalog, err error) {
	orm, err := s.repo.GetCatalog(uid)
	if err != nil {
		return dir, err
	}

	fileds, err := s.GetCatalogFields(orm.UUID)
	if err != nil {
		return dir, err
	}

	dir = domain.Catalog{
		UUID:           orm.UUID,
		Name:           orm.Name,
		FederationUUID: orm.FederationUUID,
		CompanyUUID:    orm.CompanyUUID,
		CreatedBy:      orm.CreatedBy,
		CreatedByUUID:  orm.CreatedByUUID,
		Fields:         fileds,
		Meta:           orm.Meta,

		CreatedAt: orm.CreatedAt,
		UpdatedAt: orm.UpdatedAt,
	}

	return dir, err
}

func (s *Service) GetCatalogsByCompany(companyUUID uuid.UUID) (dmns []domain.Catalog, err error) {
	orms, err := s.repo.GetCatalogsByCompany(companyUUID)
	if err != nil {
		return dmns, err
	}

	for _, orm := range orms {
		fileds, err := s.GetCatalogFields(orm.UUID)
		if err != nil {
			return dmns, err
		}

		dmns = append(dmns, domain.Catalog{
			UUID:           orm.UUID,
			FederationUUID: orm.FederationUUID,
			CompanyUUID:    orm.CompanyUUID,
			Name:           orm.Name,
			Fields:         fileds,
			Meta:           orm.Meta,
			CreatedAt:      orm.CreatedAt,
			UpdatedAt:      orm.UpdatedAt,
		})
	}

	return dmns, err
}

func (s *Service) CreateCatalogField(pf *domain.CatalogFiled) (df domain.CatalogFiled, err error) {
	orm, err := s.repo.CreateCatalogField(pf)
	if err != nil {
		return df, err
	}

	return domain.CatalogFiled{
		UUID:            orm.UUID,
		CatalogUUID:     orm.CatalogUUID,
		Name:            orm.Name,
		DataType:        domain.FieldDataType(orm.DataType),
		DataCatalogUUID: orm.DataCatalogUUID,
		Hash:            orm.Hash,
	}, err
}

func (s *Service) PutCatalogField(pf *domain.CatalogFiled) error {
	return s.repo.PutCatalogField(pf)
}

func (s *Service) GetCatalogFields(uid uuid.UUID) (items []domain.CatalogFiled, err error) {
	items, err = s.repo.GetCatalogFields(uid)
	if err != nil {
		return items, err
	}

	return items, err
}

func (s *Service) DeleteCatalogField(uid uuid.UUID) (err error) {
	return s.repo.DeletecatalogField(uid)
}

func (s *Service) ChangeCatalogName(uid uuid.UUID, name string) (err error) {
	p := domain.NewcatalogUUID(uid)
	err = p.ChangeName(name)
	if err != nil {
		return err
	}

	err = s.repo.ChangecatalogField(p.UUID, "name", p.Name)

	return err
}

func (s *Service) AddData(dm domain.CatalogData) (dd CatalogData, err error) {
	err = s.FilterCatalogFields(&dm)
	if err != nil {
		return dd, err
	}

	dd, err = s.repo.AddData(dm)

	return dd, err
}

func (s *Service) GetSortFields(catalogUUID uuid.UUID) []string {
	fields, _ := s.dict.FindCatalogFields(catalogUUID)

	allowOrder := s.repo.GetSortFields()

	for _, field := range fields {
		if domain.FieldDataType(field.DataType) == domain.Integer || domain.FieldDataType(field.DataType) == domain.Float || domain.FieldDataType(field.DataType) == domain.String {
			// fileds.a || fields.b
			allowOrder = append(allowOrder, "fields."+field.Hash+"")
		}
	}

	return allowOrder
}

func (s *Service) GetData(search dto.CatalogSearchDTO) (dmns []domain.CatalogData, total int64, err error) {
	allowOrder := s.GetSortFields(search.CatalogUUID)

	return s.repo.GetData(search, allowOrder)
}

func (s *Service) FilterCatalogFields(catalogData *domain.CatalogData) (err error) {
	if len(catalogData.RawFields) > 0 {
		projectFields, _ := s.dict.FindCatalogFields(catalogData.CatalogUUID)

		filteredFields := make(map[string]interface{}, 0)

		addedFieldsHash := []string{}
		entities := make(map[string]interface{})
		for _, pfield := range projectFields {
			if value, ok := catalogData.RawFields[pfield.Hash]; ok {
				addedFieldsHash = append(addedFieldsHash, pfield.Hash)

				switch domain.FieldDataType(pfield.DataType) {
				case domain.Integer:
					if v, ok := value.(int); ok {
						filteredFields[pfield.Hash] = v
						continue
					}

					if v, ok := value.(float64); ok {
						filteredFields[pfield.Hash] = int(v)
					} else {
						msg := fmt.Sprintf("field %s (%s) should be integer", pfield.Name, pfield.Hash)
						return errors.New(msg)
					}
				case domain.Float:
					if v, ok := value.(float64); ok {
						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be float", pfield.Name, pfield.Hash)
						return errors.New(msg)
					}
				case domain.String:
					if v, ok := value.(string); ok {
						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be string", pfield.Name, pfield.Hash)
						return errors.New(msg)
					}
				case domain.Text:
					if v, ok := value.(string); ok {
						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be text", pfield.Name, pfield.Hash)
						return errors.New(msg)
					}
				case domain.Bool:
					if v, ok := value.(bool); ok {
						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be bool", pfield.Name, pfield.Hash)
						return errors.New(msg)
					}
				case domain.Switch:
					if v, ok := value.(float64); ok {
						if v == 0 || v == 1 || v == 2 {
							filteredFields[pfield.Hash] = int(v)
						} else {
							msg := fmt.Sprintf("field %s (%s) must be switch (0|1|2)", pfield.Name, pfield.Hash)
							return errors.New(msg)
						}
						continue
					} else {
						msg := fmt.Sprintf("field %s (%s) should be switch (0|1|2)", pfield.Name, pfield.Hash)
						return errors.New(msg)
					}
				case domain.Array:
					rt := reflect.TypeOf(value)
					if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
						arrWithStrings := []string{}
						for _, i := range value.([]interface{}) {
							arrWithStrings = append(arrWithStrings, fmt.Sprintf("%v", i))
						}

						filteredFields[pfield.Hash] = arrWithStrings
					} else {
						msg := fmt.Sprintf("field %s (%s) should be array", pfield.Name, pfield.Hash)
						return errors.New(msg)
					}
				case domain.Data:
					if v, ok := value.(string); ok {
						uid := uuid.MustParse(v)

						entities[uid.String()] = uid

						filteredFields[pfield.Hash] = uid
					} else {
						msg := fmt.Sprintf("field %s (%s) should be string", pfield.Name, pfield.Hash)
						return errors.New(msg)
					}
				}

			}
		}

		if len(addedFieldsHash) != len(catalogData.RawFields) {
			canBeAdded := addedFieldsHash
			sendedToAdd := helpers.GetMapKeys(catalogData.RawFields)

			unwantedFields := helpers.ArrayNonIntersection(canBeAdded, sendedToAdd)

			if len(unwantedFields) == 0 {
				return errors.New("в каталоге нет кастомных полей")
			}

			msg := fmt.Sprintf("невозможно добавить: (%s)", strings.Join(unwantedFields, ","))

			return errors.New(msg)
		}

		catalogData.Fields = filteredFields
		catalogData.Entities = entities
	}

	return nil
}
