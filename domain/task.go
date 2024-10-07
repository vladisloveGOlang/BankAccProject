package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/samber/lo"
)

const (
	StatusUnknown    int = 0
	StatusNew        int = 1
	StatusInWork     int = 2
	StatusHold       int = 3
	StatusNeedReview int = 4
	StatusDone       int = 5
	StatusCancel     int = 6
)

func GetTaskStatuses() map[int]string {
	return map[int]string{
		StatusUnknown:    "Необработана",
		StatusNew:        "Новая",
		StatusInWork:     "В работе",
		StatusHold:       "Приостановлена",
		StatusNeedReview: "На проверке",
		StatusDone:       "Завершена",
		StatusCancel:     "Отменена",
	}
}

type Task struct {
	UUID           uuid.UUID
	ID             int
	Name           string    `validate:"lte=100,gte=3"  ru:"название"`
	Description    string    `validate:"lte=5000"  ru:"описание"`
	CreatedBy      string    `validate:"lte=100,gte=3"  ru:"автор (uuid)"`
	FederationUUID uuid.UUID `validate:"uuid"  ru:"федерация (uuid)"`
	CompanyUUID    uuid.UUID `validate:"uuid"  ru:"компания (uuid)"`
	ProjectUUID    uuid.UUID `validate:"uuid"  ru:"проект (uuid)"`

	IsEpic bool `ru:"эпик"`

	ResponsibleBy string `ru:"ответственный (uuid)"`
	ImplementBy   string `ru:"исполнитель (uuid)"`
	ManagedBy     string `ru:"постановщик (uuid)"`
	FinishedBy    string `ru:"завершил (uuid)"`

	CoWorkersBy []string
	WatchBy     []string

	People []string

	Tags        []string
	CompanyTags []Tag

	Icon     string
	Status   int
	Priority int `json:"priority"`

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ActivityAt time.Time
	DeletedAt  *time.Time
	Fields     map[string]interface{}
	RawFields  map[string]interface{}
	Meta       map[string]interface{}

	Path []string

	CommentsTotal int

	CacheExpires *time.Time

	TaskEntities map[uuid.UUID][]string

	Stops []Stop

	FinishTo   *time.Time
	FinishedAt *time.Time

	FirstOpen map[string]time.Time

	ChildrensTotal int
	ChildrensUUID  []uuid.UUID

	Activities      []Activity
	ActivitiesTotal int64

	Dirty map[string]interface{}
}

type Activity struct {
	UUID       uuid.UUID
	EntityUUID uuid.UUID

	EntityType  string
	Description string
	CreatedBy   User
	Meta        map[string]interface{}
	CreatedAt   time.Time

	Type int
}

type Stop struct {
	UUID          uuid.UUID `json:"uuid"`
	CreatedAt     time.Time `json:"created_at"`
	StatusID      int       `json:"status_id"`
	StatusName    string    `json:"status_name"`
	Comment       string    `json:"comment"`
	CreatedBy     string    `json:"created_by"`
	CreatedByUUID uuid.UUID `json:"created_by_uuid"`
}

type TaskEntity struct {
	UUID   uuid.UUID
	Fields []string
}

func NewTask(
	name string,
	federationUUID, companyUUID, projectUUID uuid.UUID,
	createdBy string,
	fields map[string]interface{},
	tags []string,
	description string,
	path []string,
	coworkersBy []string,
	implementedBy string,
	responsibleBy string,
	priority int,
	finishTo *time.Time,
	icon string,
	managedBy string,
	taskEntities map[uuid.UUID][]string,
) (task Task, err error) {
	uid := uuid.New()

	path = append(path, uid.String())
	path = helpers.Unique(path)

	coworkersBy = lo.WithoutEmpty(lo.Uniq(coworkersBy))
	path = lo.WithoutEmpty(lo.Uniq(path))
	tags = lo.WithoutEmpty(lo.Uniq(tags))

	allPeople := append([]string{}, createdBy, implementedBy, responsibleBy, managedBy)
	allPeople = append(allPeople, coworkersBy...)
	allPeople = lo.WithoutEmpty(lo.Uniq(allPeople))

	if len(path) == 0 {
		path = []string{task.UUID.String()}
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	task = Task{
		UUID:           uid,
		Name:           name,
		Description:    description,
		FederationUUID: federationUUID,
		CompanyUUID:    companyUUID,
		ProjectUUID:    projectUUID,
		CreatedBy:      createdBy,
		ManagedBy:      managedBy,
		RawFields:      fields,
		Fields:         make(map[string]interface{}),

		CreatedAt: time.Now(),

		Meta: make(map[string]interface{}),

		Path: path,
		Tags: tags,
		Icon: icon,

		CoWorkersBy:   coworkersBy,
		ImplementBy:   implementedBy,
		ResponsibleBy: responsibleBy,

		People: allPeople,

		Priority: priority,

		FinishTo: finishTo,

		TaskEntities: taskEntities,

		FirstOpen: make(map[string]time.Time),

		Dirty: make(map[string]interface{}),
	}

	errs, ok := helpers.ValidationStruct(task)
	if !ok {
		err = errors.New(helpers.Join(errs, ", "))
		return task, err
	}

	return task, err
}

func NewTaskByUUID(uid uuid.UUID) (task *Task) {
	task = &Task{UUID: uid}

	return task
}

func (t *Task) PatchName(name string) error {
	if len(name) < 3 || len(name) > 100 {
		return errors.New("название должно быть от 3 до 100 символов")
	}

	t.SafeDirty("name", t.Name)
	t.Name = name

	return nil
}

func (t *Task) PatchStatus(status int, opt ProjectOptions, comment string, sg *StatusGraph) ([]string, error) {
	if status == t.Status {
		return []string{}, errors.New("статус не изменился")
	}

	path := []string{}

	if sg == nil || len(sg.Graph) == 0 {
		sg = NewStatusGraph("0")
		sg.Graph = make(map[string][]string)
		sg.Graph["0"] = []string{"1"}
		sg.Graph["1"] = []string{"2"}
		sg.Graph["2"] = []string{"3", "4", "6"}
		sg.Graph["3"] = []string{"2"}
		sg.Graph["4"] = []string{"5", "2"}
		sg.Graph["5"] = []string{"2"}
		sg.Graph["6"] = []string{"2"}

	}

	// @todo
	sg.Current = fmt.Sprint(t.Status)

	allowMove, p := CheckPathByValue(sg, fmt.Sprint(t.Status), fmt.Sprint(status))

	if !allowMove {
		return path, fmt.Errorf("статус %v нельзя перевести в %v", t.Status, status)
	}

	path = p

	if status == StatusUnknown {
		return path, errors.New("статус нельзя перевести в необработано")
	}

	if status == StatusCancel && *opt.RequireCancelationComment && comment == "" {
		return path, errors.New("при отмене необходимо указать причину")
	}

	if status == StatusDone && *opt.RequireCancelationComment && comment == "" {
		return path, errors.New("при завершении необходимо указать причину")
	}

	if status < 0 || status > 10 {
		return path, errors.New("статус должен быть от 0 до 10")
	}

	t.SafeDirty("status", t.Status)
	t.Status = status

	return path, nil
}

type TaskManifest struct {
	Fields map[string]string // key - field name, value - field type: Приоритет: int
}

func (t *Task) SafeDirty(name string, val interface{}) {
	if t.Dirty == nil {
		t.Dirty = make(map[string]interface{})
	}
	t.Dirty[strings.ToLower(name)] = val
}
