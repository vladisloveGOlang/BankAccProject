package catalogs

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	gorm        *postgres.GDB
	rds         *redis.RDS
	histogram   *prometheus.HistogramVec
	middlewares []func(ctx context.Context, name string) error
}

func NewRepository(db *postgres.GDB, rds *redis.RDS, metrics *helpers.MetricsCounters) *Repository {
	return &Repository{
		gorm:      db,
		rds:       rds,
		histogram: metrics.RepoHistogram,
	}
}

func (r *Repository) PubUpdate() {
	err := r.rds.Publish(context.Background(), "update", "catalogs")
	logrus.Debug("pub update federation")
	if err != nil {
		logrus.Error(err)
	}
}

func (r *Repository) storeTime(name string, t *helpers.Time) {
	func() { r.histogram.WithLabelValues(name).Observe(t.Secondsf()) }()
}

func tm() *helpers.Time {
	return helpers.NewTime()
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

func (r *Repository) CreateCatalog(dir *domain.Catalog) error {
	defer r.apply(context.Background(), "CreateCatalog")

	orm := &Catalog{
		UUID:          dir.UUID,
		Name:          dir.Name,
		CreatedBy:     dir.CreatedBy,
		CreatedByUUID: dir.CreatedByUUID,

		FederationUUID: dir.FederationUUID,
		CompanyUUID:    dir.CompanyUUID,
	}

	err := r.gorm.DB.Create(orm).Error
	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) DeleteCatalog(uid uuid.UUID) error {
	res := r.gorm.DB.
		Model(&Catalog{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("запись не найдена")
	}

	if res.Error == nil {
		r.PubUpdate()
	}

	return res.Error
}

func (r *Repository) CreateCatalogField(pf *domain.CatalogFiled) (orm CatalogFields, err error) {
	if len(pf.Hash) >= 3 {
		r.gorm.DB.Model(&CatalogFields{}).
			Where("hash = ?", pf.Hash).
			Where("catalog_uuid = ?", pf.CatalogUUID).
			First(&orm)

		if orm.Hash != "" && orm.DeletedAt == nil {
			return orm, fmt.Errorf("запись (%s) уже существует", pf.Hash)
		}

		if orm.Hash != "" {
			err = r.gorm.DB.Model(&orm).
				UpdateColumn("deleted_at", nil).
				Error

			return orm, err
		}

		orm = CatalogFields{
			Name:            pf.Name,
			DataType:        int(pf.DataType),
			DataCatalogUUID: pf.DataCatalogUUID,
			Hash:            pf.Hash,
			CatalogUUID:     pf.CatalogUUID,
		}

		err = r.gorm.DB.Create(&orm).Error
		if err != nil {
			return orm, err
		}

		err = r.gorm.DB.Exec("update catalogs set updated_at = NOW() where uuid = ?", pf.CatalogUUID).Error
		if err != nil {
			return orm, err
		}

		if err == nil {
			r.PubUpdate()
		}

		return orm, err
	}

	err = r.gorm.DB.Transaction(func(tx *gorm.DB) error {
		catalog := &Catalog{}
		err = tx.Raw("select * from catalogs where uuid = ? FOR UPDATE", pf.CatalogUUID).Scan(&catalog).Error
		if err != nil {
			return err
		}

		orm = CatalogFields{
			Name:            pf.Name,
			DataType:        int(pf.DataType),
			DataCatalogUUID: pf.DataCatalogUUID,
			Hash:            helpers.IntToLetters(catalog.FieldLastName + 1),
			CatalogUUID:     pf.CatalogUUID,
		}

		err = tx.Create(&orm).Error
		if err != nil {
			return err
		}

		err = tx.Exec("update catalogs set updated_at = NOW(), field_last_name = field_last_name + 1 where uuid = ?", pf.CatalogUUID).Error

		return err
	})

	if err == nil {
		r.PubUpdate()
	}

	return orm, err
}

func (r *Repository) PutCatalogField(pf *domain.CatalogFiled) error {
	orm := CatalogFields{
		Name: pf.Name,
		UUID: pf.UUID,
	}

	err := r.gorm.DB.
		Model(&orm).
		Where("uuid = ?", pf.UUID).
		Update("name", orm.Name).
		Error
	if err != nil {
		return err
	}

	err = r.gorm.DB.Exec("update catalogs set updated_at = NOW() where uuid = ?", pf.CatalogUUID).Error
	if err != nil {
		return err
	}

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) GetCatalogFields(catalogUUID uuid.UUID) (df []domain.CatalogFiled, err error) {
	orm := []CatalogFields{}

	r.gorm.DB.Model(&orm).
		Where("catalog_uuid = ?", catalogUUID).
		Where("deleted_at is null").
		Find(&orm)

	if err != nil {
		return df, err
	}

	df = lo.Map(orm, func(item CatalogFields, index int) domain.CatalogFiled {
		return domain.CatalogFiled{
			UUID:            item.UUID,
			Name:            item.Name,
			DataType:        domain.FieldDataType(item.DataType),
			DataCatalogUUID: item.DataCatalogUUID,
			Hash:            item.Hash,
		}
	})

	return df, err
}

func (r *Repository) DeletecatalogField(uid uuid.UUID) (err error) {
	orm := CatalogFields{}

	r.gorm.DB.Model(&orm).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Find(&orm)

	if orm.UUID == uuid.Nil {
		return dto.NotFoundErr("запись не найдена")
	}

	res := r.gorm.DB.Model(&orm).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Where("deleted_at is null").
		Update("deleted_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("запись не обновлена")
	}

	err = r.gorm.DB.Exec("update catalogs set updated_at = NOW() where uuid = ?", orm.CatalogUUID).Error
	if err != nil {
		return err
	}

	if err == nil {
		r.PubUpdate()
	}

	return res.Error
}

func (r *Repository) GetCatalog(uid uuid.UUID) (orm Catalog, err error) {
	err = r.gorm.DB.Model(&orm).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		First(&orm).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return orm, dto.NotFoundErr("справочник не найден")
	}

	return orm, err
}

func (r *Repository) GetCatalogsByCompany(uid uuid.UUID) (orm []Catalog, err error) {
	err = r.gorm.DB.Model(&orm).
		Where("company_uuid = ?", uid).
		Where("deleted_at is null").
		Find(&orm).Error

	if err != nil {
		return orm, err
	}

	return orm, err
}

func (r *Repository) ChangecatalogField(uid uuid.UUID, fieldName string, value interface{}) error {
	err := r.gorm.DB.
		Model(&Catalog{}).
		Where("uuid = ?", uid).
		Update(fieldName, value).
		Update("updated_at", "now()").
		Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) AddData(dm domain.CatalogData) (orm CatalogData, err error) {
	keys := helpers.GetMapKeys(dm.Entities)

	entitesInDB := []string{}

	ent := []any{}

	if len(keys) > 0 {
		r.gorm.DB.Model(&CatalogData{}).
			Select("uuid").
			Where("uuid IN ?", keys).
			Where("federation_uuid = ?", dm.FederationUUID).
			Where("company_uuid = ?", dm.CompanyUUID).
			Find(&entitesInDB)

		if len(entitesInDB) != len(keys) {
			diff := lo.Interleave(entitesInDB, keys)

			return orm, fmt.Errorf("некоторые записи не найдены (%v)", strings.Join(diff, ", "))
		}

		for _, uid := range entitesInDB {
			ent = append(ent, domain.UUID{
				UUID: uuid.MustParse(uid),
			})
		}
	}

	orm = CatalogData{
		UUID:           dm.UUID,
		FederationUUID: dm.FederationUUID,
		CompanyUUID:    dm.CompanyUUID,
		CatalogUUID:    dm.CatalogUUID,
		Fields:         dm.Fields,
		Entities:       ent,

		CreatedBy:     dm.CreatedBy,
		CreatedByUUID: dm.CreatedByUUID,
	}

	err = r.gorm.DB.
		Create(&orm).
		Error

	return orm, err
}

func (r *Repository) GetData(filter dto.CatalogSearchDTO, allowSort []string) (dms []domain.CatalogData, total int64, err error) {
	defer r.storeTime("GetData", tm())

	orms := []CatalogData{}

	offset := 0
	limit := 25
	if filter.Limit != nil {
		limit = *filter.Limit
	}

	if filter.Offset != nil {
		offset = *filter.Offset
	}

	sqlWhere := ""
	if len(filter.Fields) > 0 {
		queryWhere := r.gorm.DB
		for _, item := range filter.Fields {
			// @todo: add regular to check array
			if strings.HasPrefix(fmt.Sprintf("%v", item.Value), "@> [") && strings.HasSuffix(fmt.Sprintf("%v", item.Value), "]") {
				v := strings.TrimPrefix(item.Value.(string), "@> ")
				queryWhere = queryWhere.Where(" fields->? @> ?", item.Name, v)
			} else {
				// @todo: add ilike search
				v := strings.ReplaceAll(fmt.Sprintf("%v", item.Value), "%", "")
				queryWhere = queryWhere.Where("fields->>? ilike ?", item.Name, v+"%")
			}
		}
		sql2 := queryWhere.ToSQL(func(tx *gorm.DB) *gorm.DB {
			return tx.Find(&orms)
		})
		logrus.Debug("queryWhere: ", strings.Split(sql2, "WHERE")[1])

		sqlWhere = "and " + strings.Split(sql2, "WHERE")[1]
	}

	query := r.gorm.DB

	orderStr := "order by created_at desc"
	if len(allowSort) > 0 && filter.Order != nil {
		by := "desc"
		if filter.By != nil && *filter.By == "asc" {
			by = "asc"
		}

		if helpers.InArray(*filter.Order, allowSort) {
			if strings.HasPrefix(*filter.Order, "fields.") {
				*filter.Order = "fields->>'" + strings.Replace(*filter.Order, "fields.", "", 1) + "'"
			}

			orderStr = "order by " + *filter.Order + " " + by
		}
	}

	// @todo: add federation limit
	query = query.Raw(` 
			with rich as (
			SELECT 
			o.uuid , 
			JSON_AGG(
				JSON_BUILD_OBJECT('uuid', u.uuid, 'c_uuid', u.catalog_uuid,  'fields', u.fields)
			) as entities_rich 
				
			FROM catalog_data o
			
			CROSS JOIN LATERAL JSONB_ARRAY_ELEMENTS(o.entities ) AS  usr
			INNER JOIN catalog_data u ON (usr->>'uuid')::text = u.uuid::text 
				
			GROUP BY o.uuid 
			) 
			select  o.*, rich.entities_rich, count(*) OVER() AS total
				from catalog_data o
				left join rich on o.uuid = rich.uuid  
				where  
				 
				o.catalog_uuid = ? 
				`+sqlWhere+" "+orderStr+" "+` 
					limit ? offset ?

					 
			`, filter.CatalogUUID, limit, offset)

	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&orms)
	})

	logrus.Debug("sql: ", sql)

	result := query.Scan(&orms)

	if result.Error != nil {
		return dms, -1, result.Error
	}

	if len(orms) > 0 {
		total = orms[0].Total
	}

	dms = helpers.Map(orms, func(item CatalogData, i int) domain.CatalogData {
		return domain.CatalogData{
			UUID: item.UUID,

			Fields: item.Fields,

			EntitiesRich: item.EntitiesRich,

			CreatedBy:     item.CreatedBy,
			CreatedByUUID: item.CreatedByUUID,

			FederationUUID: item.FederationUUID,
			CompanyUUID:    item.CompanyUUID,
			CatalogUUID:    item.CatalogUUID,

			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			DeletedAt: item.DeletedAt,
		}
	})

	return dms, total, nil
}

func (r *Repository) GetSortFields() []string {
	st := reflect.TypeOf(CatalogData{})

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
