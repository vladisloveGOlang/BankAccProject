package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/jwt"
	"github.com/krisch/crm-backend/internal/profile"
	oapi "github.com/krisch/crm-backend/internal/web/otask"
	echo "github.com/labstack/echo/v4"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

func initOpenAPITaskRouters(a *Web, e *echo.Echo) {
	logrus.WithField("route", "oTask").Debug("routes initialization")

	midlewares := []oapi.StrictMiddlewareFunc{
		ValidateStructMiddeware,
		AuthMiddeware(a.app, []string{}),
	}

	handlers := oapi.NewStrictHandler(a, midlewares)
	oapi.RegisterHandlers(e, handlers)
}

func (a *Web) PatchTaskUUIDCommentEntityUUID(ctx context.Context, request oapi.PatchTaskUUIDCommentEntityUUIDRequestObject) (oapi.PatchTaskUUIDCommentEntityUUIDResponseObject, error) {
	// @todo: need refactoring
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	form, err := request.Body.ReadForm(1000000)
	if err != nil {
		return nil, err
	}

	comment := ""
	replyUUID := uuid.Nil
	emails := []string{}
	var replyMessage *string

	if form.Value["comment"] != nil && len(form.Value["comment"]) > 0 {
		comment = form.Value["comment"][0]
	}
	if form.Value["reply_uuid"] != nil && len(form.Value["reply_uuid"]) > 0 {
		replyUUID = uuid.MustParse(form.Value["reply_uuid"][0])

		m, err := a.app.CommentService.CheckCommentText(replyUUID)
		if err != nil {
			return nil, err
		}
		replyMessage = &m
	}
	if form.Value["people"] != nil {
		emails = form.Value["people"]
	}

	dm := domain.NewComment(claims.Email, request.UUID, replyUUID, emails, comment)
	dm.UUID = request.EntityUUID

	err = a.app.TaskService.UpdateComment(ctx, request.UUID, *dm)
	if err != nil {
		return nil, err
	}

	var uploadsDTO *[]dto.UploadDTO
	if form.File["file"] != nil && len(form.File["file"]) > 0 {
		uploadsDTO = &[]dto.UploadDTO{}

		for _, f := range form.File["file"] {
			file, err := f.Open()
			if err != nil {
				return nil, err
			}
			storeFilePath := "/tmp/" + helpers.FakeString(10) + "-" + f.Filename
			dst, err := os.Create(storeFilePath)
			if err != nil {
				return nil, err
			}

			if _, err := io.Copy(dst, file); err != nil {
				return nil, err
			}

			logrus.Debug("file saved to disk:", storeFilePath)

			task, err := a.app.TaskService.GetTask(ctx, request.UUID, []string{})
			if err != nil {
				return nil, err
			}

			fileDTO, err := a.app.S3PrivateService.UploadTaskCommentFile(task.FederationUUID, task.UUID, dm.UUID, f.Filename, storeFilePath, claims.UUID)
			if err != nil {
				return nil, err
			}

			os.Remove(storeFilePath)

			url, err := a.app.S3PrivateService.PresignedURL(fileDTO.Name, fileDTO.ObjectName)
			if err != nil {
				return nil, err
			}

			*uploadsDTO = append(*uploadsDTO, dto.NewUploadDTO(fileDTO.UUID, fileDTO.Name, fileDTO.Ext, fileDTO.Size, url))
		}
	}

	var peoplesDto *[]dto.UserDTO
	if len(emails) > 0 {
		p, _ := a.app.DictionaryService.FindUsers(emails)
		peoplesDto = &p
	}

	return oapi.PatchTaskUUIDCommentEntityUUID200JSONResponse{
		Uuid:         dm.UUID,
		Comment:      dm.Comment,
		ReplyMessage: replyMessage,
		Uploads:      uploadsDTO,
		People:       peoplesDto,
	}, nil
}

