package web

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/ofederation"
)

// DeleteProfileSurveyUUID implements oapi.StrictServerInterface.
func (a *Web) DeleteProfileSurveyUUID(_ context.Context, request oapi.DeleteProfileSurveyUUIDRequestObject) (oapi.DeleteProfileSurveyUUIDResponseObject, error) {
	err := a.app.ProfileService.DeleteSurvey(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteProfileSurveyUUID200Response{}, nil
}

// GetProfileSurveyUUID implements oapi.StrictServerInterface.
func (a *Web) GetProfileSurveyUUID(_ context.Context, request oapi.GetProfileSurveyUUIDRequestObject) (oapi.GetProfileSurveyUUIDResponseObject, error) {
	dm, err := a.app.ProfileService.GetSurvey(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.GetProfileSurveyUUID200JSONResponse(dto.SurveyDTO{
		UUID:      dm.UUID,
		User:      dto.NewUserDto(dm.User, a.app.S3Service),
		Name:      dm.Name,
		Body:      dm.Body,
		CreatedAt: dm.CreatedAt,
		UpdatedAt: dm.UpdatedAt,
		DeletedAt: dm.DeletedAt,
	}), nil
}

// PostProfileSurvey implements oapi.StrictServerInterface.
func (a *Web) PostProfileSurvey(ctx context.Context, request oapi.PostProfileSurveyRequestObject) (oapi.PostProfileSurveyResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dm := domain.Survey{
		UUID: uuid.New(),
		User: domain.User{
			UUID:  claims.UUID,
			Email: claims.Email,
		},
		Name: request.Body.Name,
		Body: request.Body.Body,
	}

	err := a.app.ProfileService.CreateSurvey(dm)
	if err != nil {
		return nil, err
	}

	return oapi.PostProfileSurvey200JSONResponse{
		Uuid: dm.UUID,
	}, nil
}
