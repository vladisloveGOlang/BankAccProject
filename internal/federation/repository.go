package federation

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/profile"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	gorm        *postgres.GDB
	rds         *redis.RDS
	middlewares []func(ctx context.Context, name string) error
}

func NewRepository(db *postgres.GDB, rds *redis.RDS) *Repository {
	return &Repository{
		gorm: db,
		rds:  rds,
	}
}

func (r *Repository) Use(fn func(ctx context.Context, name string) error) {
	r.middlewares = append(r.middlewares, fn)
}

func (r *Repository) apply(ctx context.Context, name string) func() {
	c := ctx
	return func() {
		for _, fn := range r.middlewares {
			err := fn(c, name)
			if err != nil {
				logrus.Error(err)
			}
		}
	}
}

func (r *Repository) PubUpdate() {
	err := r.rds.Publish(context.Background(), "update", "federation")
	logrus.Debug("pub update federation")
	if err != nil {
		logrus.Error(err)
	}
}

func (r *Repository) CreateFederation(federation domain.Federation) error {
	defer r.apply(context.Background(), "CreateFederation")

	orm := &Federation{}
	res := r.gorm.DB.Model(orm).
		Select("id").
		Where("name = ?", federation.Name).
		Where("created_by = ?", federation.CreatedBy).
		Where("deleted_at is null").
		First(&orm)

	if res.RowsAffected > 0 {
		return errors.New("федерация с таким именем уже существует")
	}

	orm = &Federation{
		UUID:           federation.UUID,
		Name:           federation.Name,
		CreatedBy:      federation.CreatedBy,
		CreatedByBUUID: federation.CreatedByUUID,
	}

	err := r.gorm.DB.Create(&orm).Error
	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) CreateCompany(company *domain.Company) (err error) {
	orm := &Company{}

	res := r.gorm.DB.Model(orm).
		Select("uuid").
		Where("name = ?", company.Name).
		Where("federation_uuid = ?", company.FederationUUID).
		Where("deleted_at is null").
		Find(&orm)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		return errors.New("компания с таким именем уже существует")
	}

	orm = &Company{
		UUID:           company.UUID,
		Name:           company.Name,
		FederationUUID: company.FederationUUID,
		CreatedBy:      company.CreatedBy,
		CreatedByUUID:  company.CreatedByUUID,
	}

	err = r.gorm.DB.Create(&orm).Error
	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) DeleteCompany(companyUUID uuid.UUID) error {
	res := r.gorm.DB.
		Model(&Company{}).
		Where("uuid = ?", companyUUID).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("компания не найдена")
	}

	if res.Error == nil {
		r.PubUpdate()
	}

	return res.Error
}

func (r *Repository) DeleteCompanies(federationUUID uuid.UUID) (int64, error) {
	res := r.gorm.DB.
		Model(&Company{}).
		Where("federation_uuid = ?", federationUUID).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.Error == nil {
		r.PubUpdate()
	}

	return res.RowsAffected, res.Error
}

func (r *Repository) DeleteFederation(uid uuid.UUID) error {
	res := r.gorm.DB.
		Model(&Federation{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("федерация не найдена")
	}

	if res.Error == nil {
		r.PubUpdate()
	}

	return res.Error
}

func (r *Repository) DeleteProject(uid string) error {
	err := r.gorm.DB.
		Model(&Project{}).
		Where("uuid = ?", uid).
		Update("deleted_at", "now()").
		Error
	if err == nil {
		r.PubUpdate()
	}
	return err
}

func (r *Repository) DeleteProjects(federationUUID, companyUUID *uuid.UUID) error {
	if federationUUID == nil && companyUUID == nil {
		return errors.New("нужно указать хотя бы один параметр (federationUUID | companyUUID)")
	}

	q := r.gorm.DB.
		Model(&Project{})

	if companyUUID != nil {
		q.Where("company_uuid = ?", companyUUID)
	}

	if federationUUID != nil {
		q.Where("federation_uuid = ?", federationUUID)
	}

	err := q.Update("deleted_at", "now()").Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) GetCompanyFederation(companyUUID uuid.UUID) (dt dto.FederationDTO, err error) {
	orm := Federation{}

	err = r.gorm.DB.Model(&orm).
		Joins("left join companies on companies.federation_uuid = federations.uuid").
		Where("companies.uuid = ?", companyUUID).
		Where("federations.deleted_at is null").
		First(&orm).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dt, errors.New("федерация компании не найдена")
	}

	if err != nil {
		return dt, err
	}

	return dto.FederationDTO{
		UUID: orm.UUID,
		Name: orm.Name,
	}, err
}

