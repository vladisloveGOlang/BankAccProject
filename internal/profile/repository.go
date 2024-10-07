package profile

import (
	"context"
	"errors"
	"fmt"
	"strconv"

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
	gorm *postgres.GDB
	rds  *redis.RDS

	counter *prometheus.CounterVec
}

func (r *Repository) WithCounter() *Repository {
	if r.counter != nil {
		c := *r.counter
		c.With(prometheus.Labels{"repo": "total"}).Inc()
	}

	return r
}

func NewRepository(db *postgres.GDB, rds *redis.RDS, metrics *helpers.MetricsCounters) *Repository {
	return &Repository{
		gorm: db,
		rds:  rds,
		// events: events,

		counter: metrics.RepoCounter,
	}
}

func (r *Repository) PubUpdate() {
	err := r.rds.Publish(context.Background(), "update", "profile")
	logrus.Debug("pub update federation")
	if err != nil {
		logrus.Error(err)
	}
}

type UserEvent struct {
	UUID uuid.UUID
}

func (r *Repository) CreateUser(user domain.User) (err error) {
	c := *r.counter
	c.With(prometheus.Labels{"repo": "CreateUser"}).Inc()

	return r.gorm.DB.Transaction(func(tx *gorm.DB) error {
		orm := User{
			UUID:     user.UUID,
			Name:     user.Name,
			Lname:    user.Lname,
			Pname:    user.Pname,
			Email:    user.Email,
			Phone:    user.Phone,
			IsValid:  user.IsValid,
			Provider: user.Provider,
			Password: user.Password,
			HasPhoto: user.HasPhoto,
			Color:    user.Color,
		}

		err = tx.Create(&orm).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("пользователь с таким email уже существует")
		}

		ormPrefs := Preference{
			UserUUID: user.UUID,
		}

		err = tx.Create(&ormPrefs).Error
		if err != nil {
			return err
		}

		if err == nil {
			r.PubUpdate()
		}

		return err
	})
}

func (r *Repository) GetUser(uid uuid.UUID, fields ...string) (user domain.User, err error) {
	if len(fields) == 0 {
		fields = []string{"uuid"}
	}

	orm := User{}

	err = r.gorm.DB.Model(orm).
		Where("uuid = ?", uid).
		Select(fields).
		Take(&orm).
		Error

	if err != nil {
		return user, err
	}

	user = domain.User{
		UUID:  orm.UUID,
		Name:  orm.Name,
		Lname: orm.Lname,
		Pname: orm.Pname,
		Email: orm.Email,
		Phone: orm.Phone,

		Password: orm.Password,

		IsValid: orm.IsValid,
		ValidAt: orm.ValidAt,

		HasPhoto: orm.HasPhoto,
		Color:    orm.Color,

		Preferences: domain.ProfilePreferences{
			Timezone: orm.Preferences.Timezone,
		},

		CreatedAt: orm.CreatedAt,
		UpdatedAt: orm.UpdatedAt,
	}

	return user, err
}

func (r *Repository) GetUserByEmail(email string, fields ...string) (user domain.User, err error) {
	if len(fields) == 0 {
		fields = []string{"uuid"}
	}

	err = r.gorm.DB.Model(User{}).
		Where("email = ?", email).
		Select(fields).
		Take(&user).
		Error

	return user, err
}

func (r *Repository) ResetPassword(code string, user domain.User) error {
	email, err := r.rds.GetStr(context.TODO(), "resetpsswrd:"+code)
	if err != nil {
		return err
	}

	if email != user.Email {
		return errors.New("код истек или не существует")
	}

	return r.gorm.DB.Model(&User{}).
		Where("email = ?", user.Email).
		Update("password", helpers.Hash(user.Password)).
		Update("updated_at", "now()").
		Error
}

func (r *Repository) StorePasswordResetCode(user domain.User) (string, error) {
	code := helpers.GenerateResetCode()

	return code, r.rds.SetStr(context.TODO(), "resetpsswrd:"+code, user.Email, userValidationCodeExpire)
}

func (r *Repository) StoreValidationCode(email string) (string, error) {
	code := helpers.GenerateValidationCode()

	err := r.rds.SetStr(context.Background(), "validation:"+code, email, userValidationCodeExpire)
	if err != nil {
		return "", err
	}

	err = r.gorm.DB.Model(&User{}).
		Where("email = ?", email).
		Update("validation_send_at", "now()").
		Update("updated_at", "now()").
		Error

	return code, err
}

