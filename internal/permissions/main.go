package permissions

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/samber/lo"
)

type Service struct {
	repo *Repository
}

type Permissions string

const (
	PermissionsFederationAdmin Permissions = "federation:admin"
	PermissionsCompanyAdmin    Permissions = "company:admin"
	PermissionsProjectAdmin    Permissions = "project:admin"
	PermissionsUserAdmin       Permissions = "user:admin"
	BilingAdmin                Permissions = "billing:admin"
)

func (a *Service) AllowedPermissions() []string {
	return []string{
		string(PermissionsFederationAdmin),
		string(PermissionsCompanyAdmin),
		string(PermissionsProjectAdmin),
		string(PermissionsUserAdmin),
		string(BilingAdmin),
	}
}

func New(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (a *Service) UCreateGroup(crtr domain.Creator, name string, permissions []string) (uuid.UUID, error) {
	allowedPermissions := a.AllowedPermissions()
	for _, permission := range permissions {
		if lo.IndexOf(allowedPermissions, permission) == -1 {
			return uuid.Nil, fmt.Errorf("invalid permission: %s", permission)
		}
	}

	return a.repo.UCreateGroup(crtr, name, permissions)
}

func (a *Service) DeleteGroup(uid uuid.UUID) error {
	return a.repo.DeleteGroup(uid)
}

func (a *Service) AddUserToGroup(userUUID, groupUUID uuid.UUID) error {
	return a.repo.AddUserToGroup(userUUID, groupUUID)
}

// func (a *Service) RemoveUserFromGroup(name string, userUUID uuid.UUID) error {
// 	logrus.Info("RemoveUserFromGroup", name, userUUID)
// 	return nil
// }

// func (a *Service) setUserPermissions(name string, permissions []string) error {
// 	logrus.Info("setUserPermissions", name, permissions)
// 	return nil
// }

// func (a *Service) GetUserState(fuid uuid.UUID) error {
// 	logrus.Info("GetUserState", fuid)
// 	return nil
// }