func (r *Repository) CreateProject(project *domain.Project) error {
	federation, err := r.GetFederation(project.FederationUUID)
	if err != nil {
		return err
	}

	if project.StatusSort == nil {
		project.StatusSort = []int{}
	}

	if project.FieldsSort == nil {
		project.FieldsSort = []string{}
	}

	orm := &Project{
		UUID:           project.UUID,
		Name:           project.Name,
		Description:    project.Description,
		FederationUUID: federation.UUID,
		CompanyUUID:    project.CompanyUUID,
		CreatedBy:      project.CreatedBy,
		Status:         project.StatusCode,
		StatusSort:     project.StatusSort,
		FieldsSort:     project.FieldsSort,
		Options:        project.Options,
		ResponsibleBy:  project.ResponsibleBy,
	}

	err = r.gorm.DB.Create(&orm).Error
	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) CreateCompanyField(cf *domain.CompanyField) (orm CompanyFields, err error) {
	err = r.gorm.DB.Transaction(func(tx *gorm.DB) error {
		company := &Company{}
		err = tx.Raw("select * from companies where uuid = ? FOR UPDATE", cf.CompanyUUID).Scan(&company).Error
		if err != nil {
			return err
		}

		if cf.RequiredOnStatuses == nil {
			cf.RequiredOnStatuses = []int{}
		}

		orm = CompanyFields{
			Name:        cf.Name,
			Description: cf.Description,
			Icon:        cf.Icon,
			DataType:    int(cf.DataType),
			Hash:        helpers.IntToLetters(company.FieldLastName + 1),
			CompanyUUID: cf.CompanyUUID,
		}

		err = tx.Create(&orm).Error
		if err != nil {
			return err
		}

		err = tx.Exec("update companies set updated_at = now(), field_last_name = field_last_name + 1 where uuid = ?", cf.CompanyUUID).Error
		if err == nil {
			r.PubUpdate()
		}

		return err
	})

	return orm, err
}

func (r *Repository) PutCompanyField(pf *domain.CompanyField) error {
	if pf.RequiredOnStatuses == nil {
		pf.RequiredOnStatuses = []int{}
	}

	orm := CompanyFields{
		Name:        pf.Name,
		Description: pf.Description,
		UUID:        pf.UUID,
		Icon:        pf.Icon,
	}

	err := r.gorm.DB.
		Model(&orm).
		Where("uuid = ?", pf.UUID).
		Update("name", orm.Name).
		Update("description", orm.Description).
		Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) GetProjectFields(projectUUID uuid.UUID) (orm []CompanyFields, err error) {
	orm = []CompanyFields{}

	r.gorm.DB.Model(&orm).
		Select("company_fields.uuid, company_fields.icon, company_fields.name, company_fields.description, company_fields.hash, company_fields.data_type, pf.style, pf.required_on_statuses").
		Joins("left join project_fields pf on pf.company_field_uuid = company_fields.uuid").
		Where("pf.project_uuid = ?", projectUUID).
		Where("company_fields.deleted_at is null").
		Where("pf.deleted_at is null").
		Find(&orm)

	if err != nil {
		return orm, err
	}

	return orm, err
}

func (r *Repository) GetCompanyFields(companyUUID uuid.UUID) (dmns []domain.CompanyField, err error) {
	orm := []CompanyFields{}

	// Company Fields
	res := r.gorm.DB.Model(&orm).
		Select("company_fields.uuid, company_fields.icon, company_fields.name, company_fields.description, company_fields.hash, company_fields.data_type, COALESCE(json_agg(distinct pf.project_uuid) FILTER (WHERE pf.project_uuid IS NOT NULL), '[]' ) as project_uuids,"+
			"count(*) as tasks_total,"+
			"count(*) FILTER (WHERE t.fields->>company_fields.hash is not null) as tasks_filled,"+
			"count(*) FILTER (WHERE t.fields->>company_fields.hash is not null and t.finished_at is null) as tasks_active_filled",
			"count(*) FILTER (WHERE t.finished_at is null) as tasks_active_total",
		).
		Where("company_fields.deleted_at is NULL").
		Where("company_fields.company_uuid", companyUUID).
		Joins("left join project_fields pf on pf.company_field_uuid = company_fields.uuid").
		Joins("left join tasks t on t.project_uuid = pf.project_uuid ").
		Group("company_fields.uuid, company_fields.icon, company_fields.name, company_fields.hash, company_fields.data_type, pf.style, pf.required_on_statuses").
		Find(&orm)
	if res.Error != nil {
		return dmns, res.Error
	}

	return lo.Map(orm, func(item CompanyFields, index int) domain.CompanyField {
		return domain.CompanyField{
			UUID:        item.UUID,
			Hash:        item.Hash,
			Name:        item.Name,
			Description: item.Description,
			Icon:        item.Icon,
			DataType:    domain.FieldDataType(item.DataType),
			CompanyUUID: item.CompanyUUID,
			ProjectUUID: lo.Map(item.ProjectUUID, func(uid any, index int) uuid.UUID {
				return uuid.MustParse(uid.(string))
			}),

			TasksTotal:        item.TasksTotal,
			TasksFilled:       item.TasksFilled,
			TasksActiveTotal:  item.TasksActiveTotal,
			TasksActiveFilled: item.TasksActiveFilled,
		}
	}), nil
}

