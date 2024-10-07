package profile

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/configs"
	"github.com/krisch/crm-backend/internal/dictionary"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/internal/jwt"
	"github.com/krisch/crm-backend/internal/s3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	userValidationCodeExpire = 60 * 60
)

type IDict interface {
	FindUserByUUID(uid uuid.UUID) (*dto.UserDTO, bool)
}

type Service struct {
	repo    *Repository
	storage *s3.Service
	dict    *dictionary.Service
	conf    *configs.Configs
}

func New(conf *configs.Configs, repo *Repository, storage *s3.Service, dict *dictionary.Service) *Service {
	return &Service{
		repo:    repo,
		storage: storage,
		dict:    dict,
		conf:    conf,
	}
}

func (s *Service) CreateUser(user *domain.User, createValidationCode bool) (uid uuid.UUID, code string, err error) {
	err = s.repo.CreateUser(*user)
	if err != nil {
		return uid, code, err
	}

	if createValidationCode {
		code, err = s.repo.StoreValidationSimple(user.Email)
		if err != nil {
			return uid, code, err
		}
	}

	return user.UUID, code, err
}

func (s *Service) GetUser(ctx context.Context, uid uuid.UUID, fields ...string) (domain.User, error) {
	defer Span(NewSpan(ctx, "GetUser"))()

	user, err := s.repo.WithCounter().GetUser(uid, fields...)
	if err != nil {
		return user, err
	}

	var photo *domain.ProfilePhotoDTO
	if user.HasPhoto {
		photo = &domain.ProfilePhotoDTO{
			Small:  s.GetSmallPhoto(user.UUID),
			Medium: s.GetMediumPhoto(user.UUID),
			Large:  s.GetLargePhoto(user.UUID),
		}

		user.Photo = photo
	}

	return user, err
}

func (s *Service) GetUserByEmail(ctx context.Context, email string, fields ...string) (domain.User, error) {
	defer Span(NewSpan(ctx, "GetUser"))()

	user, err := s.repo.WithCounter().GetUserByEmail(email, fields...)
	if err != nil {
		return user, err
	}

	var photo *domain.ProfilePhotoDTO
	if user.HasPhoto {
		photo = &domain.ProfilePhotoDTO{
			Small:  s.GetSmallPhoto(user.UUID),
			Medium: s.GetMediumPhoto(user.UUID),
			Large:  s.GetLargePhoto(user.UUID),
		}

		user.Photo = photo
	}

	return user, err
}

func (s *Service) DeleteUser(uid uuid.UUID) (err error) {
	err = s.storage.DeletePhoto(uid)
	if err != nil {
		logrus.
			WithField("user_uuid", uid).
			Error(err)
	}

	err = s.repo.DeleteUser(uid)

	return err
}

func (s *Service) ValidateUser(code string) error {
	return s.repo.ValidateUser(code)
}

func (s *Service) ValidateSimple(ctx context.Context, email, code string) error {
	u, err := s.GetUserByEmail(ctx, email, []string{"uuid", "email", "is_valid", "validation_send_at"}...)
	if err != nil {
		return err
	}

	if u.IsValid {
		return errors.New("email уже подтвержден")
	}

	if u.ValidationSendAt == nil {
		return errors.New("нет кода для подтверждения")
	}

	if u.ValidationSendAt != nil {
		if time.Since(*u.ValidationSendAt) > time.Minute*60 {
			return errors.New("код устарел")
		}
	}

	return s.repo.ValidateSimpleUser(email, code)
}

func (s *Service) SentValidate(email string) (string, error) {
	code, err := s.repo.StoreValidationCode(email)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) SentValidateSimple(ctx context.Context, email string) (string, error) {
	u, err := s.GetUserByEmail(ctx, email, []string{"uuid", "email", "is_valid", "validation_send_at"}...)
	if err != nil {
		return "", err
	}

	if u.IsValid {
		return "", errors.New("email уже подтвержден")
	}

	code, err := s.repo.StoreValidationSimple(email)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) SentReset(email string) (string, error) {
	code, err := s.repo.StoreResetCode(email)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *Service) ResetUserPassword(code, password string) error {
	hashedPassword := helpers.Hash(password)

	return s.repo.ResetUserPassword(code, hashedPassword)
}

func (s *Service) Login(email, password string, rememberMe bool, iJwt jwt.IJWT) (uid uuid.UUID, access, refresh string, exp time.Time, err error) {
	if len(password) < 6 {
		err = errors.New("пароль от 6 символов")
		return uid, access, refresh, exp, err
	}

	user, err := s.repo.GetUserByEmail(email, "uuid", "password", "email", "name", "lname", "pname", "is_valid")
	if err != nil {
		return uid, access, refresh, exp, err
	}

	if !s.isDev() && !user.IsValid {
		err = errors.New("нужно подтвердить email")
		return uid, access, refresh, exp, err
	}

	err = helpers.VerifyHash(user.Password, password)
	if err != nil {
		err = errors.New("пароль не правильный")
		return uid, access, refresh, exp, err
	}

	access = iJwt.GenerateJWT(user.UUID, user.Email, user.Name, user.IsValid, jwt.Minute*10)

	expAfter := jwt.Day
	if rememberMe {
		expAfter = jwt.Month
	}

	refresh, exp = iJwt.GenerateRefreshToken(user.UUID, user.Email, user.Name, user.IsValid, expAfter)

	return user.UUID, access, refresh, exp, err
}

