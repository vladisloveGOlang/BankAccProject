package federation

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/catalogs"
	"github.com/krisch/crm-backend/internal/dictionary"
)

type Service struct {
	repo     *Repository
	dict     *dictionary.Service
	catalogs *catalogs.Service
}

func NewUserService(repo *Repository, dict *dictionary.Service, cs *catalogs.Service) *Service {
	return &Service{
		repo:     repo,
		dict:     dict,
		catalogs: cs,
	}
}

func (s *Service) CreateFederation(federation *domain.Federation) (err error) {
	err = s.repo.CreateFederation(*federation)
	if err != nil {
		return err
	}

	for _, federationUser := range federation.Users {
		err = s.AddUser(federationUser)

		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) GetFederation(uid uuid.UUID) (dm domain.Federation, err error) {
	dm, err = s.repo.GetFederation(uid)
	if err != nil {
		return dm, err
	}

	federationUsers, err := s.repo.GetFederationUsers(uid)
	if err != nil {
		return dm, err
	}

	federationCompanies, err := s.repo.GetCompaniesByFederation(uid)
	if err != nil {
		return dm, err
	}

	// todo: to domain
	dm.Users = federationUsers
	dm.UsersTotal = len(federationUsers)
	dm.CompaniesTotal = len(federationCompanies)
	dm.Companies = federationCompanies

	return dm, err
}

func (s *Service) GetFederations() (dm []domain.Federation, err error) {
	dm, err = s.repo.GetFederations()
	if err != nil {
		return dm, err
	}

	return dm, err
}

func (s *Service) ChangeName(uid uuid.UUID, name string) (err error) {
	federation := domain.NewFederationUUID(uid)
	err = federation.ChangeName(name)
	if err != nil {
		return err
	}

	err = s.repo.ChangeField(federation.UUID, "name", federation.Name)

	return err
}

func (s *Service) DeleteFederation(uid uuid.UUID) (err error) {
	err = s.repo.DeleteFederation(uid)
	if err != nil {
		return err
	}

	_, err = s.repo.DeleteCompanies(uid)
	if err != nil {
		return err
	}

	err = s.repo.DeleteProjects(&uid, nil)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) DeleteProject(uid string) (err error) {
	return s.repo.DeleteProject(uid)
}

func (s *Service) AddUser(fu domain.FederationUser) error {
	err := s.repo.AddUser(fu)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) DeleteUser(federationUUID, userUUID uuid.UUID) error {
	return s.repo.DeleteUser(federationUUID, userUUID)
}

func (s *Service) GetFederationsByUser(ctx context.Context, userUUID uuid.UUID) (items []domain.Federation, err error) {
	defer Span(NewSpan(ctx, "GetFederationsByUser"))()

	items, err = s.repo.GetFederationsByUser(userUUID)
	if err != nil {
		return items, err
	}

	return items, err
}

func (s *Service) GetCompanyFederation(ctx context.Context, uid uuid.UUID) (dt dto.FederationDTO, err error) {
	defer Span(NewSpan(ctx, "GetCompanyFederation"))()

	dt, err = s.repo.GetCompanyFederation(uid)
	if err != nil {
		return dt, err
	}

	return dt, err
}

func (s *Service) SearchUser(search domain.SearchUser) (items []domain.User, err error) {
	items, err = s.repo.SearchUser(search)
	if err != nil {
		return items, err
	}

	return items, err
}

func (s *Service) SearchUserInDictionary(search domain.SearchUser) (items []domain.User, err error) {
	items, err = s.dict.SearchUsers(search)
	if err != nil {
		return items, err
	}

	return items, err
}

func (s *Service) InviteUser(invite *domain.Invite) (err error) {
	err = s.repo.CreateInvite(invite)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) DeleteInvite(uid uuid.UUID) (err error) {
	err = s.repo.DeleteInvite(uid)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) GetInvites(uid uuid.UUID) ([]domain.Invite, error) {
	return s.repo.GetInvites(uid)
}