func (r *Repository) DeleteCompanyField(uid uuid.UUID) (err error) {
	res := r.gorm.DB.Model(&CompanyFields{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("запись не найдена")
	}

	return res.Error
}

func (r *Repository) ChangeCompanyField(uid uuid.UUID, fieldName string, value interface{}) error {
	err := r.gorm.DB.
		Model(&Company{}).
		Where("uuid = ?", uid).
		Update(fieldName, value).
		Update("updated_at", "now()").
		Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) GetProject(uid uuid.UUID) (orm Project, err error) {
	err = r.gorm.DB.Model(&orm).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Select("*, coalesce( (stops[0]->>'created_at')::timestamptz, created_at ) as status_updated_at").
		First(&orm).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return orm, dto.NotFoundErr("проект не найден")
	}

	if err != nil {
		return orm, err
	}

	return orm, err
}

func (r *Repository) GetProjects(federationUID uuid.UUID) (orm []Project, err error) {
	err = r.gorm.DB.Model(&orm).
		Where("federation_uuid = ?", federationUID).
		Where("deleted_at is null").
		Select("*").
		Find(&orm).Error

	if err != nil {
		return orm, err
	}

	return orm, err
}

func (r *Repository) ProjectStatistic(companyUID, uid uuid.UUID) (orm ProjectStatistic, fs []FieldStatistics, err error) {
	err = r.gorm.DB.Model(&orm).
		Table("projects").
		Select(
			"projects.uuid, count(distinct tasks.uuid) as tasks_total,"+
				" count(distinct tasks.uuid) FILTER (WHERE tasks.deleted_at is NULL AND tasks.status != 3 AND tasks.status != 5 AND tasks.status != 6 AND tasks.status != 6) as tasks_active_total, "+
				" count(distinct tasks.uuid) FILTER (WHERE tasks.deleted_at is NULL AND tasks.status = 6) as tasks_canceled_total, "+
				" count(distinct tasks.uuid) FILTER (WHERE tasks.deleted_at is NULL AND tasks.finished_at IS NOT NULL) as tasks_finished_total, "+
				" count(distinct tasks.uuid) FILTER (WHERE tasks.deleted_at is not NULL) as tasks_deleted_total ").
		Where("projects.uuid = ?", uid).
		Joins("left join tasks on tasks.project_uuid = projects.uuid").
		Where("projects.deleted_at is null").
		Group("projects.uuid").
		First(&orm).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return orm, fs, dto.NotFoundErr("проект не найден")
	}
	if err != nil {
		return orm, fs, err
	}

	// Statistics for Company Fields
	fieldStatistics := []FieldStatistics{}
	res := r.gorm.DB.Raw("with t as (select count(*) as total from tasks where project_uuid = ?) select (select name from company_fields cf where cf.hash = key and company_uuid = ?  limit 1 ) as name, key as hash, count(*), t.total, round(count(*)::decimal / t.total * 100, 2) as filled from tasks, jsonb_object_keys(fields) AS key left join t on 1 = 1 where tasks.project_uuid = ? AND tasks.fields->>key != '' group by key, jsonb_object_keys(fields), t.total", uid, companyUID, uid).Scan(&fieldStatistics)
	if res.Error != nil {
		return orm, fieldStatistics, res.Error
	}

	return orm, fieldStatistics, err
}

func (r *Repository) GetProjectsByCompany(uid uuid.UUID) (dmns []domain.Project, err error) {
	orm := []Project{}

	err = r.gorm.DB.Model(&orm).
		Where("company_uuid = ?", uid).
		Where("deleted_at is null").
		Find(&orm).Error

	if err != nil {
		return dmns, err
	}

	for _, item := range orm {
		dmns = append(dmns, domain.Project{
			UUID:           item.UUID,
			Name:           item.Name,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
			CreatedBy:      item.CreatedBy,
			CreatedAt:      item.CreatedAt,
			StatusCode:     item.Status,
			UpdatedAt:      item.UpdatedAt,
			DeletedAt:      item.DeletedAt,
		})
	}

	return dmns, err
}

func (r *Repository) GetProjectsByUser(_ context.Context, uid uuid.UUID) (dmns []domain.Project, err error) {
	orm := []Project{}

	err = r.gorm.DB.Model(&orm).
		Where("fu.federation_uuid = projects.federation_uuid and (fu.user_uuid = ? or cu.user_uuid = ?)", uid, uid).
		Select("projects.uuid, projects.status, projects.name, projects.description, projects.federation_uuid, projects.company_uuid, projects.created_at, projects.updated_at, projects.deleted_at").
		Where("projects.deleted_at is null").
		Group("projects.uuid, projects.name, projects.federation_uuid, projects.company_uuid, projects.created_at, projects.updated_at, projects.deleted_at").
		Joins("left join company_users cu on cu.company_uuid = projects.company_uuid").
		Joins("left join federation_users fu on fu.federation_uuid = projects.federation_uuid").
		Order("projects.created_at desc").
		Find(&orm).Error

	if err != nil {
		return dmns, err
	}

	for _, item := range orm {
		dmns = append(dmns, domain.Project{
			UUID:           item.UUID,
			Name:           item.Name,
			Description:    item.Description,
			StatusCode:     item.Status,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
			CreatedBy:      item.CreatedBy,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			DeletedAt:      item.DeletedAt,
		})
	}

	return dmns, err
}

