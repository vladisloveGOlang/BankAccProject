package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/activities"
	"github.com/krisch/crm-backend/internal/comments"
	"github.com/krisch/crm-backend/internal/dictionary"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/profile"
	"github.com/krisch/crm-backend/internal/s3"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	repo           *Repository
	dict           *dictionary.Service
	as             *activities.Service
	ps             *profile.Service
	commentService *comments.Service
	storage        *s3.ServicePrivate

	ttlCache *ttlcache.Cache[string, []dto.TaskDTO]

	onTaskUpdatedOrCreated func(uuid.UUID, []string) error
	onOpenTask             func(uuid.UUID, string) error
}

func New(repo *Repository, dict *dictionary.Service, as *activities.Service, ps *profile.Service, cs *comments.Service, storage *s3.ServicePrivate) *Service {
	// @todo: rm profile service from here
	ttlCache := ttlcache.New[string, []dto.TaskDTO](
		ttlcache.WithTTL[string, []dto.TaskDTO](5 * time.Second),
	)
	go ttlCache.Start()

	return &Service{
		repo:           repo,
		dict:           dict,
		as:             as,
		ps:             ps,
		commentService: cs,
		storage:        storage,

		ttlCache: ttlCache,
	}
}

func (s *Service) TaskWasUpdatedOrCreated(uid uuid.UUID, people []string) error {
	if s.onTaskUpdatedOrCreated != nil {
		return s.onTaskUpdatedOrCreated(uid, people)
	}

	logrus.Error("onTaskUpdatedOrCreated is nil")

	return nil
}

func (s *Service) TaskWasOpen(uid uuid.UUID, email string) error {
	if s.onOpenTask != nil {
		return s.onOpenTask(uid, email)
	}

	logrus.Error("onOpenTask is nil")

	return nil
}

func (s *Service) CreateTask(task domain.Task) (id int, err error) {
	filteredFields, err := s.FilterTaskFields(task)
	if err != nil {
		return id, err
	}

	// @todo: filter task_entities fields by project

	task.Fields = filteredFields

	orm, err := s.repo.CreateTask(task, false)
	if err != nil {
		return id, err
	}

	path := task.Path
	if len(path) >= 2 {
		_, err = s.repo.UpdateChildTotal(uuid.MustParse(path[0]))
		if err != nil {
			return id, err
		}
	}

	if err == nil {
		notify := lo.Filter(task.People, func(email string, _ int) bool {
			return email != task.CreatedBy
		})

		err = s.TaskWasUpdatedOrCreated(task.UUID, notify)
		if err != nil {
			logrus.Error("TaskWasUpdatedOrCreated error: ", err)
		}
	}

	return orm.ID, err
}

func (s *Service) UpdateTask(crtr domain.Creator, task domain.Task, shouldUpdate []string) (err error) {
	filteredFields, err := s.FilterTaskFields(task)
	if err != nil {
		return err
	}

	for k, v := range filteredFields {
		task.Fields[k] = v
	}

	for k, v := range task.RawFields {
		if v == nil {
			delete(task.Fields, k)
		}
	}

	oldTask, err := s.GetTask(context.Background(), task.UUID, []string{})
	if err != nil {
		return err
	}

	err = s.repo.UpdateTask(task, shouldUpdate)
	if err != nil {
		return err
	}

	if err == nil {
		notify := lo.Filter(task.People, func(email string, _ int) bool {
			// @todo: delete me from notifications
			return email != crtr.Email
		})

		err = s.TaskWasUpdatedOrCreated(task.UUID, notify)
		if err != nil {
			logrus.Error("TaskWasUpdatedOrCreated error: ", err)
		}
	}

	for _, field := range shouldUpdate {
		// Get field using reflection

		tp := reflect.TypeOf(task)
		for i := 0; i < tp.NumField(); i++ {
			if strings.EqualFold(tp.Field(i).Name, field) {
				valNew := reflect.ValueOf(task).Field(i)
				valOld := reflect.ValueOf(oldTask).Field(i)
				_, err = s.as.TaskWasChangedActivity(crtr, task.UUID, field, valNew.Interface(), valOld.Interface())
				if err != nil {
					return err
				}
			}
		}
	}

	return err
}

