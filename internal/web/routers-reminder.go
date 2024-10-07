package web

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/oreminder"
	echo "github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func initOpenAPIReminderRouters(a *Web, e *echo.Echo) {
	logrus.WithField("route", "oReminder").Debug("routes initialization")

	midlewares := []oapi.StrictMiddlewareFunc{
		ValidateStructMiddeware,
		AuthMiddeware(a.app, []string{}),
	}

	handlers := oapi.NewStrictHandler(a, midlewares)
	oapi.RegisterHandlers(e, handlers)
}

func (a *Web) DeleteReminderUUID(ctx context.Context, request oapi.DeleteReminderUUIDRequestObject) (oapi.DeleteReminderUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.RemindersService.DeleteByUUID(request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteReminderUUID200Response{}, nil
}

func (a *Web) PostReminder(ctx context.Context, request oapi.PostReminderRequestObject) (oapi.PostReminderResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	// @todo: add type
	dm := domain.Reminder{
		UUID:          uuid.New(),
		TaskUUID:      request.Body.TaskUuid,
		CreatedBy:     claims.Email,
		CreatedByUUID: claims.UUID,
		Description:   request.Body.Description,
		Type:          request.Body.Type,
		DateTo:        request.Body.DateTo,
		DateFrom:      request.Body.DateFrom,
		UserUUID:      request.Body.UserUuid,
	}

	err := a.app.RemindersService.Create(dm)
	if err != nil {
		return nil, err
	}

	return oapi.PostReminder200JSONResponse{
		Uuid: dm.UUID,
	}, nil
}

func (a *Web) PutReminderUUID(ctx context.Context, request oapi.PutReminderUUIDRequestObject) (oapi.PutReminderUUIDResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	// @todo: add type
	dm, err := a.app.RemindersService.Get(request.UUID)
	if err != nil {
		return nil, err
	}

	dm.Description = request.Body.Description
	dm.Comment = request.Body.Comment
	dm.DateFrom = request.Body.DateFrom
	dm.DateTo = request.Body.DateTo
	dm.Type = request.Body.Type
	dm.UserUUID = request.Body.UserUuid

	err = a.app.RemindersService.Put(claims.Email, dm)
	if err != nil {
		return nil, err
	}

	return oapi.PutReminderUUID200Response{}, nil
}

func (a *Web) PatchReminderUUIDStatus(ctx context.Context, request oapi.PatchReminderUUIDStatusRequestObject) (oapi.PatchReminderUUIDStatusResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dm, err := a.app.RemindersService.Get(request.UUID)
	if err != nil {
		return nil, err
	}

	err = a.app.RemindersService.PatchStatus(claims.Email, dm, request.Body.Status)
	if err != nil {
		return nil, err
	}

	return oapi.PatchReminderUUIDStatus200Response{}, nil
}

func (a *Web) GetReminder(ctx context.Context, _ oapi.GetReminderRequestObject) (oapi.GetReminderResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	dms, err := a.app.RemindersService.GetByUser(claims.UUID)
	if err != nil {
		return nil, err
	}

	dtos := lo.Map(dms, func(dm domain.Reminder, _ int) dto.ReminderDTO {
		var user *dto.UserDTO
		if dm.UserUUID != nil {
			fuser, f := a.app.DictionaryService.FindUserByUUID(*dm.UserUUID)
			if !f {
				logrus.WithField("user_uuid", dm.UserUUID).Warn("user not found in reminder")
			} else {
				user = fuser
			}
		}

		createdBy, f := a.app.DictionaryService.FindUserByUUID(dm.CreatedByUUID)
		if !f {
			logrus.WithField("user_uuid", dm.UserUUID).Warn("created by user not found in reminder")
		}

		return dto.ReminderDTO{
			UUID:        dm.UUID,
			TaskUUID:    dm.TaskUUID,
			Description: dm.Description,
			Comment:     dm.Comment,
			DateTo:      dm.DateTo,
			DateFrom:    dm.DateFrom,
			Type:        dm.Type,
			CreatedAt:   dm.CreatedAt,
			UpdatedAt:   dm.UpdatedAt,
			User:        user,
			CreatedBy:   createdBy,
			Status:      dm.Status,
		}
	})

	return oapi.GetReminder200JSONResponse{
		Count: len(dtos),
		Items: dtos,
	}, nil
}
