package task

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/cache"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	gorm      *postgres.GDB
	rds       *redis.RDS
	histogram *prometheus.HistogramVec
	cache     *cache.Service
}

func NewRepository(db *postgres.GDB, rds *redis.RDS, metrics *helpers.MetricsCounters, cs *cache.Service) *Repository {
	return &Repository{
		gorm:      db,
		rds:       rds,
		histogram: metrics.RepoHistogram,
		cache:     cs,
	}
}

func (r *Repository) CreateTask(task domain.Task, batch bool) (orm *Task, err error) {
	defer r.storeTime("CreateTask", tm())

	orm = &Task{
		UUID: task.UUID,
		Name: task.Name,

		CreatedBy:     task.CreatedBy,
		ResponsibleBy: task.ResponsibleBy,
		ImplementBy:   task.ImplementBy,
		CoWorkersBy:   task.CoWorkersBy,
		ManagedBy:     task.ManagedBy,
		WatchBy:       task.WatchBy,

		IsEpic:         task.IsEpic,
		FederationUUID: task.FederationUUID,
		CompanyUUID:    task.CompanyUUID,
		ProjectUUID:    task.ProjectUUID,

		CreatedAt: task.CreatedAt,

		Fields: task.Fields,

		Tags:      task.Tags,
		Icon:      task.Icon,
		AllPeople: task.People,

		Path: strings.Join(task.Path, "."),

		Priority: task.Priority,

		TaskEntities: task.TaskEntities,

		FinishTo: task.FinishTo,

		FirstOpen: task.FirstOpen,

		Description: task.Description,
	}

	if !batch {
		err = r.gorm.DB.Transaction(func(tx *gorm.DB) error {
			project := &Project{}
			err = tx.Raw("select * from projects where uuid = ? FOR UPDATE", task.ProjectUUID).Scan(&project).Error
			if err != nil {
				return err
			}

			task.ID = project.TaskID + 1
			orm.ID = task.ID

			err = tx.Create(&orm).Error
			if err != nil {
				return err
			}

			err = tx.Exec("update projects set task_id = ? where uuid = ?", task.ID, project.UUID).Error

			return err
		})
	}

	return orm, err
}

func (r *Repository) CreateInBatches(task []domain.Task) (err error) {
	defer r.storeTime("CreateInBatches", tm())

	taskOrms := []Task{}

	for _, item := range task {
		taskItem := item
		orm, err := r.CreateTask(taskItem, true)
		if err != nil {
			return err
		}

		taskOrms = append(taskOrms, *orm)
	}

	projectTasks := lo.GroupBy(taskOrms, func(item Task) string {
		return item.ProjectUUID.String()
	})

	projectUUIDs := lo.Keys(projectTasks)

	for _, projectUUID := range projectUUIDs {
		err = r.gorm.DB.Transaction(func(tx *gorm.DB) error {
			project := &Project{}
			err = tx.Raw("select uuid, task_id from projects where uuid = ? FOR UPDATE", projectUUID).Scan(&project).Error
			if err != nil {
				return err
			}

			taskID := project.TaskID + 1
			for i := range projectTasks[projectUUID] {
				projectTasks[projectUUID][i].ID = taskID
				taskID++
			}

			err := tx.CreateInBatches(projectTasks[projectUUID], 200).Error
			if err != nil {
				return err
			}

			err = tx.Exec("update projects set task_id = ? where uuid = ?", taskID, project.UUID).Error

			return err
		})

		if err != nil {
			return err
		}
	}

	return err
}

func (r *Repository) GetProjectFields(projectUUID uuid.UUID) (orm []CompanyFields, err error) {
	orm = []CompanyFields{}

	err = r.gorm.DB.Model(&orm).
		Joins("left join project_fields pf on pf.company_field_uuid = company_fields.uuid").
		Where("pf.project_uuid = ?", projectUUID).
		Where("company_fields.deleted_at is null").
		Where("pf.deleted_at is null").
		Find(&orm).
		Error

	if err != nil {
		return orm, err
	}

	return orm, err
}

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	valueString, err := json.Marshal(j)
	return string(valueString), err
}

func (j *JSONB) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &j); err != nil {
		return err
	}
	return nil
}