func (a *Web) PostTask(ctx context.Context, request oapi.PostTaskRequestObject) (oapi.PostTaskResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	project, find := a.app.DictionaryService.FindProject(request.Body.ProjectUuid)
	if !find {
		return nil, domain.ErrProjectNotFound
	}

	err := a.app.TaskService.CheckPath(request.Body.Path)
	if err != nil {
		return nil, err
	}

	te := make(map[uuid.UUID][]string)
	for _, v := range request.Body.TaskEntities {
		te[v.UUID] = v.Fields
	}

	task, err := domain.NewTask(
		request.Body.Name,
		project.FederationUUID,
		project.CompanyUUID,
		request.Body.ProjectUuid,
		claims.Email,
		request.Body.Fields,
		request.Body.Tags,

		request.Body.Description,
		request.Body.Path,
		request.Body.CoworkersBy,
		request.Body.ImplementBy,
		request.Body.ResponsibleBy,

		request.Body.Priority,

		request.Body.FinishTo,
		request.Body.Icon,
		request.Body.ManagedBy,

		te,
	)
	if err != nil {
		return nil, err
	}

	id, err := a.app.TaskService.CreateTask(task)
	if err != nil {
		return nil, err
	}

	return oapi.PostTask200JSONResponse{
		Uuid: task.UUID,
		Id:   id,
	}, nil
}

// Web struct should implement the missing method from otask.StrictServerInterface.
func (a *Web) DeleteTaskUUID(ctx context.Context, request oapi.DeleteTaskUUIDRequestObject) (oapi.DeleteTaskUUIDResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.TaskService.DeleteTask(domain.NewCreatorFromUser(&claims), request.UUID)

	return oapi.DeleteTaskUUID200Response{}, err
}

func (a *Web) PutTaskUUID(ctx context.Context, request oapi.PutTaskUUIDRequestObject) (oapi.PutTaskUUIDResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	task, err := a.app.TaskService.GetTask(ctx, request.UUID, []string{})
	if err != nil {
		return nil, err
	}

	shouldUpdate := []string{}

	// @todo: active record?
	if request.Body.Fields != nil {
		task.RawFields = *request.Body.Fields
		shouldUpdate = append(shouldUpdate, "fields")
	}

	if request.Body.Description != nil {
		task.Description = *request.Body.Description
		shouldUpdate = append(shouldUpdate, "description")
	}

	if request.Body.Tags != nil {
		task.Tags = *request.Body.Tags
		task.Tags = lo.Uniq(task.Tags)
		shouldUpdate = append(shouldUpdate, "tags")
	}

	if request.Body.Priority != nil {
		task.Priority = *request.Body.Priority
		shouldUpdate = append(shouldUpdate, "priority")
	}

	if request.Body.FinishTo != nil {
		task.FinishTo = request.Body.FinishTo
		shouldUpdate = append(shouldUpdate, "finish_to")
	}

	if request.Body.Icon != nil {
		task.Icon = *request.Body.Icon
		shouldUpdate = append(shouldUpdate, "icon")
	}

	err = a.app.TaskService.UpdateTask(domain.NewCreatorFromUser(&claims), task, shouldUpdate)
	if err != nil {
		return nil, err
	}

	return oapi.PutTaskUUID200Response{}, nil
}

