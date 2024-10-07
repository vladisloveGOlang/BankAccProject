package web

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/jwt"
	oapi "github.com/krisch/crm-backend/internal/web/oprofile"
	"github.com/sirupsen/logrus"
)

func (a *Web) GetProfileNotifications(ctx context.Context, _ oapi.GetProfileNotificationsRequestObject) (oapi.GetProfileNotificationsResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	// Tasks
	dtos, err := a.app.NotificationsService.GetNotification(claims.Email)
	if err != nil {
		return nil, err
	}

	// Sort
	sort.Slice(dtos, func(i, j int) bool {
		return dtos[i].Score > dtos[j].Score
	})

	//
	reminderUUIDSs := []uuid.UUID{}
	taskUUIDSs := []uuid.UUID{}

	for _, item := range dtos {
		uid, err := uuid.Parse(item.UUID)
		if err != nil {
			logrus.Warn(err)
			continue
		}

		if item.Type == "reminder" {
			reminderUUIDSs = append(reminderUUIDSs, uid)
		}

		if item.Type == "task" {
			taskUUIDSs = append(taskUUIDSs, uid)
		}
	}

	// Get Tasks
	taskWithName, err := a.app.TaskService.GetTasksNames(ctx, taskUUIDSs)
	if err != nil {
		return nil, err
	}

	// Get Reminders
	remindersWithName, err := a.app.RemindersService.GetRemindersNames(ctx, reminderUUIDSs)
	if err != nil {
		return nil, err
	}

	// Mapping
	taskWithNameMap := make(map[string]string)
	for _, item := range taskWithName {
		taskWithNameMap[item.UUID.String()] = item.Name
	}

	reminderWithNameMap := make(map[string]string)
	for _, item := range remindersWithName {
		reminderWithNameMap[item.UUID.String()] = item.Description
	}

	// DTO
	items := []interface{}{}
	for _, item := range dtos {

		taskUUID, err := uuid.Parse(item.UUID)
		if err != nil {
			logrus.Warn(err)
			continue
		}

		if item.Type == "task" {
			typeName := taskWithNameMap[item.UUID]

			state, star, err := a.app.NotificationsService.GetTaskState(claims.Email, taskUUID)
			if err != nil {
				logrus.Warnf("GetTaskState: %s", err)
			}

			count := make(map[string]interface{})
			count["comment"] = len(state.NewComments)
			count["upload"] = len(state.NewUploads)
			count["mensions"] = state.NewMentions
			count["comment_like"] = state.NewLikes
			count["reminders"] = len(state.NewReminders)

			group := 0
			if star {
				group = 1
			}

			items = append(items, dto.NotificationTaskDTO{
				UUID:  item.UUID,
				Type:  item.Type,
				Name:  typeName,
				Score: float64(state.UpdatedAt.UnixMicro()),
				Count: count,

				Star: item.Star,

				Opened: false,
				Group:  group,

				Comments:  state.NewComments,
				Reminders: state.NewReminders,
				Uploads:   state.NewUploads,
			})
		}
	}

	return oapi.GetProfileNotifications200JSONResponse{
		Count: len(dtos),
		Items: items,
	}, nil
}

func (a *Web) DeleteProfileNotifications(ctx context.Context, _ oapi.DeleteProfileNotificationsRequestObject) (oapi.DeleteProfileNotificationsResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.NotificationsService.RemoveNotifications(ctx, claims.Email)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteProfileNotifications200Response{}, nil
}

func (a *Web) PostProfileNotificationsTaskUUIDStar(ctx context.Context, request oapi.PostProfileNotificationsTaskUUIDStarRequestObject) (oapi.PostProfileNotificationsTaskUUIDStarResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.NotificationsService.ToggleStarNotification(ctx, "task", claims.Email, request.UUID, true)
	if err != nil {
		return nil, err
	}

	return oapi.PostProfileNotificationsTaskUUIDStar200Response{}, nil
}

func (a *Web) DeleteProfileNotificationsTaskUUIDStar(ctx context.Context, request oapi.DeleteProfileNotificationsTaskUUIDStarRequestObject) (oapi.DeleteProfileNotificationsTaskUUIDStarResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.NotificationsService.ToggleStarNotification(ctx, "task", claims.Email, request.UUID, false)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteProfileNotificationsTaskUUIDStar200Response{}, nil
}

func (a *Web) PostProfileNotificationsTaskUUIDHide(ctx context.Context, request oapi.PostProfileNotificationsTaskUUIDHideRequestObject) (oapi.PostProfileNotificationsTaskUUIDHideResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.NotificationsService.HideNotification(ctx, "task", claims.Email, request.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.PostProfileNotificationsTaskUUIDHide200Response{}, nil
}