func (r *Repository) StoreResetCode(email string) (string, error) {
	code := helpers.GenerateResetCode()

	err := r.rds.SetStr(context.Background(), "reset:"+code, email, userValidationCodeExpire)
	if err != nil {
		return "", err
	}

	err = r.gorm.DB.Model(&User{}).
		Where("email = ?", email).
		Update("reset_send_at", "now()").
		Update("updated_at", "now()").
		Error

	return code, err
}

func (r *Repository) ValidateUser(code string) error {
	email, err := r.rds.GetStr(context.TODO(), "validation:"+code)
	if err != nil {
		return err
	}

	if email == "" {
		return errors.New("код истек или не существует")
	}

	err = r.gorm.DB.Model(&User{}).
		Where("email = ?", email).
		Update("is_valid", true).
		Update("valid_at", "now()").
		Update("updated_at", "now()").
		Error
	if err != nil {
		return err
	}

	err = r.rds.Del(context.TODO(), "validation:"+code)

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) StoreValidationSimple(email string) (string, error) {
	code := helpers.GenerateValidationSimpleCode()

	key := fmt.Sprintf("validation:simple:%s:%s", email, code)

	// @todo: mv to one transaction
	err := r.rds.HSET(context.Background(), key, "code", code)
	if err != nil {
		return "", err
	}
	err = r.rds.HSET(context.Background(), key, "validation_send_at", helpers.DateNowMilli())
	if err != nil {
		return "", err
	}
	err = r.rds.HSET(context.Background(), key, "email", email)
	if err != nil {
		return "", err
	}
	err = r.rds.HSET(context.Background(), key, "tries", 3)
	if err != nil {
		return "", err
	}

	err = r.rds.Expire(context.Background(), key, userValidationCodeExpire)
	if err != nil {
		return "", err
	}

	err = r.gorm.DB.Model(&User{}).
		Where("email = ?", email).
		Update("validation_send_at", "now()").
		Update("updated_at", "now()").
		Error

	return code, err
}

func (r *Repository) ValidateSimpleUser(email, code string) error {
	key := fmt.Sprintf("validation:simple:%s:%s", email, code)

	data, err := r.rds.HGetAll(context.TODO(), key)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return errors.New("код истек или не существует")
	}

	email, ok := data["email"]
	if !ok {
		return errors.New("нет emailа в data")
	}

	tries, ok := data["tries"]
	if !ok {
		return errors.New("количество не найдено")
	}

	triesInt, err := strconv.Atoi(tries)
	if err != nil {
		return err
	}

	if triesInt <= 0 {
		return errors.New("количество попыток исчерпано")
	}

	if code != data["code"] {
		triesInt--
		err = r.rds.HSET(context.TODO(), "validation:simple:"+code, "tries", triesInt)
		if err != nil {
			return err
		}

		return errors.New("код не совпадает")
	}

	err = r.gorm.DB.Model(&User{}).
		Where("email = ?", email).
		Update("is_valid", true).
		Update("valid_at", "now()").
		Update("updated_at", "now()").
		Error
	if err != nil {
		return err
	}

	err = r.rds.Del(context.TODO(), key)
	if err != nil {
		logrus.Error(err, "email", email, "key", key)
	}

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) ResetUserPassword(code, password string) error {
	email, err := r.rds.GetStr(context.TODO(), "reset:"+code)
	if err != nil {
		return err
	}

	if email == "" {
		return errors.New("код истек или не существует")
	}

	err = r.gorm.DB.Model(&User{}).
		Where("email = ?", email).
		Update("password", password).
		Update("updated_at", "now()").
		Error

	if err != nil {
		return err
	}

	err = r.rds.Del(context.TODO(), "reset:"+code)
	if err != nil {
		return err
	}

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) ChangeField(uid uuid.UUID, fieldName string, value interface{}) error {
	err := r.gorm.DB.
		Model(&User{}).
		Where("uuid = ?", uid).
		UpdateColumn(fieldName, value).
		UpdateColumn("updated_at", "now()").
		Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) ChangeFieldTx(tx *gorm.DB, uid uuid.UUID, fieldName string, value interface{}) (*gorm.DB, error) {
	hasExternalTransaction := true
	if tx == nil {
		tx = r.gorm.DB.Begin()
		hasExternalTransaction = false
	}

	err := tx.Model(&User{}).
		Where("uuid = ?", uid).
		UpdateColumn(fieldName, value).
		UpdateColumn("updated_at", "now()").
		Error

	if !hasExternalTransaction {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()

			if err == nil {
				r.PubUpdate()
			}
		}
	}

	return tx, err
}