func (a *Web) GetTaskUUID(ctx context.Context, request oapi.GetTaskUUIDRequestObject) (oapi.GetTaskUUIDResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}
	//nolint
	ctx = context.WithValue(ctx, "userUUID", claims.UUID)
	//nolint
	ctx = context.WithValue(ctx, "userEmail", claims.Email)

	// cache
	dtoFromCache, err := a.app.CacheService.GetTask(ctx, request.UUID)
	if err != nil {
		return nil, err
	}

	// likes
	isLiked, err := a.app.ProfileService.CheckEntityHasLike(string(profile.Task), request.UUID, claims.UUID)
	if err != nil {
		return nil, err
	}

	if err == nil && dtoFromCache.UUID != uuid.Nil {
		firstOpenDTO := a.patchFirstOpen(&dtoFromCache, claims.UUID)

		dtoFromCache.IsLiked = &isLiked
		dtoFromCache.FirstOpen = firstOpenDTO
		dtoFromCache.Views = len(firstOpenDTO)

		err := a.app.TaskService.TaskWasOpen(request.UUID, claims.Email)
		if err != nil {
			logrus.Error(err)
		}

		logrus.Info("[module:router] GetTask: from redis")
		return oapi.GetTaskUUID200JSONResponse{
			Body: dtoFromCache,
			Headers: oapi.GetTaskUUID200ResponseHeaders{
				CacheControl: "private",
			},
		}, nil
	}

	// db
	dm, err := a.app.TaskService.GetTask(ctx, request.UUID, []string{"activities"})
	if err != nil {
		return nil, err
	}

	// comments
	comments, err := a.app.CommentService.GetTaskComments(dm.UUID, true, true)
	if err != nil {
		return nil, err
	}

	// files
	files, err := a.app.S3PrivateService.GetTaskFiles(dm.UUID, true)
	if err != nil {
		return nil, err
	}

	// reminders
	reminders, err := a.app.RemindersService.GetByTask(dm.UUID)
	if err != nil {
		return nil, err
	}

	// task linked fields
	linkedFieldsData := make(map[uuid.UUID]interface{})

	for linkedUID, linkedFields := range dm.TaskEntities {
		linkedDm, err := a.app.TaskService.GetTask(ctx, linkedUID, []string{})
		if err == nil {
			linkedFieldsData[linkedDm.UUID] = lo.Map(linkedFields, func(fieldHash string, index int) interface{} {
				m := make(map[string]interface{})
				m[fieldHash] = linkedDm.Fields[fieldHash]
				return m
			})
		} else {
			logrus.Errorf("[uuid:%s] linked task not found", linkedUID)
		}
	}

	taskDto := dto.NewTaskDTO(dm, comments, files, reminders, linkedFieldsData, a.app.DictionaryService, a.app.ProfileService)

	go a.app.CacheService.CacheTask(ctx, &taskDto)
	taskDto.IsLiked = &isLiked

	// First Open
	firstOpenDTO := a.patchFirstOpen(&taskDto, claims.UUID)
	taskDto.FirstOpen = firstOpenDTO
	taskDto.Views = len(firstOpenDTO)

	return oapi.GetTaskUUID200JSONResponse{
		Body: taskDto,
		Headers: oapi.GetTaskUUID200ResponseHeaders{
			CacheControl: "no-cache",
		},
	}, nil
}

func (a *Web) patchFirstOpen(dtoFromCache *dto.TaskDTO, me uuid.UUID) []dto.OpenByDTO {
	// Search me in team
	shouldAddToOpenBy := true
	for _, o := range dtoFromCache.FirstOpen {
		if o.OpenBy.UUID == me {
			shouldAddToOpenBy = false
			break
		}
	}

	if shouldAddToOpenBy {
		o, f := a.app.DictionaryService.FindUserByUUID(me)
		if f {
			dtoFromCache.FirstOpen = append(dtoFromCache.FirstOpen, dto.OpenByDTO{
				OpenAt: dtoFromCache.CreatedAt,
				OpenBy: *o,
			})

			err := a.app.TaskService.PatchFirstOpenBy(context.TODO(), dtoFromCache.UUID, me)
			if err != nil {
				logrus.Error(err)
			}
		} else {
			logrus.WithField("uuid", me).Error("user not found by uuid")
		}
	}

	return dtoFromCache.FirstOpen
}

func (a *Web) GetTask(ctx context.Context, request oapi.GetTaskRequestObject) (oapi.GetTaskResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	filterDto, err := dto.NewFilterDTO(request.Params.Fields)
	if err != nil {
		return nil, err
	}

	filter := dto.TaskSearchDTO{
		MyEmail: &claims.Email,

		Name:           request.Params.Name,
		Offset:         request.Params.Offset,
		Limit:          request.Params.Limit,
		IsMy:           request.Params.IsMy,
		IsEpic:         request.Params.IsEpic,
		Status:         request.Params.Status,
		Participated:   request.Params.Participated,
		FederationUUID: request.Params.FederationUuid,
		ProjectUUID:    request.Params.ProjectUuid,
		Tags:           request.Params.Tags,
		Fields:         filterDto,
		Path:           request.Params.Path,

		Order: request.Params.Order,
		By:    request.Params.By,
	}

	err = filter.Validate()
	if err != nil {
		return nil, err
	}

	dtos, total, err := a.app.TaskService.GetTasksDto(ctx, filter)
	if err != nil {
		return nil, err
	}

	if request.Params.Format != nil && *request.Params.Format == "xlsx" {
		f, err := toExcel(dtos, "")
		if err != nil {
			return nil, err
		}

		buf, err := f.WriteToBuffer()
		if err != nil {
			return nil, err
		}

		project, _ := a.app.DictionaryService.FindProject(filter.ProjectUUID)
		name := project.Name

		contentDisposition := fmt.Sprintf("attachment; filename=\"%s.xlsx\";", helpers.Scientific(name))

		return oapi.GetTask200ApplicationxlsxResponse{
			Body: buf,

			Headers: oapi.GetTask200ResponseHeaders{
				CacheControl:       "no-cache",
				ContentType:        "application/octet-stream",
				ContentDisposition: contentDisposition,
			},
		}, nil

	}

	return oapi.GetTask200JSONResponse{
		Body: struct {
			Count int            `json:"count"`
			Items []dto.TaskDTOs `json:"items"`
			Total int64          `json:"total"`
		}{
			Count: len(dtos),
			Items: dtos,
			Total: total,
		},
		Headers: oapi.GetTask200ResponseHeaders{
			CacheControl: "no-cache",
			ContentType:  "application/json",
		},
	}, nil
}

