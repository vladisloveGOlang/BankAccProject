package gates

import (
	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
)

type IDictionary interface {
	GetUserFederatons(userUUID uuid.UUID) []uuid.UUID
	GetUserCompanies(userUUID uuid.UUID) []uuid.UUID
	FindFederation(federationUUID uuid.UUID) (*dto.FederationDTO, bool)
	FindCompany(uuid uuid.UUID) (*dto.CompanyDTO, bool)
}

type Service struct {
	dict IDictionary
	repo *Repository

	federationLimit int
	companiesLimit  int
	usersLimit      int
	projectLimits   int
	commentsLimit   int
}

type Permissions string

const (
	PermissionsFederationAdmin Permissions = "federation:admin"
	PermissionsCompanyAdmin    Permissions = "company:admin"
	PermissionsProjectAdmin    Permissions = "project:admin"
	PermissionsUserAdmin       Permissions = "user:admin"
)

func New(repo *Repository, dict IDictionary) *Service {
	return &Service{
		dict: dict,
		repo: repo,

		federationLimit: 3,
		companiesLimit:  9,
		projectLimits:   20,
		commentsLimit:   300,
		usersLimit:      1000,
	}
}

func (a *Service) CreateOrUpdatePermisson(perm *domain.Permission) error {
	return a.repo.CreateOrUpdatePermisson(perm)
}

func (a *Service) DeletePermission(userUUID uuid.UUID) error {
	return a.repo.DeletePermission(userUUID)
}

func (a *Service) GetPermisson(uid uuid.UUID) (domain.Permission, error) {
	return a.repo.GetPermisson(uid)
}