func (s *Service) FilterTaskFields(task domain.Task) (filteredFields map[string]interface{}, err error) {
	filteredFields = make(map[string]interface{}, 0)

	if len(task.RawFields) > 0 {
		projectFields, err := s.repo.GetProjectFields(task.ProjectUUID)
		if err != nil {
			return filteredFields, err
		}

		addedFieldsHash := []string{}
		for _, pfield := range projectFields {
			if value, ok := task.RawFields[pfield.Hash]; ok {
				addedFieldsHash = append(addedFieldsHash, pfield.Hash)

				if value == nil {
					continue
				}

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
						return filteredFields, errors.New(msg)
					}
				case domain.Float:
					if v, ok := value.(float64); ok {
						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be float", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				case domain.String:
					if v, ok := value.(string); ok {
						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be string", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				case domain.Text:
					if v, ok := value.(string); ok {
						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be text", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				case domain.Bool:
					if v, ok := value.(bool); ok {
						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be bool", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				case domain.Switch:
					if v, ok := value.(float64); ok {
						if v == 0 || v == 1 || v == 2 {
							filteredFields[pfield.Hash] = int(v)
						} else {
							msg := fmt.Sprintf("field %s (%s) must be switch (0|1|2)", pfield.Name, pfield.Hash)
							return filteredFields, errors.New(msg)
						}
						continue
					} else {
						msg := fmt.Sprintf("field %s (%s) should be switch (0|1|2)", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
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
						return filteredFields, errors.New(msg)
					}
				case domain.Phone:
					if v, ok := value.(int); ok {
						filteredFields[pfield.Hash] = v
						continue
					}

					if v, ok := value.(float64); ok {
						filteredFields[pfield.Hash] = int(v)
					} else {
						msg := fmt.Sprintf("field %s (%s) should be integer (phone)", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				case domain.Link:
					if v, ok := value.(string); ok {
						// url: [text](url)
						exp := `^\[.*]\((http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/|\/|\/\/)?[A-z0-9_-]*?[:]?[A-z0-9_-]*?[@]?[A-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?\)$`
						//nolint
						match, err := regexp.MatchString(exp, v)
						if err != nil {
							return filteredFields, err
						}

						if !match {
							msg := fmt.Sprintf("field %s (%s) should be link [text](url)", pfield.Name, pfield.Hash)
							return filteredFields, errors.New(msg)
						}

						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be string", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				case domain.Email:
					if v, ok := value.(string); ok {
						exp := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
						//nolint
						match, err := regexp.MatchString(exp, v)
						if err != nil {
							return filteredFields, err
						}

						if !match {
							msg := fmt.Sprintf("field %s (%s) should be email", pfield.Name, pfield.Hash)
							return filteredFields, errors.New(msg)
						}

						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be string", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				case domain.Time:
					if v, ok := value.(string); ok {
						_, err := time.Parse(time.RFC3339, v)
						if err != nil {
							return filteredFields, err
						}

						filteredFields[pfield.Hash] = v[11:]
					} else {
						msg := fmt.Sprintf("field %s (%s) should be string", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				case domain.DateTime:
					if v, ok := value.(string); ok {
						_, err := time.Parse(time.RFC3339, v)
						if err != nil {
							return filteredFields, err
						}

						filteredFields[pfield.Hash] = v
					} else {
						msg := fmt.Sprintf("field %s (%s) should be string", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				case domain.People:
					rt := reflect.TypeOf(value)
					if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
						arrWithStrings := []string{}
						for _, i := range value.([]interface{}) {

							exp := `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
							//nolint
							match, err := regexp.MatchString(exp, fmt.Sprintf("%v", i))
							if err != nil {
								return filteredFields, err
							}

							if !match {
								msg := fmt.Sprintf("field %s (%s) - %s should be email", pfield.Name, pfield.Hash, fmt.Sprintf("%v", i))
								return filteredFields, errors.New(msg)
							}

							arrWithStrings = append(arrWithStrings, fmt.Sprintf("%v", i))
						}

						filteredFields[pfield.Hash] = arrWithStrings
					} else {
						msg := fmt.Sprintf("field %s (%s) should be array", pfield.Name, pfield.Hash)
						return filteredFields, errors.New(msg)
					}
				}
			}
		}

		if len(addedFieldsHash) != len(task.RawFields) {
			canBeAdded := addedFieldsHash
			sendedToAdd := helpers.GetMapKeys(task.RawFields)

			unwantedFields := helpers.ArrayNonIntersection(canBeAdded, sendedToAdd)

			if len(unwantedFields) == 0 {
				return filteredFields, errors.New("в проекте нет кастомных полей")
			}

			msg := fmt.Sprintf("невозможно добавить: (%s)", strings.Join(unwantedFields, ","))

			return filteredFields, errors.New(msg)
		}
	}

	return filteredFields, nil
}

func (s *Service) CreateTaskBatch(updaterEmail string, tasks []domain.Task) (err error) {
	err = s.repo.CreateInBatches(tasks)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if err == nil {
			notify := lo.Filter(task.People, func(email string, _ int) bool {
				// @todo: delete me from notifications
				return email != updaterEmail
			})

			err = s.TaskWasUpdatedOrCreated(task.UUID, notify)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func (s *Service) CreateTaskCustomStringField(task *domain.Task, name, value string) (uid string, err error) {
	errs, ok := helpers.ValidationStruct(task)
	if !ok {
		err = errors.New(helpers.Join(errs, ", "))
		return uid, err
	}

	err = s.repo.CreateTaskCustomStringField(*task, name, value)
	if err != nil {
		return uid, err
	}

	return uid, err
}

func (s *Service) CreateTaskCustomIntField(task *domain.Task, name string, value int) (uid string, err error) {
	errs, ok := helpers.ValidationStruct(task)
	if !ok {
		err = errors.New(helpers.Join(errs, ", "))
		return uid, err
	}

	err = s.repo.CreateTaskCustomIntField(*task, name, value)

	return uid, err
}

func (s *Service) CreateTaskCustomTextField(task *domain.Task, name, value string) (uid string, err error) {
	errs, ok := helpers.ValidationStruct(task)
	if !ok {
		err = errors.New(helpers.Join(errs, ", "))
		return uid, err
	}

	err = s.repo.CreateTaskCustomTextField(*task, name, value)

	return uid, err
}

func (s *Service) CreateTaskCustomBoolField(task *domain.Task, name string, value bool) (uid string, err error) {
	errs, ok := helpers.ValidationStruct(task)
	if !ok {
		err = errors.New(helpers.Join(errs, ", "))
		return uid, err
	}

	err = s.repo.CreateTaskCustomBoolField(*task, name, value)

	return uid, err
}

func (s *Service) GetTaskGetTaskWithDeleted(ctx context.Context, uid uuid.UUID) (dm domain.Task, err error) {
	dm, err = s.repo.GetTaskWithDeleted(ctx, uid)
	if err != nil {
		return dm, err
	}

	return dm, err
}

func (s *Service) GetTask(ctx context.Context, uid uuid.UUID, fields []string) (dm domain.Task, err error) {
	dm, err = s.repo.GetTask(ctx, uid)
	if err != nil {
		return dm, err
	}

	if email, ok := ctx.Value("userEmail").(string); ok {
		err := s.TaskWasOpen(uid, email)
		if err != nil {
			logrus.Error("TaskWasOpen error: ", err)
		}
	}

	if len(fields) > 0 {
		if lo.IndexOf(fields, "activities") != -1 {
			actvts, total, err := s.as.GetTaskActivities(uid, 200, 0)
			if err != nil {
				return dm, err
			}

			// @todo!
			dm.Activities = actvts
			dm.ActivitiesTotal = total
		}
	}

	return dm, err
}

func (s *Service) GetActivities(taskUUID uuid.UUID, limit, offset int) (dms []domain.Activity, total int64, err error) {
	dms, total, err = s.as.GetTaskActivities(taskUUID, limit, offset)

	return dms, total, err
}

func (s *Service) GetTasksNames(ctx context.Context, uid []uuid.UUID) (taskWithName []domain.Task, err error) {
	if len(uid) == 0 {
		return taskWithName, nil
	}

	return s.repo.GetTaskNames(ctx, uid)
}

func (s *Service) GetTasks(ctx context.Context, filter dto.TaskSearchDTO) (dm []domain.Task, total int64, err error) {
	allowSort := s.GetSortFields(filter.ProjectUUID)

	dm, total, err = s.repo.GetTasks(ctx, filter, allowSort)
	if err != nil {
		return dm, -1, err
	}

	return dm, total, err
}

func (s *Service) GetTasksDto(ctx context.Context, filter dto.TaskSearchDTO) (dtos []dto.TaskDTOs, total int64, err error) {
	dms, total, err := s.GetTasks(ctx, filter)
	dtos = []dto.TaskDTOs{}

	for _, dm := range dms {
		d, err := dto.NewTaskDTOs(dm, s.dict), err
		if err != nil {
			return dtos, -1, err
		}

		dtos = append(dtos, d)
	}

	return dtos, total, err
}

func (s *Service) ConvertToDto(dm domain.Task) (d dto.TaskDTO, err error) {
	return dto.NewTaskDTO(dm, []domain.Comment{}, []domain.File{}, []domain.Reminder{}, make(map[uuid.UUID]interface{}), s.dict, s.ps), err
}

func (s *Service) PatchProject(crt domain.Creator, task domain.Task, project dto.ProjectDTO, status int, comment string) (err error) {
	logrus.Warn(task.UUID, task.ProjectUUID)
	logrus.Warn(project.UUID)

	if task.FederationUUID != project.FederationUUID {
		return errors.New("невозможно переместить задачу в другую федерацию")
	}

	if task.CompanyUUID != project.CompanyUUID {
		return errors.New("невозможно переместить задачу в другую компанию")
	}

	if task.ProjectUUID == project.UUID {
		return errors.New("задача уже находится в этом проекте")
	}

	err = s.repo.gorm.DB.Transaction(func(tx *gorm.DB) error {
		err := s.repo.ChangeField(task.UUID, "project_uuid", project.UUID.String())
		if err != nil {
			return err
		}

		_, _, err = s.PatchStatus(crt, project, task, status, comment)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	_, err = s.as.TaskWasChangedActivity(crt, task.UUID, "project_uuid", task.ProjectUUID, project.UUID)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) PatchTaskParent(ctx context.Context, uid uuid.UUID, parentUUID *uuid.UUID) (err error) {
	if parentUUID != nil && uid == *parentUUID {
		return errors.New("uuid и parentUUID не могут быть одинаковыми")
	}

	task, err := s.repo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	if parentUUID != nil {
		parentTask, err := s.repo.GetTask(ctx, *parentUUID)
		if err != nil {
			return err
		}

		_, err = s.repo.GetParentLvl(parentTask.FederationUUID, parentTask.ProjectUUID, parentTask.UUID)
		if err != nil {
			return err
		}

		path := parentTask.Path
		path = append(path, uid.String())

		err = s.repo.ChangeField(task.UUID, "path", strings.Join(path, "."))
		if err != nil {
			return err
		}

		_, err = s.repo.UpdateChildTotal(uuid.MustParse(path[0]))
		if err != nil {
			return err
		}
	} else {

		err = s.repo.RemoveChildParent(task.FederationUUID, task.ProjectUUID, task.UUID.String())
		if err != nil {
			return err
		}

		_, err = s.repo.UpdateChildTotal(task.UUID)
		if err != nil {
			return err
		}

		_, err = s.repo.UpdateChildTotal(uuid.MustParse(task.Path[0]))
		if err != nil {
			return err
		}
	}

	// if parentLvl+len(task.Path) > 5 {
	// 	// @todo: fix
	// 	//return fmt.Errorf("вложенность не может быть больше 5. %v + %v > 5", parentLvl, len(task.Path))
	// }
	//
	// Parent A (a) (4 lvl)
	//    a
	//   / \
	//  b   c
	// / \	\
	//     \
	//      g
	//	     \
	//        h

	// Task T (t) (2 lvl)
	//    t
	//   / \
	//  y   n

	// Task G (g) (2 lvl)
	//    g
	//     \
	//      h

	return err
}

func (s *Service) PatchName(crt domain.Creator, uid uuid.UUID, name string) (err error) {
	task, err := s.GetTask(context.Background(), uid, []string{})
	if err != nil {
		return err
	}

	err = task.PatchName(name)
	if err != nil {
		return err
	}

	err = s.repo.ChangeField(task.UUID, "name", task.Name)

	if err == nil {
		notify := lo.Filter(task.People, func(email string, _ int) bool {
			return email != crt.Email
		})

		err = s.TaskWasUpdatedOrCreated(task.UUID, notify)
		if err != nil {
			return err
		}
	}

	_, err = s.as.TaskWasChangedActivity(crt, task.UUID, "name", task.Name, task.Dirty["name"])
	if err != nil {
		return err
	}

	return err
}

func (s *Service) PatchStatus(crtr domain.Creator, project dto.ProjectDTO, task domain.Task, status int, comment string) (stopUUID uuid.UUID, path []string, err error) {
	stopUUID = uuid.New()

	// @todo: mv to domain
	fields, _ := s.dict.FindProjectFields(task.ProjectUUID)
	for _, field := range fields {
		if field.RequiredOnStatuses != nil {
			if lo.IndexOf(field.RequiredOnStatuses, status) != -1 {
				if _, ok := task.Fields[field.Hash]; !ok {
					return stopUUID, path, fmt.Errorf("field %s (%s) is required", field.Name, field.Hash)
				}
			}
		}
	}

	sg, err := domain.NewStatusGraphFromMap(*project.StatusGraph)
	if err != nil {
		return stopUUID, path, err
	}

	path, err = task.PatchStatus(status, domain.ProjectOptions{
		RequireCancelationComment: project.Options.RequireCancelationComment,
		RequireDoneComment:        project.Options.RequireDoneComment,
		StatusEnable:              project.Options.StatusEnable,
	}, comment, sg)
	if err != nil {
		return stopUUID, path, err
	}

	err = s.repo.gorm.DB.Transaction(func(tx *gorm.DB) error {
		err = s.repo.ChangeField(task.UUID, "status", task.Status)

		if task.Status == domain.StatusDone {
			err = s.repo.ChangeField(task.UUID, "finished_at", time.Now())
			if err != nil {
				return err
			}

			err = s.repo.ChangeField(task.UUID, "finished_by", crtr.Email)
			if err != nil {
				return err
			}
		}

		stop := Stop{
			UUID:          stopUUID,
			CreatedAt:     time.Now(),
			StatusID:      task.Status,
			StatusName:    "todo",
			Comment:       comment,
			CreatedBy:     crtr.Email,
			CreatedByUUID: crtr.UUID,
		}

		err = tx.Exec("UPDATE tasks SET stops = stops::jsonb || ?  WHERE uuid = ?", stop, task.UUID).Error

		return err
	})

	if err == nil {
		notify := lo.Filter(task.People, func(email string, _ int) bool {
			return email != crtr.Email
		})

		err = s.TaskWasUpdatedOrCreated(task.UUID, notify)
		if err != nil {
			return stopUUID, path, err
		}
	}

	//
	if project.Statuses == nil {
		logrus.WithField("project_uuid", project.UUID).Error("projects statuses is nil")
		return stopUUID, path, errors.New("projects statuses is nil")
	}

	oldStatus, _ := lo.Find(*project.Statuses, func(item dto.ProjectStatusDTO) bool {
		return item.Number == task.Dirty["status"]
	})

	newStatus, _ := lo.Find(*project.Statuses, func(item dto.ProjectStatusDTO) bool {
		return item.Number == task.Status
	})

	_, err = s.as.TaskWasChangedStatusActivity(crtr, task.UUID, oldStatus.ToDTOs(), newStatus.ToDTOs())
	if err != nil {
		return stopUUID, path, err
	}

	return stopUUID, path, err
}

func (s *Service) PatchFirstOpenBy(ctx context.Context, uid, userUUID uuid.UUID) (err error) {
	task, err := s.GetTask(ctx, uid, []string{})
	if err != nil {
		return err
	}

	if _, ok := task.FirstOpen[userUUID.String()]; !ok {
		task.FirstOpen[userUUID.String()] = time.Now()
	}

	jsn, err := json.Marshal(task.FirstOpen)
	if err != nil {
		return err
	}

	err = s.repo.ChangeField(uid, "first_open", jsn)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) DeleteStop(_ context.Context, uid, stopUUID uuid.UUID, deletedBy string) (err error) {
	task, err := s.GetTask(context.Background(), uid, []string{})
	if err != nil {
		return err
	}

	user, _ := s.dict.FindUser(deletedBy)
	if user == nil {
		return errors.New("user not found")
	}

	stops := Stops{}

	for _, stop := range task.Stops {
		if stop.UUID != stopUUID {
			stops = append(stops, Stop{
				UUID:          stop.UUID,
				CreatedAt:     stop.CreatedAt,
				StatusID:      stop.StatusID,
				StatusName:    stop.StatusName,
				Comment:       stop.Comment,
				CreatedBy:     stop.CreatedBy,
				CreatedByUUID: stop.CreatedByUUID,
			})
		}
	}

	err = s.repo.ChangeField(task.UUID, "stops", stops)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) PatchTeam(ctx context.Context, crtr domain.Creator, uid uuid.UUID, implementedBy, responsibleBy *string, coworkersBy, watchedBy *[]string, managedBy *string) (err error) {
	people := []string{}

	if implementedBy != nil {
		people = append(people, *implementedBy)
	}
	if responsibleBy != nil {
		people = append(people, *responsibleBy)
	}
	if managedBy != nil {
		people = append(people, *managedBy)
	}
	if coworkersBy != nil {
		people = append(people, *coworkersBy...)
	}
	if watchedBy != nil {
		people = append(people, *watchedBy...)
	}

	// Validate team
	notFoundEmails := lo.Filter(people, func(email string, _ int) bool {
		if email == "" {
			return false
		}

		_, ok := s.dict.FindUser(email)
		return !ok
	})
	if len(notFoundEmails) > 0 {
		return fmt.Errorf("пользователи не найдены: %v", notFoundEmails)
	}

	task, err := s.GetTask(ctx, uid, []string{})
	if err != nil {
		return err
	}

	// @todo: to domain logic
	if implementedBy != nil {
		usersOld, _ := s.dict.FindUsers([]string{task.ImplementBy})
		users, _ := s.dict.FindUsers([]string{*implementedBy})

		err = s.repo.ChangeField(task.UUID, "implement_by", &implementedBy)
		if err != nil {
			return err
		}

		_, err = s.as.TaskWasChangedTeamActivity(crtr, task.UUID, "implement_by", usersOld, users)
		if err != nil {
			return err
		}

	}

	if responsibleBy != nil {
		usersOld, _ := s.dict.FindUsers([]string{task.ResponsibleBy})
		users, _ := s.dict.FindUsers([]string{*responsibleBy})

		err = s.repo.ChangeField(task.UUID, "responsible_by", *responsibleBy)
		if err != nil {
			return err
		}

		_, err = s.as.TaskWasChangedTeamActivity(crtr, task.UUID, "responsible_by", usersOld, users)
		if err != nil {
			return err
		}

	}

	if coworkersBy != nil {
		usersOld, _ := s.dict.FindUsers(task.CoWorkersBy)
		users, emails := s.dict.FindUsers(*coworkersBy)

		err = s.repo.ChangeField(task.UUID, "co_workers_by", &emails)
		if err != nil {
			return err
		}

		_, err = s.as.TaskWasChangedTeamActivity(
			crtr, task.UUID, "co_workers_by", usersOld, users,
		)
		if err != nil {
			return err
		}
	}

	if watchedBy != nil {
		usersOld, _ := s.dict.FindUsers(task.WatchBy)
		users, emails := s.dict.FindUsers(*watchedBy)

		err = s.repo.ChangeField(task.UUID, "watch_by", &emails)
		if err != nil {
			return err
		}

		_, err = s.as.TaskWasChangedTeamActivity(
			crtr, task.UUID, "watch_by", usersOld, users,
		)
		if err != nil {
			return err
		}
	}

	if managedBy != nil {
		usersOld, _ := s.dict.FindUsers([]string{task.ManagedBy})
		users, _ := s.dict.FindUsers([]string{*managedBy})

		err = s.repo.ChangeField(task.UUID, "managed_by", *managedBy)
		if err != nil {
			return err
		}

		_, err = s.as.TaskWasChangedTeamActivity(crtr, task.UUID, "managed_by", usersOld, users)
		if err != nil {
			return err
		}

	}

	err = s.repo.ChangeField(task.UUID, "all_people", &people)
	if err != nil {
		return err
	}

	// @todo: notify
	notify := lo.Filter(task.People, func(email string, _ int) bool {
		return email != crtr.Email
	})

	err = s.TaskWasUpdatedOrCreated(task.UUID, notify)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) CheckPath(path []string) (err error) {
	path = helpers.Unique(path)

	if len(path) == 0 {
		return nil
	}

	return s.repo.CheckPath(path)
}

func (s *Service) DeleteTask(crt domain.Creator, uid uuid.UUID) (err error) {
	t, err := s.GetTask(context.TODO(), uid, []string{})
	if err != nil {
		return err
	}

	err = s.repo.DeleteTask(uid)
	if err != nil {
		return err
	}

	// files
	files, err := s.storage.GetTaskFiles(t.UUID, false)
	if err != nil {
		return err
	}

	err = s.TaskWasUpdatedOrCreated(uid, t.People)
	if err != nil {
		return err
	}

	for _, file := range files {
		err := s.storage.Delete(file.UUID)
		if err != nil {
			return err
		}
	}

	// @todo: add action
	_, err = s.as.TaskWasDeleted(crt, t.UUID, t.Name)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) DeleteTaskFile(crt domain.Creator, taskUUID, fileUUID uuid.UUID) (err error) {
	files, err := s.storage.GetTaskFiles(taskUUID, false)
	if err != nil {
		return err
	}

	file, f := lo.Find(files, func(f domain.File) bool {
		return f.UUID == fileUUID
	})
	if !f {
		return errors.New("file not found")
	}

	err = s.storage.Delete(file.UUID)
	if err != nil {
		return err
	}

	s.ResetCache(taskUUID)

	_, err = s.as.TaskFileWasDeleted(crt, taskUUID, file)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) ResetCache(uid uuid.UUID) {
	s.repo.cache.ClearTask(context.TODO(), uid)
}

func (s *Service) GetSortFields(projectUUID uuid.UUID) []string {
	fields, _ := s.dict.FindCompanyFields(projectUUID)

	allowOrder := s.repo.GetSortFields()

	for _, field := range fields {
		if domain.FieldDataType(field.DataType) == domain.Integer || domain.FieldDataType(field.DataType) == domain.Float || domain.FieldDataType(field.DataType) == domain.String {
			// fileds.a || fields.b
			allowOrder = append(allowOrder, "fields."+field.Hash+"")
		}
	}

	return allowOrder
}