func toExcel(dtos []dto.TaskDTOs, storeToDisk string) (f *excelize.File, err error) {
	f = excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheet := "Sheet1"

	max := 0
	maxMp := make(map[string]float64)

	err = f.SetSheetVisible(sheet, true)
	if err != nil {
		logrus.Error(err)
	}

	maxHeaderLevels := 0
	for idx, dto := range dtos {

		row, names := parseStruct(dto, make([]interface{}, 0), "", []string{})
		if len(row) > max {
			max = len(row)
		}

		currentHeaderLevel := 0
		if idx == 0 {
			headerCell, err := excelize.CoordinatesToCellName(1, 1)
			if err != nil {
				return f, err
			}

			for _, name := range names {
				lvl := strings.Count(name, ".")
				if lvl > maxHeaderLevels {
					maxHeaderLevels = lvl
				}
			}

			if maxHeaderLevels == 0 {
				err = f.SetSheetRow(sheet, headerCell, &names)
				if err != nil {
					logrus.Error(err)
				}
			} else {
				for i, name := range names {
					lvl := strings.Count(name, ".")

					namePart := strings.Split(name, ".")

					if lvl > currentHeaderLevel {
						for _j, _name := range names[i:] {
							_lvl := strings.Count(_name, ".")
							if _lvl == currentHeaderLevel {
								a, err := excelize.CoordinatesToCellName(i+1, currentHeaderLevel+1)
								if err != nil {
									return f, err
								}

								b, err := excelize.CoordinatesToCellName(i+_j, currentHeaderLevel+1)
								if err != nil {
									return f, err
								}

								err = f.MergeCell(sheet, a, b)
								if err != nil {
									return f, err
								}

								err = f.SetCellStr(sheet, a, namePart[lvl-1])
								if err != nil {
									return f, err
								}

								style, err := f.NewStyle(&excelize.Style{
									Alignment: &excelize.Alignment{
										Horizontal: "center",
									},
								})
								if err != nil {
									return f, err
								}

								err = f.SetCellStyle(sheet, a, a, style)
								if err != nil {
									logrus.Error(err)
								}

								break
							}
						}
					}

					currentHeaderLevel = lvl
					hcell, err := excelize.CoordinatesToCellName(i+1, lvl+1)
					if err != nil {
						logrus.Error(err)
					}
					err = f.SetCellStr(sheet, hcell, namePart[lvl])
					if err != nil {
						logrus.Error(err)
					}
				}
			}
		}

		cell, err := excelize.CoordinatesToCellName(1, idx+maxHeaderLevels+2)
		if err != nil {
			return f, err
		}

		err = f.SetSheetRow(sheet, cell, &row)
		if err != nil {
			return f, err
		}

		for i, r := range row {
			column := string(rune(65 + i))
			strLen := float64(len(fmt.Sprintf("%v", r)))
			if maxMp[column] < strLen {
				maxMp[column] = math.Min(strLen, 100)
			}
		}
	}

	for column, strLen := range maxMp {
		err := f.SetColWidth(sheet, column, column, strLen+5)
		if err != nil {
			return f, err
		}
	}

	err = f.AutoFilter(sheet, fmt.Sprintf("A1:%s1", string(rune(64+max))), []excelize.AutoFilterOptions{})
	if err != nil {
		return f, err
	}

	if len(storeToDisk) > 5 {
		if err := f.SaveAs("Book1.xlsx"); err != nil {
			return f, err
		}
	}

	return f, err
}

