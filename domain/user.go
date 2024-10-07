package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
)

type User struct {
	UUID             uuid.UUID
	Name             string `validate:"lte=30"  ru:"имя"`
	Lname            string `validate:"lte=30"  ru:"фамилия"`
	Pname            string `validate:"lte=30"  ru:"отчество"`
	Email            string `validate:"email"`
	Phone            int    `validate:"min=10000000000,max=9999999999999" ru:"телефон"`
	IsValid          bool
	ValidationSendAt *time.Time `json:"validation_send_at,omitempty"`
	Password         string     `validate:"gte=30,lte=100" ru:"пароль"`
	Provider         int        `validate:"min=0,max=9"`

	Color    string
	HasPhoto bool
	Photo    *ProfilePhotoDTO

	Preferences ProfilePreferences

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	ValidAt   *time.Time

	Meta string
}

type ProfilePreferences struct {
	Timezone *string `json:"timezone,omitempty"`
}

type ProfilePhotoDTO struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

const (
	ProviderEmail = iota
	ProviderTest
)

func NewUser(name, lname, pname, email string, phone int, password string) *User {
	user := &User{
		UUID:     uuid.New(),
		Name:     name,
		Lname:    lname,
		Pname:    pname,
		Email:    email,
		Phone:    phone,
		Password: helpers.Hash(password),
		Provider: int(ProviderEmail),
		HasPhoto: false,
		Color:    "#000000",

		Meta: "{}",
	}

	errs, ok := helpers.ValidationStruct(user)
	if !ok {
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return user
}

func NewUserByUUID(uid uuid.UUID) *User {
	user := &User{
		UUID: uid,
	}

	return user
}

func (u *User) ChangePassword(newPassword, oldPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("пароль должен быть не менее 6 символов")
	}

	err := helpers.VerifyHash(u.Password, oldPassword)
	if err != nil {
		return errors.New("старый пароль не совпадает")
	}

	u.Password = helpers.Hash(newPassword)

	return nil
}

func (u *User) ChangeColor(color string) error {
	err := helpers.ValidateColor(color)
	if err != nil {
		return err
	}

	u.Color = color
	return nil
}

func (u *User) ChangeFIO(name, lname, pname *string) error {
	if name != nil && *name == "" {
		return errors.New("имя не может быть пустым")
	}

	if name != nil {
		if len(*name) > 30 {
			return errors.New("имя не может быть больше 30 символов")
		}

		u.Name = *name
	}

	if lname != nil {
		if len(*lname) > 30 {
			return errors.New("фамилия не может быть больше 30 символов")
		}

		u.Lname = *lname
	}

	if pname != nil {
		if len(*pname) > 30 {
			return errors.New("отчество не может быть больше 30 символов")
		}

		u.Pname = *pname
	}

	return nil
}

func (u *User) ChangePhone(phone int) error {
	if phone < 10000000000 || phone > 9999999999999 {
		return errors.New("неверный формат телефона (79999999999)")
	}

	u.Phone = phone

	return nil
}

type SearchUser struct {
	FederationUUID uuid.UUID
	CompanyUUID    *uuid.UUID `json:"company_uuid,omitempty"`
	Search         string
}