func tm() *helpers.Time {
	return helpers.NewTime()
}

func (r *Repository) CreateTaskCustomStringField(task domain.Task, name, value string) error {
	defer r.storeTime("CreateTaskCustomStringField", tm())

	mp := make(map[string]string)
	mp[name] = value
	js, err := json.Marshal(mp)
	if err != nil {
		return err
	}

	err = r.gorm.DB.Exec(`UPDATE tasks set fields = fields || ? where uuid = ?`, js, task.UUID).Error

	if err == nil {
		go r.ResetCache(task.UUID)
	}

	return err
}

func (r *Repository) CreateTaskCustomTextField(task domain.Task, name, value string) error {
	mp := make(map[string]string)
	mp[name] = value
	js, err := json.Marshal(mp)
	if err != nil {
		return err
	}

	err = r.gorm.DB.Exec(`UPDATE tasks set fields = fields || ? where uuid = ?`, js, task.UUID).Error

	if err == nil {
		go r.ResetCache(task.UUID)
	}

	return err
}

func (r *Repository) CreateTaskCustomBoolField(task domain.Task, name string, value bool) error {
	mp := make(map[string]bool)
	mp[name] = value
	js, err := json.Marshal(mp)
	if err != nil {
		return err
	}

	err = r.gorm.DB.Exec(`UPDATE tasks set fields = fields || ? where uuid = ?`, js, task.UUID).Error

	if err == nil {
		go r.ResetCache(task.UUID)
	}

	return err
}

func (r *Repository) CreateTaskCustomIntField(task domain.Task, name string, value int) error {
	mp := make(map[string]int)
	mp[name] = value
	js, err := json.Marshal(mp)
	if err != nil {
		return err
	}

	err = r.gorm.DB.Exec(`UPDATE tasks set fields = fields || ? where uuid = ?`, js, task.UUID).Error

	if err == nil {
		go r.ResetCache(task.UUID)
	}

	return err
}

func (r *Repository) GetTask(_ context.Context, uid uuid.UUID) (dm domain.Task, err error) {
	defer r.storeTime("GetTask", tm())

	orm := &Task{}
	err = r.gorm.DB.
		Where("uuid = ?", uid.String()).
		Where("deleted_at is null").
		Take(&orm).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dm, dto.NotFoundErr("задача не найдена")
	}

	if err != nil {
		return dm, err
	}

	dm = domain.Task{
		UUID:           orm.UUID,
		Name:           orm.Name,
		ID:             orm.ID,
		ProjectUUID:    orm.ProjectUUID,
		CompanyUUID:    orm.CompanyUUID,
		FederationUUID: orm.FederationUUID,

		Priority: orm.Priority,
		IsEpic:   orm.IsEpic,

		Path: strings.Split(orm.Path, "."),

		Description: orm.Description,

		CreatedBy:     orm.CreatedBy,
		CoWorkersBy:   orm.CoWorkersBy,
		WatchBy:       orm.WatchBy,
		ResponsibleBy: orm.ResponsibleBy,
		ImplementBy:   orm.ImplementBy,
		ManagedBy:     orm.ManagedBy,
		FinishedBy:    orm.FinishedBy,

		People: orm.AllPeople,

		Icon:   orm.Icon,
		Tags:   orm.Tags,
		Status: orm.Status,
		Fields: orm.Fields,

		CommentsTotal: orm.CommentsTotal,

		CreatedAt:  orm.CreatedAt,
		UpdatedAt:  orm.UpdatedAt,
		ActivityAt: orm.ActivityAt,
		DeletedAt:  orm.DeletedAt,

		TaskEntities: orm.TaskEntities,

		Stops: lo.Map(orm.Stops, func(item Stop, _ int) domain.Stop {
			return domain.Stop{
				UUID:          item.UUID,
				CreatedAt:     item.CreatedAt,
				StatusID:      item.StatusID,
				StatusName:    item.StatusName,
				Comment:       item.Comment,
				CreatedBy:     item.CreatedBy,
				CreatedByUUID: item.CreatedByUUID,
			}
		}),

		FinishTo:   orm.FinishTo,
		FinishedAt: orm.FinishedAt,

		FirstOpen: orm.FirstOpen,

		ChildrensTotal: orm.ChildrensTotal,
		ChildrensUUID: lo.Map(orm.ChildrensUUID, func(item string, _ int) uuid.UUID {
			return uuid.MustParse(item)
		}),
	}

	return dm, nil
}