func parseStruct(dt interface{}, rows []interface{}, level string, names []string) ([]interface{}, []string) {
	elemType := reflect.TypeOf(dt)
	elemValue := reflect.ValueOf(dt)

	for j := 0; j < elemType.NumField(); j++ {
		field := elemType.Field(j)
		column := field.Tag.Get("xlsx")

		if column == "" {
			continue
		}

		value := elemValue.Field(j).Interface()
		kind := reflect.ValueOf(value).Kind()
		if kind.String() == "struct" {
			columnLocale := field.Tag.Get("ru")
			columnName := field.Name
			if columnLocale != "" {
				columnName = columnLocale
			}

			rows, names = parseStruct(value, rows, level+"."+columnName, names)
		} else {
			columnLocale := field.Tag.Get("ru")
			columnName := field.Name
			if columnLocale != "" {
				columnName = columnLocale
			}

			if fmt.Sprintf("%v", value) == "<nil>" {
				value = ""
			}

			if kind == reflect.Slice {
				value = strings.Join(value.([]string), ", ")
			}

			rows = append(rows, value)

			names = append(names, strings.TrimLeft(level+"."+columnName, "."))
		}
	}

	return rows, names
}

func (a *Web) PatchTaskUUIDParent(ctx context.Context, request oapi.PatchTaskUUIDParentRequestObject) (oapi.PatchTaskUUIDParentResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.TaskService.PatchTaskParent(ctx, request.UUID, request.Body.Uuid)
	if err != nil {
		return nil, err
	}

	return oapi.PatchTaskUUIDParent200Response{}, err
}

func (a *Web) PatchTaskUUIDProject(ctx context.Context, request oapi.PatchTaskUUIDProjectRequestObject) (oapi.PatchTaskUUIDProjectResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	task, err := a.app.TaskService.GetTask(ctx, request.UUID, []string{})
	if err != nil {
		return nil, err
	}

	project, err := a.app.AgregateService.GetProject(ctx, request.Body.Uuid)
	if err != nil {
		return nil, err
	}

	err = a.app.TaskService.PatchProject(domain.NewCreatorFromUser(&claims), task, project, request.Body.Status, request.Body.Comment)
	if err != nil {
		return nil, err
	}

	return oapi.PatchTaskUUIDProject200Response{}, err
}

func (a *Web) PatchTaskUUIDName(ctx context.Context, request oapi.PatchTaskUUIDNameRequestObject) (oapi.PatchTaskUUIDNameResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.TaskService.PatchName(domain.NewCreatorFromUser(&claims), request.UUID, request.Body.Name)
	if err != nil {
		return nil, err
	}

	return oapi.PatchTaskUUIDName200Response{}, err
}

func (a *Web) PatchTaskUUIDStatus(ctx context.Context, request oapi.PatchTaskUUIDStatusRequestObject) (oapi.PatchTaskUUIDStatusResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	task, err := a.app.TaskService.GetTask(ctx, request.UUID, []string{})
	if err != nil {
		return nil, err
	}

	project, err := a.app.AgregateService.GetProject(ctx, task.ProjectUUID)
	if err != nil {
		return nil, err
	}

	stopUUID, path, err := a.app.TaskService.PatchStatus(domain.NewCreatorFromUser(&claims), project, task, request.Body.Status, request.Body.Comment)
	if err != nil {
		return nil, err
	}

	return oapi.PatchTaskUUIDStatus200JSONResponse{
		StopUuid: stopUUID,
		Path:     path,
	}, err
}

func (a *Web) DeleteTaskUUIDStopEntityUUID(ctx context.Context, request oapi.DeleteTaskUUIDStopEntityUUIDRequestObject) (oapi.DeleteTaskUUIDStopEntityUUIDResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.TaskService.DeleteStop(ctx, request.UUID, request.EntityUUID, claims.Email)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteTaskUUIDStopEntityUUID200Response{}, err
}