func (r *Repository) AddUser(fu domain.FederationUser) (err error) {
	existingRecord := &FederationUser{}

	res := r.gorm.DB.
		Select("uuid").
		Where("federation_uuid = ?", fu.FederationUUID).
		Where("user_uuid = ?", fu.User.UUID).
		Find(&existingRecord)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		return errors.New("пользователь уже добавлен")
	}

	existingRecord = &FederationUser{
		FederationUUID: fu.FederationUUID,
		UserUUID:       fu.User.UUID,
		UUID:           uuid.New(),
	}

	err = r.gorm.DB.
		Create(&existingRecord).Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) DeleteUser(federationUUID, userUUID uuid.UUID) (err error) {
	orm := &FederationUser{}

	err = r.gorm.DB.
		Where("federation_uuid", federationUUID).
		Where("user_uuid", userUUID).
		Delete(orm).
		Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) AddUserToCompany(cu domain.CompanyUser) (err error) {
	existingRecord := &CompanyUser{}

	res := r.gorm.DB.
		Where("company_uuid = ?", cu.CompanyUUID).
		Where("federation_uuid = ?", cu.FederationUUID).
		Where("user_uuid = ?", cu.User.UUID).
		Find(&existingRecord)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		return errors.New("пользователь уже добавлен в компанию")
	}

	existingRecord = &CompanyUser{
		FederationUUID: cu.FederationUUID,
		CompanyUUID:    cu.CompanyUUID,
		UserUUID:       cu.User.UUID,
		UUID:           cu.UUID,
	}

	err = r.gorm.DB.
		Create(&existingRecord).Error
	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) DeleteUserFromCompany(companyUUID, userUUID uuid.UUID) (err error) {
	orm := &CompanyUser{}

	err = r.gorm.DB.
		Where("company_uuid", companyUUID).
		Where("user_uuid", userUUID).
		Delete(orm).
		Error

	if err == nil {
		r.PubUpdate()
	}
	return err
}

func (r *Repository) GetFederationsByUser(userUUID uuid.UUID) (items []domain.Federation, err error) {
	orm := []Federation{}

	r.gorm.DB.Model(&orm).
		Select("federations.uuid, federations.name", "federations.created_at").
		Joins("left join federation_users fu on fu.federation_uuid = federations.uuid").
		Where("federations.deleted_at is null").
		Where("fu.user_uuid = ?", userUUID).
		Find(&orm)

	for _, item := range orm {
		items = append(items, domain.Federation{
			UUID:      item.UUID,
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
		})
	}

	return items, err
}

func (r *Repository) GetFederation(uid uuid.UUID) (item domain.Federation, err error) {
	orm := Federation{}

	err = r.gorm.DB.Model(&orm).
		Select("federations.uuid", "federations.name", "federations.created_at", "federations.deleted_at").
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		First(&orm).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, dto.NotFoundErr("федерация не найдена")
	}

	return domain.Federation{
		UUID:      orm.UUID,
		Name:      orm.Name,
		CreatedAt: orm.CreatedAt,
		DeletedAt: orm.DeletedAt,
	}, err
}

func (r *Repository) GetFederations() ([]domain.Federation, error) {
	orm := []Federation{}

	err := r.gorm.DB.Model(&orm).
		Select("federations.uuid", "federations.name", "federations.created_at", "federations.deleted_at").
		Find(&orm).
		Error
	if err != nil {
		return nil, err
	}

	return lo.Map(orm, func(item Federation, index int) domain.Federation {
		return domain.Federation{
			UUID:      item.UUID,
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
			DeletedAt: item.DeletedAt,
		}
	}), err
}

func (r *Repository) GetCompany(uid uuid.UUID) (dm domain.Company, err error) {
	orm := Company{}

	err = r.gorm.DB.Model(&orm).
		Select("companies.uuid, companies.name, companies.federation_uuid, companies.created_at, companies.updated_at").
		Where("companies.uuid = ?", uid).
		Where("companies.deleted_at is null").
		First(&orm).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dm, errors.New("компания не найдена")
	}

	if err != nil {
		return dm, err
	}

	return domain.Company{
		UUID:           orm.UUID,
		Name:           orm.Name,
		FederationUUID: orm.FederationUUID,
		CreatedAt:      orm.CreatedAt,
		UpdatedAt:      orm.UpdatedAt,
	}, err
}

func (r *Repository) GetCompaniesByUser(userUUID uuid.UUID) (items []domain.Company, err error) {
	orm := []Company{}

	r.gorm.DB.Model(&orm).
		Select("companies.uuid, companies.name, companies.federation_uuid").
		Joins("left join company_users cu on cu.company_uuid = companies.uuid").
		Where("cu.user_uuid = ?", userUUID).
		Where("companies.deleted_at is null").
		Find(&orm)

	if err != nil {
		return items, err
	}

	items = lo.Map(orm, func(item Company, index int) domain.Company {
		return domain.Company{
			UUID:           item.UUID,
			Name:           item.Name,
			FederationUUID: item.FederationUUID,
		}
	})

	return items, err
}