func (r *Repository) GetTaskWithDeleted(_ context.Context, uid uuid.UUID) (dm domain.Task, err error) {
	defer r.storeTime("GetTaskWithDeleted", tm())

	orm := &Task{}
	err = r.gorm.DB.
		Where("uuid = ?", uid.String()).
		Take(&orm).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dm, dto.NotFoundErr("задача не найдена")
	}

	if err != nil {
		return dm, err
	}

	dm = domain.Task{
		UUID:           orm.UUID,
		Name:           orm.Name,
		ID:             orm.ID,
		ProjectUUID:    orm.ProjectUUID,
		FederationUUID: orm.FederationUUID,

		Priority: orm.Priority,
		IsEpic:   orm.IsEpic,

		Path: strings.Split(orm.Path, "."),

		Description: orm.Description,

		CreatedBy:     orm.CreatedBy,
		CoWorkersBy:   orm.CoWorkersBy,
		WatchBy:       orm.WatchBy,
		ResponsibleBy: orm.ResponsibleBy,
		ImplementBy:   orm.ImplementBy,
		ManagedBy:     orm.ManagedBy,
		FinishedBy:    orm.FinishedBy,

		People: orm.AllPeople,

		Icon:   orm.Icon,
		Tags:   orm.Tags,
		Status: orm.Status,
		Fields: orm.Fields,

		CommentsTotal: orm.CommentsTotal,

		CreatedAt:  orm.CreatedAt,
		UpdatedAt:  orm.UpdatedAt,
		ActivityAt: orm.ActivityAt,
		DeletedAt:  orm.DeletedAt,

		TaskEntities: orm.TaskEntities,

		Stops: lo.Map(orm.Stops, func(item Stop, _ int) domain.Stop {
			return domain.Stop{
				UUID:          item.UUID,
				CreatedAt:     item.CreatedAt,
				StatusID:      item.StatusID,
				StatusName:    item.StatusName,
				Comment:       item.Comment,
				CreatedBy:     item.CreatedBy,
				CreatedByUUID: item.CreatedByUUID,
			}
		}),

		FinishTo:   orm.FinishTo,
		FinishedAt: orm.FinishedAt,

		FirstOpen: orm.FirstOpen,
	}

	return dm, nil
}

func (r *Repository) GetTaskNames(_ context.Context, uids []uuid.UUID) (taskWithName []domain.Task, err error) {
	defer r.storeTime("GetTaskNames", tm())

	err = r.gorm.DB.
		Model(&Task{}).
		Select("uuid, name").
		Where("uuid in ?", uids).
		Where("deleted_at is null").
		Find(&taskWithName).
		Error

	if err != nil {
		return taskWithName, err
	}

	return taskWithName, nil
}

func (r *Repository) GetSortFields() []string {
	st := reflect.TypeOf(Task{})

	allowSort := []string{}
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)

		d := field.Tag

		if strings.Contains(string(d), "order") {
			allowSort = append(allowSort, helpers.ToLowerSnake(field.Name))
		}
	}

	return allowSort
}