func (a *Web) PostTaskUUIDComment(ctx context.Context, request oapi.PostTaskUUIDCommentRequestObject) (oapi.PostTaskUUIDCommentResponseObject, error) {
	// @todo: need refactoring
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	form, err := request.Body.ReadForm(1000000)
	if err != nil {
		return nil, err
	}

	comment := ""
	replyUUID := uuid.Nil
	emails := []string{}
	var replyMessage *string

	if form.Value["comment"] != nil && len(form.Value["comment"]) > 0 {
		comment = form.Value["comment"][0]
	}
	if form.Value["reply_uuid"] != nil && len(form.Value["reply_uuid"]) > 0 {
		replyUUID = uuid.MustParse(form.Value["reply_uuid"][0])

		m, err := a.app.CommentService.CheckCommentText(replyUUID)
		if err != nil {
			return nil, err
		}
		replyMessage = &m
	}
	if form.Value["people"] != nil {
		emails = form.Value["people"]
	}

	dm := domain.NewComment(claims.Email, request.UUID, replyUUID, emails, comment)

	err = a.app.TaskService.CreateComment(ctx, request.UUID, *dm)
	if err != nil {
		return nil, err
	}

	var uploadsDTO *[]dto.UploadDTO

	if form.File["file"] != nil && len(form.File["file"]) > 0 {
		uploadsDTO = &[]dto.UploadDTO{}

		for _, f := range form.File["file"] {
			file, err := f.Open()
			if err != nil {
				return nil, err
			}
			storeFilePath := "/tmp/" + helpers.FakeString(10) + "-" + f.Filename
			dst, err := os.Create(storeFilePath)
			if err != nil {
				return nil, err
			}

			if _, err := io.Copy(dst, file); err != nil {
				return nil, err
			}

			logrus.Debug("file saved to disk:", storeFilePath)

			task, err := a.app.TaskService.GetTask(ctx, request.UUID, []string{})
			if err != nil {
				return nil, err
			}

			fileDTO, err := a.app.S3PrivateService.UploadTaskCommentFile(task.FederationUUID, task.UUID, dm.UUID, f.Filename, storeFilePath, claims.UUID)
			if err != nil {
				return nil, err
			}

			os.Remove(storeFilePath)

			url, err := a.app.S3PrivateService.PresignedURL(fileDTO.Name, fileDTO.ObjectName)
			if err != nil {
				return nil, err
			}

			*uploadsDTO = append(*uploadsDTO, dto.NewUploadDTO(fileDTO.UUID, fileDTO.Name, fileDTO.Ext, fileDTO.Size, url))
		}
	}

	var peoplesDto *[]dto.UserDTO
	if len(emails) > 0 {
		p, _ := a.app.DictionaryService.FindUsers(emails)
		peoplesDto = &p
	}

	return oapi.PostTaskUUIDComment200JSONResponse{
		Uuid:         dm.UUID,
		Comment:      dm.Comment,
		ReplyMessage: replyMessage,
		Uploads:      uploadsDTO,
		People:       peoplesDto,
	}, nil
}

func (a *Web) PatchTaskUUIDCommentEntityUUIDLike(ctx context.Context, request oapi.PatchTaskUUIDCommentEntityUUIDLikeRequestObject) (oapi.PatchTaskUUIDCommentEntityUUIDLikeResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	likes, liked, err := a.app.CommentService.LikeComment(ctx, request.EntityUUID, claims.Email)
	if err != nil {
		return nil, err
	}

	// @todo: mv upload to service
	a.app.TaskService.ResetCache(request.UUID)

	// @todo:
	comment, err := a.app.CommentService.GetComment(ctx, request.EntityUUID)
	if err != nil {
		return nil, err
	}

	notify := []string{comment.CreatedBy}

	err = a.app.TaskService.TaskWasUpdatedOrCreated(comment.TaskUUID, notify)
	if err != nil {
		return nil, err
	}

	return oapi.PatchTaskUUIDCommentEntityUUIDLike200JSONResponse{
		Liked: liked,
		Likes: likes,
	}, nil
}

func (a *Web) PatchTaskUUIDCommentEntityUUIDPin(ctx context.Context, request oapi.PatchTaskUUIDCommentEntityUUIDPinRequestObject) (oapi.PatchTaskUUIDCommentEntityUUIDPinResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CommentService.PinComment(ctx, request.EntityUUID)
	if err != nil {
		return nil, err
	}

	// @todo: mv upload to service
	a.app.TaskService.ResetCache(request.UUID)

	return oapi.PatchTaskUUIDCommentEntityUUIDPin200Response{}, nil
}

