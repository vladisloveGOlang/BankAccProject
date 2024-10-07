package dto

import (
	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/helpers"
)

type IStorage interface {
	GetSmallPhoto(uuid uuid.UUID) string
	GetMediumPhoto(uuid uuid.UUID) string
	GetLargePhoto(uuid uuid.UUID) string
}

type UserDTO struct {
	UUID     uuid.UUID `json:"uuid"`
	Name     string    `json:"name" xlsx:"A" ru:"Имя"`
	Lname    string    `json:"lname" xlsx:"B" ru:"Фамилия"`
	Pname    string    `json:"pname" xlsx:"C" ru:"Отчество"`
	Email    string    `json:"email"  xlsx:"D" ru:"Email"`
	Phone    int       `json:"phone"  xlsx:"E" ru:"Телефон"`
	HasPhoto bool      `json:"has_photo"`

	Photo *ProfilePhotoDTO `json:"photo"`
}

type UserShortDTO struct {
	UUID  uuid.UUID `json:"uuid"`
	Name  string    `json:"name"`
	Lname string    `json:"lname"`
	Pname string    `json:"pname"`
	Email string    `json:"email"`

	Color string `json:"color"`

	PhotoSmallURL *string `json:"photo_small_url,omitempty"`
}

type UserLikeDTO struct {
	UnixAt int64        `json:"unix_at"`
	User   UserShortDTO `json:"user"`
}

type UserFederationDTO struct {
	UUID           uuid.UUID `json:"uuid"`
	FederationUUID uuid.UUID `json:"federation_uuid"`
}

type UserCompanyDTO struct {
	UUID           uuid.UUID `json:"uuid"`
	FederationUUID uuid.UUID `json:"federation_uuid"`
}

func NewUserDto(dm domain.User, s IStorage) UserDTO {
	var photo *ProfilePhotoDTO
	if s != nil {
		if dm.HasPhoto {
			photo = &ProfilePhotoDTO{
				Small:  s.GetSmallPhoto(dm.UUID),
				Medium: s.GetMediumPhoto(dm.UUID),
				Large:  s.GetLargePhoto(dm.UUID),
			}
		}
	}

	return UserDTO{
		UUID:     dm.UUID,
		Name:     dm.Name,
		Lname:    dm.Lname,
		Pname:    dm.Pname,
		Email:    dm.Email,
		Phone:    dm.Phone,
		HasPhoto: dm.HasPhoto,
		Photo:    photo,
	}
}

func NewUserShotDto(dm domain.User, s IStorage) UserShortDTO {
	var photoSmallURL *string
	if s != nil {
		if dm.HasPhoto {
			photoSmallURL = helpers.Ptr(s.GetSmallPhoto(dm.UUID))
		}
	}

	return UserShortDTO{
		UUID:  dm.UUID,
		Name:  dm.Name,
		Lname: dm.Lname,
		Pname: dm.Pname,
		Email: dm.Email,

		PhotoSmallURL: photoSmallURL,
	}
}
