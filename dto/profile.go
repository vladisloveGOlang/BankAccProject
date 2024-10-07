package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/samber/lo"
)

type ProfileDTO struct {
	UUID    uuid.UUID `json:"uuid"`
	Name    string    `json:"name"`
	IsValid bool      `json:"is_valid"`
	Lname   string    `json:"lname"`
	Pname   string    `json:"pname"`

	Email string `json:"email"`
	Phone int    `json:"phone"`

	Photo *ProfilePhotoDTO `json:"photo"`
	Color string           `json:"color"`

	Federations []FederationDTOs  `json:"federations"`
	Companies   []UserCompanyDTOs `json:"companies"`
	Projects    []ProjectDTOs     `json:"projects"`
	Surveys     []SurveyDTOs      `json:"surveys"`

	NotificationsTotal int64 `json:"notifications_total"`

	Groups []GroupDTOs `json:"groups"`

	Preferences domain.ProfilePreferences `json:"preferences"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewProfileDto(dm domain.User, federations []FederationDTOs, companies []UserCompanyDTOs, projects []ProjectDTOs, notificationsTotal int64, groups []domain.Group, surveys []SurveyDTOs, prefs domain.ProfilePreferences) *ProfileDTO {
	var photo *ProfilePhotoDTO
	if dm.Photo != nil {
		photo = &ProfilePhotoDTO{
			Small:  dm.Photo.Small,
			Medium: dm.Photo.Medium,
			Large:  dm.Photo.Large,
		}
	}

	return &ProfileDTO{
		UUID:               dm.UUID,
		NotificationsTotal: notificationsTotal,
		Email:              dm.Email,
		IsValid:            dm.IsValid,
		Name:               dm.Name,
		Pname:              dm.Pname,
		Lname:              dm.Lname,
		Phone:              dm.Phone,
		Federations:        federations,
		Companies:          companies,
		Projects:           projects,
		Surveys:            surveys,
		Photo:              photo,
		Color:              dm.Color,

		Groups: lo.Map(groups, func(g domain.Group, i int) GroupDTOs {
			return GroupDTOs{
				UUID: g.UUID,
				Name: g.Name,
			}
		}),

		Preferences: prefs,

		CreatedAt: dm.CreatedAt,
		UpdatedAt: dm.UpdatedAt,
	}
}