func (r *Repository) GetCompaniesByFederation(federationUUID uuid.UUID) (items []domain.Company, err error) {
	orm := []Company{}

	r.gorm.DB.Model(&orm).
		Select("companies.uuid, companies.name, companies.federation_uuid,  companies.created_at,  companies.updated_at, companies.deleted_at, count(*) FILTER(where cu.uuid is not null) as users_total, count(*) FILTER(where pr.uuid is not null) as projects_total").
		Where("companies.federation_uuid = ?", federationUUID).
		Where("companies.deleted_at is null").
		Joins("left join company_users cu on cu.company_uuid = companies.uuid").
		Joins("left join projects pr on pr.company_uuid = companies.uuid").
		Group("companies.uuid, companies.name, companies.federation_uuid, companies.created_at, companies.updated_at, companies.deleted_at").
		Find(&orm)

	if err != nil {
		return items, err
	}

	items = lo.Map(orm, func(item Company, index int) domain.Company {
		return domain.Company{
			UUID:           item.UUID,
			Name:           item.Name,
			FederationUUID: item.FederationUUID,
			UserTotal:      item.UsersTotal,
			ProjectsTotal:  item.ProjectsTotal,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			DeletedAt:      item.DeletedAt,
		}
	})

	return items, err
}

func (r *Repository) SearchUser(search domain.SearchUser) (dmns []domain.User, err error) {
	str := strings.ToLower(search.Search)

	orm := []profile.User{}

	searchPurified := strings.ReplaceAll(str, "%", "") + "%"

	query := r.gorm.DB.Model(&orm).
		Select("users.uuid, users.name, users.lname, users.pname, users.phone, users.has_photo, users.email").
		Joins("left join federation_users fu on fu.user_uuid = users.uuid").
		Where("fu.federation_uuid = ?", search.FederationUUID).
		Limit(50).
		Where("users.email ilike ? or users.name ilike ? or users.lname ilike ? or users.pname ilike ?", searchPurified, searchPurified, searchPurified, searchPurified)

	if search.CompanyUUID != nil {
		query = query.Joins("left join company_users cu on cu.user_uuid = users.uuid").
			Where("cu.company_uuid = ?", search.CompanyUUID)
	}

	err = query.Find(&orm).
		Error

	if err != nil {
		return dmns, err
	}

	for _, item := range orm {
		dmns = append(dmns, domain.User{
			UUID:     item.UUID,
			Name:     item.Name,
			Lname:    item.Lname,
			Pname:    item.Pname,
			Phone:    item.Phone,
			Email:    item.Email,
			HasPhoto: item.HasPhoto,
		})
	}

	return dmns, err
}

func (r *Repository) GetFederationUsers(uid uuid.UUID) (dms []domain.FederationUser, err error) {
	orm := []User{}

	err = r.gorm.DB.Model(&orm).
		Select("users.uuid, users.name, users.lname, users.pname, users.phone, users.email, users.is_valid, users.has_photo, fu.created_at as added_to_federation").
		Joins("left join federation_users fu on fu.user_uuid = users.uuid").
		Where("fu.federation_uuid = ?", uid).
		Where("fu.deleted_at is null").
		Find(&orm).
		Error

	if err != nil {
		return dms, err
	}

	for _, item := range orm {
		dms = append(dms, domain.FederationUser{
			UUID: item.UUID,
			User: domain.User{
				UUID:     item.UUID,
				Name:     item.Name,
				Lname:    item.Lname,
				Pname:    item.Pname,
				Phone:    item.Phone,
				Email:    item.Email,
				IsValid:  item.IsValid,
				HasPhoto: item.HasPhoto,
			},
			AddedAt:        *item.AddedToFederation,
			FederationUUID: uid,
		})
	}

	return dms, err
}

func (r *Repository) GetCompanyUsers(uid uuid.UUID) (dms []domain.CompanyUser, err error) {
	orm := []User{}

	err = r.gorm.DB.Model(&orm).
		Select("users.uuid, users.name, users.lname, users.pname, users.phone, users.email, users.is_valid, users.has_photo, cu.created_at as added_to_company").
		Joins("left join company_users cu on cu.user_uuid = users.uuid").
		Where("cu.deleted_at is null").
		Where("cu.company_uuid = ?", uid).
		Find(&orm).
		Error

	if err != nil {
		return dms, err
	}

	for _, item := range orm {
		dms = append(dms, domain.CompanyUser{
			UUID: item.UUID,
			User: domain.User{
				UUID:  item.UUID,
				Name:  item.Name,
				Lname: item.Lname,
				Pname: item.Pname,
				Phone: item.Phone,
				Email: item.Email,
			},
			AddedAt:        *item.AddedToCompany,
			FederationUUID: uid,
		})
	}

	return dms, err
}

func (r *Repository) GetProjectUsers(uid uuid.UUID) (dms []domain.ProjectUser, err error) {
	orm := []User{}

	err = r.gorm.DB.Model(&orm).
		Select("users.uuid, users.name, users.lname, users.pname, users.phone, users.email, users.is_valid, users.has_photo, pu.created_at as added_to_project").
		Joins("left join project_users pu on pu.user_uuid = users.uuid").
		Where("pu.project_uuid = ?", uid).
		Where("pu.deleted_at is null").
		Find(&orm).
		Error

	if err != nil {
		return dms, err
	}

	for _, item := range orm {
		dms = append(dms, domain.ProjectUser{
			UUID: item.UUID,
			User: domain.User{
				UUID:  item.UUID,
				Name:  item.Name,
				Lname: item.Lname,
				Pname: item.Pname,
				Phone: item.Phone,
				Email: item.Email,
			},
			AddedAt:        *item.AddedToProject,
			FederationUUID: uid,
		})
	}

	return dms, err
}