func (r *Repository) DeleteUser(uid uuid.UUID) error {
	// todo: mv user to deleted table
	err := r.gorm.DB.
		Where("uuid = ?", uid).
		Delete(&User{}).
		Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) AcceptInvite(uid uuid.UUID) (err error) {
	orm := Invite{}

	res := r.gorm.DB.
		Where("uuid = ?", uid).
		Where("accepted_at is null").
		Where("declined_at is null").
		Where("deleted_at is null").
		Find(&orm)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("приглашение не найдено")
	}

	err = r.gorm.DB.
		Model(&Invite{}).
		Where("uuid = ?", uid).
		Update("accepted_at", "now()").
		Error
	if err != nil {
		return err
	}

	user, err := r.GetUserByEmail(orm.Email)
	if err != nil {
		return err
	}

	if user.UUID == uuid.Nil {
		return dto.NotFoundErrf("[email:%v] пользователь не найден", orm.Email)
	}

	// Add user to company
	if orm.CompanyUUID != nil {
		// Add user to federation
		err = r.AddUser(user.UUID, orm.FederationUUID)
		if err != nil {
			return err
		}

		cu := domain.CompanyUser{
			UUID:           uuid.New(),
			User:           user,
			FederationUUID: orm.FederationUUID,
			CompanyUUID:    *orm.CompanyUUID,
		}

		err = r.AddUserToCompany(cu)
		if err != nil {
			return err
		}
	} else {
		// Add user to federation
		err = r.AddUser(user.UUID, orm.FederationUUID)
		if err != nil {
			return err
		}
	}

	return err
}

func (r *Repository) DeclineInvite(uid uuid.UUID) (err error) {
	orm := Invite{}

	res := r.gorm.DB.
		Where("uuid = ?", uid).
		Where("declined_at is null").
		Where("accepted_at is null").
		Where("deleted_at is null").
		Find(&orm)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("приглашение не найдено")
	}

	err = r.gorm.DB.
		Model(&Invite{}).
		Where("uuid = ?", uid).
		Update("declined_at", "now()").
		Error

	if err != nil {
		return err
	}

	return err
}

func (r *Repository) AddUser(userUUID, federationUUID uuid.UUID) (err error) {
	existingRecord := &FederationUser{}

	res := r.gorm.DB.
		Select("uuid").
		Where("federation_uuid = ?", federationUUID).
		Where("user_uuid = ?", userUUID).
		Find(&existingRecord)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		return errors.New("пользователь уже добавлен")
	}

	existingRecord = &FederationUser{
		FederationUUID: federationUUID,
		UserUUID:       userUUID,
		UUID:           uuid.New(),
	}

	err = r.gorm.DB.Create(&existingRecord).Error

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

func (r *Repository) GetInvites(email string) (items []domain.Invite, err error) {
	orm := []Invite{}

	err = r.gorm.DB.Model(&orm).
		Where("email = ?", email).
		Where("deleted_at is null").
		Where("declined_at is null").
		Where("accepted_at is null").
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
			CreatedAt:      item.CreatedAt,
		}
	})

	return items, err
}

func (r *Repository) UpdatePreference(uid uuid.UUID, key, value string) error {
	err := r.gorm.DB.
		Model(&Preference{}).
		Where("user_uuid = ?", uid).
		Update("preferences", gorm.Expr("preferences || ?", gorm.Expr("jsonb_build_object(?, ?)", key, value))).
		Update("updated_at", "now()").
		Error

	if err == nil {
		r.PubUpdate()
	}

	return err
}

func (r *Repository) CreateSurvey(survey domain.Survey) (err error) {
	orm := Survey{
		UUID:      survey.UUID,
		UserUUID:  survey.User.UUID,
		UserEmail: survey.User.Email,
		Name:      survey.Name,
		Body:      survey.Body,
	}

	err = r.gorm.DB.Create(&orm).Error

	return err
}

func (r *Repository) GetSurvey(uid uuid.UUID) (orm Survey, err error) {
	err = r.gorm.DB.Model(Survey{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		First(&orm).
		Error

	return orm, err
}

func (r *Repository) GetSurveyByUserUUID(uid uuid.UUID) (orms []Survey, err error) {
	err = r.gorm.DB.Model(Survey{}).
		Where("user_uuid = ?", uid).
		Find(&orms).
		Error

	return orms, err
}

func (r *Repository) DeleteSurvey(uid uuid.UUID) (err error) {
	res := r.gorm.DB.
		Model(&Survey{}).
		Where("uuid = ?", uid).
		Where("deleted_at is null").
		Update("updated_at", "now()").
		Update("deleted_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("анкета уже удалена или не найдена")
	}

	return res.Error
}