func (r *Repository) GetTasks(_ context.Context, filter dto.TaskSearchDTO, allowSort []string) (dms []domain.Task, total int64, err error) {
	defer r.storeTime("GetTasks", tm())

	orms := []Task{}

	query := r.gorm.DB

	if len(allowSort) > 0 && filter.Order != nil {
		by := "desc"
		if filter.By != nil && *filter.By == "asc" {
			by = "asc"
		}

		if helpers.InArray(*filter.Order, allowSort) {
			if strings.HasPrefix(*filter.Order, "fields.") {
				*filter.Order = "fields->>'" + strings.Replace(*filter.Order, "fields.", "", 1) + "'"
			}

			query = query.Order(*filter.Order + " " + by)
		}
	} else {
		query = query.Order("created_at desc")
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	query = query.Where("project_uuid = ?", filter.ProjectUUID)

	query = query.Where("federation_uuid = ?", filter.FederationUUID)

	if filter.Participated != nil && len(*filter.Participated) > 0 {
		for _, item := range *filter.Participated {
			query = query.Where("? = ANY (all_people)", item)
		}
	}

	if filter.Tags != nil && len(*filter.Tags) > 0 {
		for _, item := range *filter.Tags {
			query = query.Where("? = ANY (tags)", item)
		}
	}

	if filter.Name != nil && len(*filter.Name) >= 1 {
		// WHERE  name iLIKE 'My%' OR name LIKE '% My%'
		query = query.Where("name iLIKE ? OR name iLIKE ?", *filter.Name+"%", "% "+*filter.Name+"%")
	}

	if filter.IsMy != nil && *filter.IsMy && filter.MyEmail != nil {
		query = query.Where("? = ANY (all_people)", filter.MyEmail)
	}

	if filter.IsEpic != nil {
		if *filter.IsEpic {
			query = query.Where("nlevel(path) = 1")
		}
	}

	if len(filter.Fields) > 0 {
		for _, item := range filter.Fields {
			logrus.Warn(item.Value)
			logrus.Warn(item.Value)
			// @todo: add regular to check array
			if strings.HasPrefix(fmt.Sprintf("%v", item.Value), "@> [") && strings.HasSuffix(fmt.Sprintf("%v", item.Value), "]") {
				v := strings.TrimPrefix(item.Value.(string), "@> ")
				query = query.Where(" fields->? @> ?", item.Name, v)
			} else {
				query = query.Where("? = fields->>? ", item.Value, item.Name)
			}
		}
	}

	if filter.Path != nil {
		query = query.Where("path ~ ?", *filter.Path)
	}

	if filter.Limit != nil {
		query = query.Limit(*filter.Limit)
	} else {
		query = query.Limit(5)
	}

	if filter.Offset != nil {
		query = query.Offset(*filter.Offset)
	} else {
		query = query.Offset(0)
	}

	query = query.Where("deleted_at is null")

	query = query.Select("*, count(*) OVER() AS total")

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&orms)
	})

	logrus.Debug("sql: ", sql)

	result := query.Find(&orms)

	if result.Error != nil {
		return dms, -1, result.Error
	}

	if len(orms) > 0 {
		total = orms[0].Total
	}

	dms = helpers.Map(orms, func(item Task, i int) domain.Task {
		return domain.Task{
			UUID:           item.UUID,
			Name:           item.Name,
			ID:             item.ID,
			ProjectUUID:    item.ProjectUUID,
			FederationUUID: item.FederationUUID,
			Priority:       item.Priority,
			IsEpic:         item.IsEpic,
			CreatedBy:      item.CreatedBy,
			CoWorkersBy:    item.CoWorkersBy,
			WatchBy:        item.WatchBy,
			ResponsibleBy:  item.ResponsibleBy,
			ImplementBy:    item.ImplementBy,
			Tags:           item.Tags,
			Status:         item.Status,
			Fields:         item.Fields,

			ActivityAt:     item.ActivityAt,
			ChildrensTotal: item.ChildrensTotal,
			FinishTo:       item.FinishTo,
			FinishedAt:     item.FinishedAt,

			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	})

	return dms, total, nil
}