func (r *Repository) ChangeField(uid uuid.UUID, fieldName string, value interface{}) error {
	err := r.gorm.DB.
		Model(&Federation{}).
		Where("uuid = ?", uid).
		Update(fieldName, value).
		Update("updated_at", "now()").
		Error
	if err == nil {
		r.PubUpdate()
	}
	return err
}

func (r *Repository) ChangeProjectField(uid uuid.UUID, fieldName string, value interface{}) error {
	err := r.gorm.DB.
		Model(&Project{}).
		Where("uuid = ?", uid).
		Update(fieldName, value).
		Update("updated_at", "now()").
		Error

	if err == nil {
		r.PubUpdate()
	}
	return err
}

func (r *Repository) CreateInvite(invite *domain.Invite) error {
	existingRecord := &Invite{}

	res := r.gorm.DB.
		Where("federation_uuid = ?", invite.FederationUUID).
		Where("company_uuid", invite.CompanyUUID).
		Where("email = ?", invite.Email).
		Where("(deleted_at is NULL and declined_at is NULL and accepted_at is NULL)").
		Find(&existingRecord)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		return errors.New("приглашение уже отправлено")
	}

	orm := &Invite{
		FederationUUID: invite.FederationUUID,
		CompanyUUID:    invite.CompanyUUID,
		Email:          invite.Email,
		UUID:           invite.UUID,
	}

	err := r.gorm.DB.
		Create(&orm).Error

	return err
}

func (r *Repository) GetInvites(federationUUID uuid.UUID) (items []domain.Invite, err error) {
	orm := []Invite{}

	err = r.gorm.DB.Model(&orm).
		Where("federation_uuid = ?", federationUUID).
		Where("deleted_at is null").
		Find(&orm).Error

	if err != nil {
		return items, err
	}

	items = lo.Map(orm, func(item Invite, index int) domain.Invite {
		return domain.Invite{
			UUID:           item.UUID,
			Email:          item.Email,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
			AcceptedAt:     item.AcceptedAt,
			DeclinedAt:     item.DeclinedAt,
			CreatedAt:      item.CreatedAt,
		}
	})

	return items, err
}

func (r *Repository) DeleteInvite(uid uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&Invite{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null ").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("запись не найдена")
	}

	return res.Error
}

func (r *Repository) AddUserToProject(cu *domain.ProjectUser) (err error) {
	existingRecord := &ProjectUser{}

	res := r.gorm.DB.
		Where("project_uuid = ?", cu.ProjectUUID).
		Where("user_uuid = ?", cu.User.UUID).
		Where("deleted_at is null").
		Find(&existingRecord)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		return errors.New("пользователь уже добавлен в проект")
	}

	existingRecord = &ProjectUser{
		ProjectUUID: cu.ProjectUUID,
		UserUUID:    cu.User.UUID,
		UUID:        cu.UUID,
	}

	err = r.gorm.DB.Create(&existingRecord).Error

	return err
}

func (r *Repository) DeleteUserFromProject(projectUUID, userUUID uuid.UUID) (err error) {
	orm := &ProjectUser{}

	res := r.gorm.DB.
		Model(&orm).
		Where("project_uuid", projectUUID).
		Where("user_uuid", userUUID).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("запись не найдена")
	}

	return res.Error
}

func (r *Repository) CreateCatalogData(cd domain.ProjectCatalogData) (err error) {
	orm := &ProjectCatalogData{
		FederationUUID: cd.FederationUUID,
		CompanyUUID:    cd.CompanyUUID,
		ProjectUUID:    cd.ProjectUUID,
		UUID:           cd.UUID,
		Name:           string(cd.Name),
		Value:          cd.Value,
	}

	err = r.gorm.DB.Create(&orm).Error

	return err
}

func (r *Repository) GetProjectCatalogData(uid uuid.UUID, catalogName *string) (dtos []domain.ProjectCatalogData, err error) {
	orm := []ProjectCatalogData{}

	q := r.gorm.DB.Model(&orm).
		Where("project_uuid = ?", uid).
		Where("deleted_at is null")

	if catalogName != nil {
		q = q.Where("name = ?", *catalogName)
	}

	err = q.Find(&orm).Error

	if err != nil {
		return dtos, err
	}

	for _, item := range orm {
		dtos = append(dtos, domain.ProjectCatalogData{
			UUID:           item.UUID,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
			ProjectUUID:    item.ProjectUUID,
			Name:           domain.ProjectCatalogType(item.Name),
			Value:          item.Value,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			DeletedAt:      item.DeletedAt,
		})
	}

	return dtos, err
}

