package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/samber/lo"
)

type FederationDTO struct {
	UUID      uuid.UUID  `json:"uuid"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	CreatedByUUID *uuid.UUID `json:"created_by_uuid,omitempty"`

	UsersTotal     int `json:"users_total"`
	CompaniesTotal int `json:"companies_total"`

	Companies []CompanyDTO        `json:"companies"`
	Users     []FederationUserDto `json:"users"`
}

type FederationDTOs struct {
	UUID    uuid.UUID `json:"uuid"`
	Name    string    `json:"name"`
	IsLiked *bool     `json:"is_liked,omitempty"`
}

func NewFederationDTOs(dto *FederationDTO) FederationDTOs {
	if dto == nil {
		return FederationDTOs{}
	}

	return FederationDTOs{
		UUID: dto.UUID,
		Name: dto.Name,
	}
}

type FederationUserDto struct {
	UUID    uuid.UUID   `json:"uuid"`
	User    UserDTO     `json:"user"`
	AddedAt time.Time   `json:"added_at"`
	Groups  []GroupDTOs `json:"groups"`
}

func NewFederationUserDto(dm domain.FederationUser, usersGroups map[uuid.UUID][]domain.Group, s IStorage) *FederationUserDto {
	groups := []GroupDTOs{}
	_, ok := usersGroups[dm.UUID]
	if ok {
		groups = lo.Map(usersGroups[dm.UUID], func(item domain.Group, index int) GroupDTOs {
			return GroupDTOs{
				UUID: item.UUID,
				Name: item.Name,
			}
		})
	}

	return &FederationUserDto{
		UUID:    dm.UUID,
		User:    NewUserDto(dm.User, s),
		AddedAt: dm.AddedAt,
		Groups:  groups,
	}
}