func (r *Repository) ChangeField(uid uuid.UUID, fieldName string, value interface{}) error {
	defer r.storeTime("ChangeField", tm())

	res := r.gorm.DB.
		Model(&Task{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update(fieldName, value).
		Update("activity_at", "now()").
		Update("updated_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("нельзя обновлять удаленную задачу")
	}

	if res.Error == nil {
		go r.ResetCache(uid)
	}

	return res.Error
}

func (r *Repository) storeTime(name string, t *helpers.Time) {
	func() { r.histogram.WithLabelValues(name).Observe(t.Secondsf()) }()
}

func (r *Repository) GetParentLvl(federationUUID, projectUUID, uid uuid.UUID) (maxLvl int, err error) {
	maxLvl = 1
	err = r.gorm.DB.Raw("select  COALESCE(max(nlevel(path)),1) as maxLvl from tasks where federation_uuid = ? and project_uuid = ? and path ~ ? ", federationUUID, projectUUID, uid.String()+".*").Scan(&maxLvl).Error

	return maxLvl, err
}

func (r *Repository) RemoveChildParent(federationUUID, projectUUID uuid.UUID, taskUUID string) (err error) {
	// a,b,c,d
	// a,b,c,d,e
	// a,b

	// a,b -> z,b

	// z,b,c,d
	// z,b,c,d,e
	// z,b

	err = r.gorm.DB.Exec("update tasks set path = subpath(path, index(path, ?)) where federation_uuid = ? and project_uuid = ? and path ~ ? ", taskUUID, federationUUID, projectUUID, "*."+taskUUID+".*").Error

	// @total: 1 fix reset all caches
	// if err != nil {
	// 	r.ResetCache(uuid)
	// }

	return err
}

type taskCildrens struct {
	Total int64
	U     UuidsArray
}

type UuidsArray []uuid.UUID

func (j *UuidsArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := []uuid.UUID{}
	err := json.Unmarshal(bytes, &result)
	*j = result
	return err
}

func (j UuidsArray) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (r *Repository) UpdateChildTotal(taskUUID uuid.UUID) (mp map[uuid.UUID]int64, err error) {
	orm := []Task{}

	err = r.gorm.DB.
		Model(&Task{}).
		Where("path ~ ?", taskUUID.String()+".*").
		Where("deleted_at is null").
		Find(&orm).Error

	if err != nil {
		return mp, err
	}

	uuids := []uuid.UUID{}
	for _, item := range orm {
		uuids = append(uuids, item.UUID)
	}
	uuids = lo.Uniq(uuids)

	mp = make(map[uuid.UUID]int64)

	for _, u := range uuids {
		taskCildrens := taskCildrens{}

		err = r.gorm.DB.
			Model(&Task{}).
			Select("count(*) as total, json_agg(uuid) as u").
			Where("path ~ ?", "*."+u.String()+".*").
			Where("uuid != ?", u).
			Where("deleted_at is null").
			Find(&taskCildrens).Error

		if err != nil {
			return mp, err
		}

		mp[u] = taskCildrens.Total

		childUuids := lo.Map(taskCildrens.U, func(item uuid.UUID, _ int) string {
			return item.String()
		})

		err := r.gorm.DB.Exec("update tasks set childrens_total = ?, childrens_uuid = ? where uuid = ?", taskCildrens.Total, "{"+strings.Join(childUuids, ",")+"}", u).Error
		if err != nil {
			return mp, err
		}

		go r.ResetCache(u)
	}

	// a.b.c.d #d
	// a.b.c.d.e.z #z
	// a.i.o.p #p
	// a.i
	// a.i.o
	// a.b.c
	// a.b
	// a
	// a.b.c.d.e

	return mp, err
}

func (r *Repository) UpdateTask(task domain.Task, shouldUpdate []string) (err error) {
	for _, item := range shouldUpdate {
		switch item {
		case "priority":
			err = r.ChangeField(task.UUID, "priority", task.Priority)
		case "tags":
			err = r.ChangeField(task.UUID, "tags", "{"+strings.Join(task.Tags, ",")+"}")
		case "fields":
			err = r.ChangeField(task.UUID, "fields", task.Fields)
		case "finish_to":
			err = r.ChangeField(task.UUID, "finish_to", task.FinishTo)
		case "description":
			err = r.ChangeField(task.UUID, "description", task.Description)
		}
	}

	return err
}

func (r *Repository) CheckPath(path []string) (err error) {
	count := 0
	err = r.gorm.DB.Raw("select count(uuid) as count from tasks where uuid in (?)", path).Scan(&count).Error

	if count != len(path) {
		return fmt.Errorf("один (или все) uuid не существует (%v)", path)
	}
	return err
}

func (r *Repository) DeleteTask(uid uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&Task{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("задача не найдена")
	}

	if res.Error == nil {
		go r.ResetCache(uid)
	}

	return res.Error
}

func (r *Repository) ResetCache(uid uuid.UUID) {
	r.cache.ClearTask(context.TODO(), uid)
}