func (r *Repository) GetCompanyProjectCatalogData(companyUUID uuid.UUID, catalogName string) (dtos []domain.ProjectCatalogData, err error) {
	orm := []ProjectCatalogData{}

	err = r.gorm.DB.Model(&orm).
		Where("company_uuid = ?", companyUUID).
		Where("name = ?", catalogName).
		Where("deleted_at is null").
		Find(&orm).
		Error

	if err != nil {
		return dtos, err
	}

	for _, item := range orm {
		dtos = append(dtos, domain.ProjectCatalogData{
			UUID:           item.UUID,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
			ProjectUUID:    item.ProjectUUID,
			Name:           domain.ProjectCatalogType(item.Name),
			Value:          item.Value,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			DeletedAt:      item.DeletedAt,
		})
	}

	return dtos, err
}

func (r *Repository) DeleteCatalogData(uid uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&ProjectCatalogData{}).
		Where("uuid", uid).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("запись не найдена")
	}

	return res.Error
}

func (r *Repository) CreateGroup(group *domain.Group) (err error) {
	orm := &Group{}

	res := r.gorm.DB.Model(orm).
		Select("uuid").
		Where("name = ?", group.Name).
		Where("company_uuid = ?", group.CompanyUUID).
		Where("deleted_at is null").
		Find(&orm)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		return errors.New("группа с таким именем уже существует")
	}

	orm = &Group{
		UUID:           group.UUID,
		Name:           group.Name,
		FederationUUID: group.FederationUUID,
		CompanyUUID:    group.CompanyUUID,
	}

	err = r.gorm.DB.Create(&orm).Error

	return err
}

func (r *Repository) GetGroup(uid uuid.UUID) (dm domain.Group, err error) {
	orm := Group{}

	err = r.gorm.DB.Model(&orm).
		Select("*").
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		First(&orm).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dm, errors.New("группа не найдена")
	}

	if err != nil {
		return dm, err
	}

	return domain.Group{
		UUID:           orm.UUID,
		Name:           orm.Name,
		FederationUUID: orm.FederationUUID,
		CompanyUUID:    orm.CompanyUUID,
		CreatedAt:      orm.CreatedAt,
		UpdatedAt:      orm.UpdatedAt,
	}, err
}

func (r *Repository) GetCompanyGroups(companyUUID uuid.UUID) (dmns []domain.Group, err error) {
	orm := []Group{}

	err = r.gorm.DB.Model(&orm).
		Select("groups.uuid, groups.name, groups.federation_uuid, groups.company_uuid, groups.created_at, groups.updated_at, groups.deleted_at, coalesce( json_agg(gu.user_uuid) FILTER (WHERE gu.user_uuid is not null), '[]' ) as user_uuids").
		Joins("LEFT JOIN group_users gu ON gu.group_uuid = groups.uuid").
		Where("groups.company_uuid = ?", companyUUID).
		Where("groups.deleted_at is null").
		Group("groups.uuid, groups.name, groups.federation_uuid, groups.company_uuid, groups.created_at, groups.updated_at, groups.deleted_at").
		Find(&orm).Error

	if err != nil {
		return dmns, err
	}

	return lo.Map(orm, func(item Group, index int) domain.Group {
		return domain.Group{
			UUID:           item.UUID,
			Name:           item.Name,
			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			DeletedAt:      item.DeletedAt,
			UsersUUIDS:     item.UserUuids,
		}
	}), nil
}

func (r *Repository) ChangeGroupField(uid uuid.UUID, fieldName string, value interface{}) error {
	err := r.gorm.DB.
		Model(&Group{}).
		Where("uuid = ?", uid).
		Update(fieldName, value).
		Update("updated_at", "now()").
		Error
	if err == nil {
		r.PubUpdate()
	}
	return err
}

func (r *Repository) DeleteGroup(uid uuid.UUID) error {
	res := r.gorm.DB.
		Model(&Group{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("группа не найдена")
	}

	return res.Error
}

func (r *Repository) AddUserToGroup(userUUID, groupUUID uuid.UUID, createdBy string, createdByUUID uuid.UUID) (err error) {
	existingRecord := &GroupUser{}

	res := r.gorm.DB.
		Where("group_uuid = ?", groupUUID).
		Where("user_uuid = ?", userUUID).
		Find(&existingRecord)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		return errors.New("пользователь уже добавлен в группу")
	}

	existingRecord = &GroupUser{
		UUID:          uuid.New(),
		GroupUUID:     groupUUID,
		UserUUID:      userUUID,
		CreatedBy:     createdBy,
		CreatedByUUID: createdByUUID,
	}

	err = r.gorm.DB.Create(&existingRecord).Error
	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) GetGroupUsers(groupUUID uuid.UUID) (dms []domain.User, err error) {
	orm := []User{}

	err = r.gorm.DB.Model(&orm).
		Select("users.*").
		Joins("left join group_users gu on gu.user_uuid = users.uuid").
		Where("gu.group_uuid = ?", groupUUID).
		Where("gu.deleted_at is null").
		Where("users.deleted_at is null").
		Find(&orm).
		Error

	if err != nil {
		return dms, err
	}

	return lo.Map(orm, func(item User, index int) domain.User {
		return domain.User{
			UUID:     item.UUID,
			Name:     item.Name,
			Lname:    item.Lname,
			Pname:    item.Pname,
			Phone:    item.Phone,
			Email:    item.Email,
			IsValid:  item.IsValid,
			HasPhoto: item.HasPhoto,
		}
	}), nil
}