func (a *Web) DeleteTaskUUIDCommentEntityUUID(ctx context.Context, request oapi.DeleteTaskUUIDCommentEntityUUIDRequestObject) (oapi.DeleteTaskUUIDCommentEntityUUIDResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CommentService.DeleteComment(ctx, request.UUID, request.EntityUUID, &claims.UUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteTaskUUIDCommentEntityUUID200Response{}, nil
}

// Web struct should implement the missing method from otask.StrictServerInterface.
func (a *Web) GetTaskUUIDComment(_ context.Context, request oapi.GetTaskUUIDCommentRequestObject) (oapi.GetTaskUUIDCommentResponseObject, error) {
	dms, err := a.app.CommentService.GetTaskComments(request.UUID, true, true)
	if err != nil {
		return nil, err
	}

	dtos := []dto.CommentDTO{}
	for _, dm := range dms {
		dtos = append(dtos, dto.NewCommentDTO(dm, a.app.DictionaryService, a.app.ProfileService))
	}

	return oapi.GetTaskUUIDComment200JSONResponse{
		Count: len(dtos),
		Items: dtos,
	}, nil
}

// Web struct should implement the missing method from otask.StrictServerInterface.
func (a *Web) PatchTaskUUIDTeam(ctx context.Context, request oapi.PatchTaskUUIDTeamRequestObject) (oapi.PatchTaskUUIDTeamResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.TaskService.PatchTeam(ctx, domain.NewCreatorFromUser(&claims), request.UUID, request.Body.ImplementBy, request.Body.ResponsibleBy, request.Body.CoworkersBy, request.Body.WatchedBy, request.Body.ManagedBy)
	if err != nil {
		return nil, err
	}

	task, err := a.app.TaskService.GetTask(ctx, request.UUID, []string{})
	if err != nil {
		return nil, err
	}

	taskDto := dto.NewTaskDTO(task, []domain.Comment{}, []domain.File{}, []domain.Reminder{}, map[uuid.UUID]interface{}{}, a.app.DictionaryService, a.app.ProfileService)

	return oapi.PatchTaskUUIDTeam200JSONResponse{
		CoworkersBy:   taskDto.CoWorkersBy,
		ImplementBy:   taskDto.ImplementBy,
		ManagedBy:     taskDto.ManagedBy,
		ResponsibleBy: taskDto.ResponsibleBy,
		WatchedBy:     taskDto.WatchBy,
	}, nil
}

func (a *Web) PatchTaskUUIDUpload(ctx context.Context, request oapi.PatchTaskUUIDUploadRequestObject) (oapi.PatchTaskUUIDUploadResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	file, err := request.Body.NextPart()
	if errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("file is required: %w", err)
	}
	defer file.Close()

	storeFilePath := "/tmp/" + helpers.FakeString(10) + "-" + file.FileName()
	dst, err := os.Create(storeFilePath)
	if err != nil {
		logrus.Errorf("error creating file: %s", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		logrus.Error(err)
	}

	logrus.Debug("file saved to disk:", storeFilePath)

	defer os.Remove(storeFilePath)

	task, err := a.app.TaskService.GetTask(ctx, request.UUID, []string{})
	if err != nil {
		return nil, err
	}

	fileDTO, err := a.app.S3PrivateService.UploadTaskFile(task.FederationUUID, task.UUID, file.FileName(), storeFilePath, claims.UUID)
	if err != nil {
		return nil, err
	}

	url, err := a.app.S3PrivateService.PresignedURL(fileDTO.Name, fileDTO.ObjectName)
	if err != nil {
		return nil, err
	}

	// @todo: mv upload to service
	a.app.TaskService.ResetCache(request.UUID)

	notify := lo.Filter(task.People, func(email string, _ int) bool {
		return email != claims.Email
	})

	err = a.app.TaskService.TaskWasUpdatedOrCreated(request.UUID, notify)
	if err != nil {
		return nil, err
	}

	return oapi.PatchTaskUUIDUpload200JSONResponse(dto.NewUploadDTO(fileDTO.UUID, fileDTO.Name, fileDTO.Ext, fileDTO.Size, url)), nil
}

func (a *Web) DeleteTaskUUIDUploadEntityUUID(ctx context.Context, request oapi.DeleteTaskUUIDUploadEntityUUIDRequestObject) (oapi.DeleteTaskUUIDUploadEntityUUIDResponseObject, error) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.TaskService.DeleteTaskFile(domain.NewCreatorFromUser(&claims), request.UUID, request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteTaskUUIDUploadEntityUUID200Response{}, nil
}

func (a *Web) PostTaskUUIDUploadEntityUUIDRename(ctx context.Context, request oapi.PostTaskUUIDUploadEntityUUIDRenameRequestObject) (oapi.PostTaskUUIDUploadEntityUUIDRenameResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.S3PrivateService.Rename(request.EntityUUID, request.Body.Name)
	if err != nil {
		return nil, err
	}

	// @todo: mv upload to service
	a.app.TaskService.ResetCache(request.UUID)

	return oapi.PostTaskUUIDUploadEntityUUIDRename200Response{}, nil
}

func (a *Web) GetTaskUUIDUploadEntityUUID(ctx context.Context, request oapi.GetTaskUUIDUploadEntityUUIDRequestObject) (oapi.GetTaskUUIDUploadEntityUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	url, err := a.app.S3PrivateService.PresignedURLFromFile(request.EntityUUID)
	if err != nil {
		return nil, err
	}

	return oapi.GetTaskUUIDUploadEntityUUID302Response{
		Headers: oapi.GetTaskUUIDUploadEntityUUID302ResponseHeaders{
			Location: url,
		},
	}, nil
}

func (a *Web) GetTaskUUIDUpload(ctx context.Context, request oapi.GetTaskUUIDUploadRequestObject) (oapi.GetTaskUUIDUploadResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	commentFiles, err := a.app.CommentService.GetTaskCommentsFiles(request.UUID)
	if err != nil {
		return nil, err
	}

	files, err := a.app.S3PrivateService.GetTaskFiles(request.UUID, true)
	if err != nil {
		return nil, err
	}

	uploads := files
	uploads = append(uploads, commentFiles...)

	items := lo.Map(uploads, func(item domain.File, index int) dto.UploadDTO {
		return dto.UploadDTO{
			UUID: item.UUID,
			Name: item.Name,
			EXT:  item.Ext,
			Size: item.Size,
			URL:  item.URL,
		}
	})

	return oapi.GetTaskUUIDUpload200JSONResponse{
		Count: len(items),
		Items: items,
	}, nil
}

func (a *Web) GetTaskUUIDActivity(ctx context.Context, request oapi.GetTaskUUIDActivityRequestObject) (oapi.GetTaskUUIDActivityResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	offset := helpers.If(request.Params.Offset == nil, 0, *request.Params.Offset)
	limit := helpers.If(request.Params.Limit == nil, 0, *request.Params.Limit)

	dms, total, err := a.app.TaskService.GetActivities(request.UUID, limit, offset)
	if err != nil {
		return nil, err
	}

	return oapi.GetTaskUUIDActivity200JSONResponse{
		Count: len(dms),
		Items: lo.Map(dms, func(item domain.Activity, i int) dto.ActivityDTO {
			createdBy, f := a.app.DictionaryService.FindUser(item.CreatedBy.Email)

			if !f {
				logrus.Errorf("activity created by not found: %s", item.CreatedBy.Email)
			}

			return *dto.NewActivityDTO(item, *createdBy)
		}),
		Total: total,
	}, nil
}

// DeleteTaskUUIDCommentEntityUUIDFileFileUUID implements otask.StrictServerInterface.
func (a *Web) DeleteTaskUUIDCommentEntityUUIDFileFileUUID(ctx context.Context, request oapi.DeleteTaskUUIDCommentEntityUUIDFileFileUUIDRequestObject) (oapi.DeleteTaskUUIDCommentEntityUUIDFileFileUUIDResponseObject, error) {
	_, ok := ctx.Value(claimsKey).(jwt.Claims)
	if !ok {
		return nil, ErrInvalidAuthHeader
	}

	err := a.app.CommentService.DeleteCommentFile(request.EntityUUID, request.FileUUID)
	if err != nil {
		return nil, err
	}

	return oapi.DeleteTaskUUIDCommentEntityUUIDFileFileUUID200Response{}, nil
}
