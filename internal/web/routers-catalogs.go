package web

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/catalogs"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ocatalog"
	echo "github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func initOpenAPIcatalogRouters(a *Web, e *echo.Echo) {
	logrus.WithField("route", "ocatalog").Debug("routes initialization")

	midlewares := []oapi.StrictMiddlewareFunc{
		ValidateStructMiddeware,
		AuthMiddeware(a.app, []string{}),
	}

	handlers := oapi.NewStrictHandler(a, midlewares)
	oapi.RegisterHandlers(e, handlers)
}

func (a *Web) DeleteCatalogUUID(ctx context.Context, request oapi.DeleteCatalogUUIDRequestObject) (oapi.DeleteCatalogUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CatalogService.DeleteCatalog(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteCatalogUUID200Response{}, nil
}

func (a *Web) PostCatalog(ctx context.Context, request oapi.PostCatalogRequestObject) (oapi.PostCatalogResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	federationDTO, err := a.app.FederationService.GetCompanyFederation(ctx, request.Body.CompanyUuid)
	if err != nil {
		return nil, err
	}

	if federationDTO.UUID == uuid.Nil {
		return nil, dto.NotFoundErr("Компания или федерация не найдены")
	}

	if helpers.InArray(request.Body.Name, catalogs.GetReservedNames()) {
		return nil, errors.New("название каталога зарезервировано (он уже создан)")
	}

	catalog := &domain.Catalog{
		UUID:           uuid.New(),
		FederationUUID: federationDTO.UUID,
		CompanyUUID:    request.Body.CompanyUuid,
		Name:           request.Body.Name,
		CreatedBy:      claims.Email,
		CreatedByUUID:  claims.UUID,
	}

	err = a.app.CatalogService.CreateCatalog(catalog)
	if err != nil {
		return nil, err
	}

	return oapi.PostCatalog200JSONResponse{
		Uuid: catalog.UUID,
	}, nil
}

func (a *Web) GetCatalogUUID(ctx context.Context, request oapi.GetCatalogUUIDRequestObject) (oapi.GetCatalogUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dmn, err := a.app.CatalogService.GetCatalog(request.UUID)
	if err != nil {
		return nil, err
	}

	f, _ := a.app.DictionaryService.FindFederation(dmn.FederationUUID)
	c, _ := a.app.DictionaryService.FindCompany(dmn.CompanyUUID)

	allowSort := a.app.CatalogService.GetSortFields(dmn.UUID)

	dt := dto.CatalogDTO{
		UUID:      dmn.UUID,
		Name:      dmn.Name,
		CreatedAt: dmn.CreatedAt,
		UpdatedAt: dmn.UpdatedAt,

		Federation: dto.FederationDTOs{
			UUID: f.UUID,
			Name: f.Name,
		},
		Company: dto.CompanyDTOs{
			UUID: c.UUID,
			Name: c.Name,
		},

		FieldsTotal: len(dmn.Fields),
		Fields: helpers.Map(dmn.Fields, func(item domain.CatalogFiled, index int) dto.CatalogFieldDTO {
			return dto.CatalogFieldDTO{
				UUID:     item.UUID,
				Name:     item.Name,
				Hash:     item.Hash,
				DataType: int(item.DataType),
				DataDesc: item.FieldTypeDesc(),
			}
		}),

		AllowSort: allowSort,
	}

	return oapi.GetCatalogUUID200JSONResponse(dt), nil
}

func (a *Web) GetCatalog(ctx context.Context, request oapi.GetCatalogRequestObject) (oapi.GetCatalogResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dmns, err := a.app.CatalogService.GetCatalogsByCompany(request.Body.CompanyUuid)
	if err != nil {
		return nil, err
	}

	dtos := helpers.Map(dmns, func(dmn domain.Catalog, index int) dto.CatalogDTO {
		f, _ := a.app.DictionaryService.FindFederation(dmn.FederationUUID)
		c, _ := a.app.DictionaryService.FindCompany(dmn.CompanyUUID)

		return dto.CatalogDTO{
			UUID: dmn.UUID,
			Name: dmn.Name,

			Federation: dto.FederationDTOs{
				UUID: f.UUID,
				Name: f.Name,
			},
			Company: dto.CompanyDTOs{
				UUID: c.UUID,
				Name: c.Name,
			},

			FieldsTotal: len(dmn.Fields),
			Fields: helpers.Map(dmn.Fields, func(item domain.CatalogFiled, index int) dto.CatalogFieldDTO {
				return dto.CatalogFieldDTO{
					UUID:     item.UUID,
					Name:     item.Name,
					Hash:     item.Hash,
					DataType: int(item.DataType),
					DataDesc: item.FieldTypeDesc(),
				}
			}),

			CreatedAt: dmn.CreatedAt,
			UpdatedAt: dmn.UpdatedAt,
		}
	})

	return oapi.GetCatalog200JSONResponse{
		Count: len(dtos),
		Items: dtos,
	}, nil
}

func (a *Web) PostCatalogUUIDFields(ctx context.Context, request oapi.PostCatalogUUIDFieldsRequestObject) (oapi.PostCatalogUUIDFieldsResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	pf := domain.NewCatalogFiled(request.Body.Name, "", request.Body.DataType, request.Body.DataUuid, request.UUID, claims.Email)

	dt, err := a.app.CatalogService.CreateCatalogField(pf)
	if err != nil {
		return nil, err
	}

	return oapi.PostCatalogUUIDFields200JSONResponse{
		Uuid:            dt.UUID,
		CatalogUuid:     dt.CatalogUUID,
		Hash:            dt.Hash,
		Type:            dt.DataType,
		TypeUuid:        dt.DataCatalogUUID,
		TypeDescription: pf.FieldTypeDesc(),
	}, nil
}

func (a *Web) PostCatalogUUIDFieldsNamed(ctx context.Context, request oapi.PostCatalogUUIDFieldsNamedRequestObject) (oapi.PostCatalogUUIDFieldsNamedResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	pf := domain.NewCatalogFiled(request.Body.Name, request.Body.Hash, request.Body.DataType, nil, request.UUID, claims.Email)

	dt, err := a.app.CatalogService.CreateCatalogField(pf)
	if err != nil {
		return nil, err
	}

	return oapi.PostCatalogUUIDFieldsNamed200JSONResponse{
		Uuid:            dt.UUID.String(),
		Hash:            dt.Hash,
		Type:            dt.DataType,
		TypeDescription: pf.FieldTypeDesc(),
	}, nil
}

func (a *Web) PutCatalogUUIDFieldsEntityUUID(ctx context.Context, request oapi.PutCatalogUUIDFieldsEntityUUIDRequestObject) (oapi.PutCatalogUUIDFieldsEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	pf := &domain.CatalogFiled{
		CatalogUUID: request.UUID,
		UUID:        request.EntityUUID,
		Name:        request.Body.Name,
	}

	err := a.app.CatalogService.PutCatalogField(pf)
	if err != nil {
		return nil, err
	}

	return oapi.PutCatalogUUIDFieldsEntityUUID200Response{}, nil
}

func (a *Web) GetCatalogUUIDFields(ctx context.Context, request oapi.GetCatalogUUIDFieldsRequestObject) (oapi.GetCatalogUUIDFieldsResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	items, err := a.app.CatalogService.GetCatalogFields(request.UUID)
	if err != nil {
		return nil, err
	}

	fields := helpers.Map(items, func(item domain.CatalogFiled, index int) dto.CatalogFieldDTO {
		return dto.CatalogFieldDTO{
			UUID:            item.UUID,
			Name:            item.Name,
			Hash:            item.Hash,
			DataType:        int(item.DataType),
			DataCatalogUUID: item.DataCatalogUUID,
			DataDesc:        item.FieldTypeDesc(),
		}
	})

	return oapi.GetCatalogUUIDFields200JSONResponse{
		Count: len(items),
		Items: fields,
	}, nil
}

func (a *Web) PatchCatalogUUIDName(ctx context.Context, request oapi.PatchCatalogUUIDNameRequestObject) (oapi.PatchCatalogUUIDNameResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CatalogService.ChangeCatalogName(request.UUID, request.Body.Name)
	if err != nil {
		return nil, err
	}

	return oapi.PatchCatalogUUIDName200Response{}, nil
}

func (a *Web) DeleteCatalogUUIDFieldsEntityUUID(ctx context.Context, request oapi.DeleteCatalogUUIDFieldsEntityUUIDRequestObject) (oapi.DeleteCatalogUUIDFieldsEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CatalogService.DeleteCatalogField(request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteCatalogUUIDFieldsEntityUUID200Response{}, nil
}

func (a *Web) PostCatalogUUIDData(ctx context.Context, request oapi.PostCatalogUUIDDataRequestObject) (oapi.PostCatalogUUIDDataResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	catalog, err := a.app.CatalogService.GetCatalog(request.UUID)
	if err != nil {
		return nil, err
	}

	catalogData := domain.CatalogData{
		UUID:           uuid.New(),
		FederationUUID: catalog.FederationUUID,
		CompanyUUID:    catalog.CompanyUUID,
		CatalogUUID:    catalog.UUID,
		RawFields:      request.Body.Fields,
		Entities:       make(map[string]interface{}),

		CreatedBy:     claims.Email,
		CreatedByUUID: claims.UUID,
	}

	_, err = a.app.CatalogService.AddData(catalogData)
	if err != nil {
		return nil, err
	}

	return oapi.PostCatalogUUIDData200JSONResponse{
		Uuid: catalogData.UUID,
	}, nil
}

func (a *Web) GetCatalogUUIDData(ctx context.Context, request oapi.GetCatalogUUIDDataRequestObject) (oapi.GetCatalogUUIDDataResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	filterDto, err := dto.NewFilterDTO(request.Params.Fields)
	if err != nil {
		return nil, err
	}

	search := dto.CatalogSearchDTO{
		CatalogUUID: request.UUID,
		Offset:      request.Params.Offset,
		Limit:       request.Params.Limit,
		Fields:      filterDto,

		Order: request.Params.Order,
		By:    request.Params.By,
	}

	err = search.Validate()
	if err != nil {
		return nil, err
	}

	dmns, total, err := a.app.CatalogService.GetData(search)
	if err != nil {
		return nil, err
	}

	dtos := lo.Map(dmns, func(dmn domain.CatalogData, index int) map[string]interface{} {
		// @todo
		dto := dto.NewCatalogDataDTO(dmn, a.app.DictionaryService)

		mp := make(map[string]interface{})

		mp["uuid"] = dto.UUID
		mp["created_at"] = dto.CreatedAt
		mp["updated_at"] = dto.UpdatedAt

		// @todo
		entitiesByUUID := make(map[string]map[string]interface{})
		for _, e := range dmn.EntitiesRich {
			if entity, ok := e.(map[string]interface{}); ok {
				if uuid, ok := entity["uuid"].(string); ok {
					if v, ok := e.(map[string]interface{}); ok {
						entitiesByUUID[uuid] = v
					}
				}
			}
		}

		// @todo
		for _, field := range dto.Fields {
			mp[field.Hash] = field.Value

			if v, ok := field.Value.(string); ok {
				if e, ok := entitiesByUUID[v]; ok {
					if entityFields, ok := e["fields"].(map[string]interface{}); ok {
						for k, v := range entityFields {
							e[k] = v
						}

						delete(e, "fields")
						mp[field.Hash] = e
					}
				}
			}
		}

		return mp
	})

	return oapi.GetCatalogUUIDData200JSONResponse{
		Body: struct {
			Count int                      `json:"count"`
			Items []map[string]interface{} `json:"items"`
			Total int64                    `json:"total"`
		}{
			Count: len(dtos),
			Items: dtos,
			Total: total,
		},
		Headers: oapi.GetCatalogUUIDData200ResponseHeaders{
			CacheControl: "no-cache",
		},
	}, nil
}
