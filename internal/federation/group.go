package federation

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/helpers"
)

func (s *Service) CreateGroup(group *domain.Group) (err error) {
	errs, ok := helpers.ValidationStruct(group)
	if !ok {
		err = errors.New(helpers.Join(errs, ", "))
		return err
	}

	err = s.repo.CreateGroup(group)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) GetGroup(uid uuid.UUID) (dm domain.Group, err error) {
	dm, err = s.repo.GetGroup(uid)
	if err != nil {
		return dm, err
	}

	users, err := s.repo.GetGroupUsers(uid)
	if err != nil {
		return dm, err
	}

	dm.Users = users

	return dm, err
}

func (s *Service) ChangeGroupName(uid uuid.UUID, name *string) (err error) {
	c := domain.NewGroupByUUID(uid)

	if name != nil {
		err = c.ChangeName(*name)
		if err != nil {
			return err
		}

		err = s.repo.ChangeGroupField(c.UUID, "name", c.Name)
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) DeleteGroup(uid uuid.UUID) (err error) {
	return s.repo.DeleteGroup(uid)
}

func (s *Service) GetCompanyGroups(ctx context.Context, companyUUID uuid.UUID) (items []domain.Group, err error) {
	defer Span(NewSpan(ctx, "GetCompanyGroups"))()

	items, err = s.repo.GetCompanyGroups(companyUUID)
	if err != nil {
		return items, err
	}

	return items, err
}

func (s *Service) GetGroupUsers(ctx context.Context, groupUUID uuid.UUID) (items []domain.User, err error) {
	defer Span(NewSpan(ctx, "GetGroupUsers"))()

	items, err = s.repo.GetGroupUsers(groupUUID)
	if err != nil {
		return items, err
	}

	return items, err
}

func (s *Service) GetUserGroups(ctx context.Context, userUUID uuid.UUID) (items []domain.Group, err error) {
	defer Span(NewSpan(ctx, "GetUserGroups"))()

	items, err = s.repo.GetUserGroups(userUUID)
	if err != nil {
		return items, err
	}

	return items, err
}

func (s *Service) GetUsersGroups(ctx context.Context, userUUIDs []uuid.UUID) (items map[uuid.UUID][]domain.Group, err error) {
	defer Span(NewSpan(ctx, "GetUsersGroups"))()

	items, err = s.repo.GetUsersGroups(userUUIDs)
	if err != nil {
		return items, err
	}

	return items, err
}

func (s *Service) AddUserToGroup(userUUID, groupUUID uuid.UUID, createdBy string, createdByUUID uuid.UUID) (err error) {
	err = s.repo.AddUserToGroup(userUUID, groupUUID, createdBy, createdByUUID)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) RemoveUserFromGroups(groupUUID, userUUID uuid.UUID) (err error) {
	err = s.repo.RemoveUserFromGroups(groupUUID, userUUID)
	if err != nil {
		return err
	}

	return err
}