func (s *Service) LoginAs(email string, rememberMe bool, iJwt jwt.IJWT) (uid uuid.UUID, access, refresh string, exp time.Time, err error) {
	user, err := s.repo.GetUserByEmail(email, "uuid", "password", "email", "name", "lname", "pname", "is_valid")
	if err != nil {
		return uid, access, refresh, exp, err
	}

	access = iJwt.GenerateJWT(user.UUID, user.Email, user.Name, user.IsValid, jwt.Minute*10)

	expAfter := jwt.Day
	if rememberMe {
		expAfter = jwt.Month
	}

	refresh, exp = iJwt.GenerateRefreshToken(user.UUID, user.Email, user.Name, user.IsValid, expAfter)

	return user.UUID, access, refresh, exp, err
}

func (s *Service) ChangePassword(uid uuid.UUID, newPassword, oldPassword string) (err error) {
	user, err := s.repo.GetUser(uid, "uuid", "password")
	if err != nil {
		return err
	}

	err = user.ChangePassword(newPassword, oldPassword)
	if err != nil {
		return err
	}

	err = s.repo.ChangeField(user.UUID, "password", user.Password)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) GetSmallPhoto(uid uuid.UUID) string {
	return s.storage.GetPhotoURL(uid, s3.SmallPhotoSize)
}

func (s *Service) GetMediumPhoto(uid uuid.UUID) string {
	return s.storage.GetPhotoURL(uid, s3.MediumPhotoSize)
}

func (s *Service) GetLargePhoto(uid uuid.UUID) string {
	return s.storage.GetPhotoURL(uid, s3.LargePhotoSize)
}

func (s *Service) UploadPhoto(ctx context.Context, path string, uid uuid.UUID) (err error) {
	defer Span(NewSpan(ctx, "UploadPhoto"))()

	err = s.storage.UploadPhoto(ctx, path, uid)

	if err != nil {
		logrus.
			WithField("user_uuid", uid).
			Error(err)
		return err
	}

	err = s.repo.ChangeField(uid, "has_photo", true)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) DeletePhoto(uid uuid.UUID) (err error) {
	err = s.repo.ChangeField(uid, "has_photo", false)
	if err != nil {
		return err
	}

	err = s.storage.DeletePhoto(uid)
	if err != nil {
		logrus.
			WithField("user_uuid", uid).
			Error(err)
		return err
	}

	return err
}

func (s *Service) ChangeColor(uid uuid.UUID, color string) error {
	user := domain.NewUserByUUID(uid)
	err := user.ChangeColor(color)
	if err != nil {
		return err
	}

	err = s.repo.ChangeField(uid, "color", color)

	return err
}

func (s *Service) ChangeFIO0(uid uuid.UUID, name, lname, pname *string) error {
	user := domain.NewUserByUUID(uid)

	err := user.ChangeFIO(name, lname, pname)
	if err != nil {
		return err
	}

	if name != nil {
		err = s.repo.ChangeField(uid, "name", name)
		if err != nil {
			return err
		}
	}

	if lname != nil {
		err = s.repo.ChangeField(uid, "lname", *lname)
		if err != nil {
			return err
		}
	}

	if pname != nil {
		err = s.repo.ChangeField(uid, "pname", *pname)
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) ChangeFIO(uid uuid.UUID, name, lname, pname *string) error {
	user := domain.NewUserByUUID(uid)

	err := user.ChangeFIO(name, lname, pname)
	if err != nil {
		return err
	}

	err = s.repo.gorm.DB.Transaction(func(tx *gorm.DB) error {
		if name != nil {
			_, err = s.repo.ChangeFieldTx(tx, uid, "name", name)
			if err != nil {
				return err
			}
		}

		if lname != nil {
			_, err = s.repo.ChangeFieldTx(tx, uid, "lname", *lname)
			if err != nil {
				return err
			}
		}

		if pname != nil {
			_, err = s.repo.ChangeFieldTx(tx, uid, "pname", *pname)
			if err != nil {
				return err
			}
		}

		s.repo.PubUpdate()

		return nil
	})

	return err
}

func (s *Service) ChangePhone(uid uuid.UUID, phone int) error {
	user := domain.NewUserByUUID(uid)

	err := user.ChangePhone(phone)
	if err != nil {
		return err
	}

	err = s.repo.ChangeField(uid, "phone", user.Phone)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) AcceptInvite(uid uuid.UUID) error {
	return s.repo.AcceptInvite(uid)
}

func (s *Service) DeclineInvite(uid uuid.UUID) error {
	return s.repo.DeclineInvite(uid)
}

func (s *Service) GetInvites(email string) ([]domain.Invite, error) {
	return s.repo.GetInvites(email)
}

func (s *Service) UpdatePreference(uid uuid.UUID, key, value string) error {
	allowedKeys := []string{
		"timezone",
		"language",
		"currency",
		"theme",
	}

	if !helpers.InArray(key, allowedKeys) {
		return errors.New("нельзя изменить этот параметр")
	}

	if value == "" {
		return errors.New("пустое значение")
	}

	if len(value) > 50 {
		return errors.New("слишком длинное значение")
	}

	return s.repo.UpdatePreference(uid, key, value)
}

func (s *Service) ChangePreferences(uid uuid.UUID, prefs domain.ProfilePreferences) (err error) {
	j, err := json.Marshal(prefs)
	if err != nil {
		return err
	}

	if prefs.Timezone != nil {
		_, err = time.LoadLocation(*prefs.Timezone)
		if err != nil {
			return err
		}
	}

	err = s.repo.gorm.DB.
		Exec("UPDATE users SET updated_at = NOW(), preferences = preferences || ? WHERE uuid = ?", j, uid).
		Error

	return err
}

func (s *Service) isDev() bool {
	return s.conf.ENV == "dev"
}