func (r *Repository) GetUserGroups(userUUID uuid.UUID) (dms []domain.Group, err error) {
	orm := []Group{}

	err = r.gorm.DB.Model(&orm).
		Select("groups.*").
		Joins("left join group_users gu on gu.group_uuid = groups.uuid").
		Where("gu.user_uuid = ?", userUUID).
		Where("gu.deleted_at is null").
		Find(&orm).
		Error

	if err != nil {
		return dms, err
	}

	return lo.Map(orm, func(item Group, index int) domain.Group {
		return domain.Group{
			UUID: item.UUID,
			Name: item.Name,
		}
	}), nil
}

func (r *Repository) GetUsersGroups(userUUIDs []uuid.UUID) (dms map[uuid.UUID][]domain.Group, err error) {
	dms = make(map[uuid.UUID][]domain.Group)
	// @todo: to one db request
	for _, uid := range userUUIDs {
		dm, err := r.GetUserGroups(uid)
		if err != nil {
			return dms, err
		}

		dms[uid] = dm
	}

	return dms, err
}

func (r *Repository) RemoveUserFromGroups(groupUUID, userUUID uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&GroupUser{}).
		Where("group_uuid", groupUUID).
		Where("user_uuid", userUUID).
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("пользователя нет в группе")
	}

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) AddProjectField(projectUUID, companyUUID, companyFieldUUID uuid.UUID, requiredOnStatuses []int, style string) (err error) {
	existingRecord := &ProjectField{}

	if style != "" && style != "hide_when_empty" && style != "show_when_empty" {
		return errors.New("неверный стиль hide_when_empty|always_hide")
	}

	res := r.gorm.DB.
		Where("project_uuid", projectUUID).
		Where("company_field_uuid", companyFieldUUID).
		Where("deleted_at is null").
		Find(&existingRecord)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 &&
		style == existingRecord.Style &&
		helpers.EquelSlices(existingRecord.RequiredOnStatuses, requiredOnStatuses) {
		return errors.New("поле уже добавлено")
	}

	if res.RowsAffected > 0 {
		err = r.gorm.DB.
			Model(&ProjectField{}).
			Where("project_uuid", projectUUID).
			Where("company_field_uuid", companyFieldUUID).
			Where("deleted_at is null").
			Update("required_on_statuses", IntArray(requiredOnStatuses)).
			Update("style", style).
			Error
	} else {
		existingRecord = &ProjectField{
			UUID:               uuid.New(),
			CompanyUUID:        companyUUID,
			ProjectUUID:        projectUUID,
			CompanyFieldUUID:   companyFieldUUID,
			RequiredOnStatuses: requiredOnStatuses,
			Style:              style,
		}

		err = r.gorm.DB.Create(&existingRecord).Error
	}
	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) RemoveProjectField(projectUUID, companyFieldUUID uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&ProjectField{}).
		Where("project_uuid", projectUUID).
		Where("company_field_uuid", companyFieldUUID).
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("пользователя нет в группе")
	}

	if err == nil {
		r.PubUpdate()
	}

	return err
}

///

func (r *Repository) CreateProjectStatus(cp domain.ProjectStatus) (err error) {
	orm := &ProjectStatus{}
	res := r.gorm.DB.Model(orm).
		Where("number = ?", cp.Number).
		Where("company_uuid = ?", cp.CompanyUUID).
		Where("project_uuid = ?", cp.ProjectUUID).
		Where("deleted_at is null").
		First(&orm)

	if res.RowsAffected > 0 {
		return errors.New("статус с таким номером уже существует")
	}

	orm = &ProjectStatus{
		UUID:        cp.UUID,
		Name:        cp.Name,
		Number:      cp.Number,
		Color:       cp.Color,
		CompanyUUID: cp.CompanyUUID,
		ProjectUUID: cp.ProjectUUID,
		Description: cp.Description,
	}

	err = r.gorm.DB.Create(&orm).Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) GetProjectStatus(uid uuid.UUID) (orm ProjectStatus, err error) {
	err = r.gorm.DB.
		Where("uid = ?", uid).
		Where("deleted_at is null").
		First(&orm).
		Error
	return orm, err
}

func (r *Repository) GetProjectStatuses(projectUUID uuid.UUID) (tags []ProjectStatus, err error) {
	orm := []ProjectStatus{}
	err = r.gorm.DB.
		Model(&orm).
		Where("project_uuid = ?", projectUUID).
		Where("deleted_at is null").
		Find(&orm).
		Error

	return orm, err
}

func (r *Repository) UpdateProjectStatus(uid uuid.UUID, name, color, description string) (err error) {
	res := r.gorm.DB.
		Model(&ProjectStatus{}).
		Where("uuid = ?", uid).
		Update("name", name).
		Update("color", color).
		Update("description", description).
		Where("deleted_at is null").
		Update("updated_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("статус удален или не найден")
	}

	if res.Error == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) DeleteProjectStatus(uid uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&ProjectStatus{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("статус не найден или уже удален")
	}

	if res.Error == nil {
		r.PubUpdate()
	}

	return err
}
