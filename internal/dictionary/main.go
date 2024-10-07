package dictionary

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

type Service struct {
	repo *Repository
	lock sync.RWMutex

	usersByEmail               map[string]dto.UserDTO
	usersByUUID                map[uuid.UUID]dto.UserDTO
	federationByUUID           map[uuid.UUID]dto.FederationDTO
	companiesByUUID            map[uuid.UUID]dto.CompanyDTO
	projectsByUUID             map[uuid.UUID]dto.ProjectDTO
	companyFieldsByProjectUUID map[uuid.UUID][]dto.CompanyFieldDTO
	projectFields              map[uuid.UUID][]dto.ProjectFieldDTO
	catalogFieldsByCatalogUUID map[uuid.UUID][]dto.CatalogFieldDTO
	tagsByUUID                 map[uuid.UUID]dto.TagDTO

	usersFederations  map[uuid.UUID][]uuid.UUID
	usersCompanies    map[uuid.UUID][]uuid.UUID
	companyPriorities map[uuid.UUID][]dto.CompanyPriorityDTO

	federationUsers map[uuid.UUID][]dto.UserDTO

	shouldUpdate  bool
	lastUpdatedAt map[string]time.Time

	gauge *prometheus.GaugeVec

	storage IStorage
}

type IStorage interface {
	GetSmallPhoto(uuid uuid.UUID) string
	GetMediumPhoto(uuid uuid.UUID) string
	GetLargePhoto(uuid uuid.UUID) string
}

func (s *Service) MarkToUpdate(v bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.shouldUpdate = v
}

func (s *Service) ShoulUpdate() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.shouldUpdate
}

func New(repo *Repository, metrics *helpers.MetricsCounters, storage IStorage) *Service {
	s := &Service{
		repo:                       repo,
		usersByEmail:               make(map[string]dto.UserDTO),
		usersByUUID:                make(map[uuid.UUID]dto.UserDTO),
		federationByUUID:           make(map[uuid.UUID]dto.FederationDTO),
		companiesByUUID:            make(map[uuid.UUID]dto.CompanyDTO),
		projectsByUUID:             make(map[uuid.UUID]dto.ProjectDTO),
		companyFieldsByProjectUUID: make(map[uuid.UUID][]dto.CompanyFieldDTO),
		projectFields:              make(map[uuid.UUID][]dto.ProjectFieldDTO),
		catalogFieldsByCatalogUUID: make(map[uuid.UUID][]dto.CatalogFieldDTO),
		tagsByUUID:                 make(map[uuid.UUID]dto.TagDTO),
		companyPriorities:          make(map[uuid.UUID][]dto.CompanyPriorityDTO),

		usersFederations: make(map[uuid.UUID][]uuid.UUID),
		usersCompanies:   make(map[uuid.UUID][]uuid.UUID),

		federationUsers: make(map[uuid.UUID][]dto.UserDTO),

		shouldUpdate: false,

		lastUpdatedAt: make(map[string]time.Time),

		gauge: metrics.DicGauge,

		storage: storage,
	}

	go func() {
		for {
			totalCompanyFields := 0
			for t := range s.companyFieldsByProjectUUID {
				totalCompanyFields += len(s.companyFieldsByProjectUUID[t])
			}

			totalcatalogFields := 0
			for t := range s.catalogFieldsByCatalogUUID {
				totalcatalogFields += len(s.catalogFieldsByCatalogUUID[t])
			}

			if s.gauge != nil {
				s.gauge.WithLabelValues("users").Set(float64(len(s.usersByEmail)))
				s.gauge.WithLabelValues("tags").Set(float64(len(s.tagsByUUID)))
				s.gauge.WithLabelValues("users_by_uuid").Set(float64(len(s.usersByUUID)))
				s.gauge.WithLabelValues("federations").Set(float64(len(s.federationByUUID)))
				s.gauge.WithLabelValues("companies").Set(float64(len(s.companiesByUUID)))
				s.gauge.WithLabelValues("projects").Set(float64(len(s.projectsByUUID)))
				s.gauge.WithLabelValues("CompanyFields").Set(float64(len(s.companyFieldsByProjectUUID)))
				s.gauge.WithLabelValues("ProjectFields").Set(float64(len(s.projectFields)))
				s.gauge.WithLabelValues("CompanyFieldsTotal").Set(float64(totalCompanyFields))
				s.gauge.WithLabelValues("CompanyFieldsTotal").Set(float64(totalcatalogFields))
				s.gauge.WithLabelValues("usersCompaniesTotal").Set(float64(len(s.usersCompanies)))
				s.gauge.WithLabelValues("usersFederationsTotal").Set(float64(len(s.usersFederations)))
				s.gauge.WithLabelValues("companyPrioritiesTotal").Set(float64(len(s.companyPriorities)))
			}

			time.Sleep(10 * time.Second)
		}
	}()

	return s
}

func (s *Service) Info() {
	totalCompanyFields := 0
	for t := range s.companyFieldsByProjectUUID {
		totalCompanyFields += len(s.companyFieldsByProjectUUID[t])
	}

	totalcatalogFields := 0
	for t := range s.catalogFieldsByCatalogUUID {
		totalcatalogFields += len(s.catalogFieldsByCatalogUUID[t])
	}

	logrus.
		WithField("u", s.lastUpdatedAt["users"]).
		WithField("t", s.lastUpdatedAt["tags"]).
		WithField("f", s.lastUpdatedAt["federations"]).
		WithField("c", s.lastUpdatedAt["companies"]).
		WithField("p", s.lastUpdatedAt["projects"]).
		WithField("cf", s.lastUpdatedAt["company_fields"]).
		WithField("pf", s.lastUpdatedAt["project_fields"]).
		WithField("df", s.lastUpdatedAt["catalog_fields"]).
		WithField("uf", s.lastUpdatedAt["user_federation"]).
		WithField("uc", s.lastUpdatedAt["user_company"]).
		WithField("cp", s.lastUpdatedAt["company_priorities"]).
		Info("last updatedAt")

	logrus.
		WithField("users", len(s.usersByEmail)).
		WithField("tags", len(s.tagsByUUID)).
		WithField("users_by_uuid", len(s.usersByUUID)).
		WithField("federations", len(s.federationByUUID)).
		WithField("companies", len(s.companiesByUUID)).
		WithField("projects", len(s.projectsByUUID)).
		WithField("CompanyFields", len(s.companyFieldsByProjectUUID)).
		WithField("ProjectFields", len(s.projectFields)).
		WithField("CompanyFields", totalCompanyFields).
		WithField("catalogsFields", totalcatalogFields).
		WithField("usersFederations", len(s.usersFederations)).
		WithField("usersCompanies", len(s.usersCompanies)).
		WithField("companyPriorities", len(s.companyPriorities)).
		Info("dict items")
}

func (s *Service) SyncUsers() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	users, err := s.repo.FetchUsers(s.lastUpdatedAt["users"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)

	federationUsers := s.federationUsers

	for _, user := range users {
		if user.DeletedAt != nil {
			continue
		}

		dm := domain.User{
			UUID:     user.UUID,
			Name:     user.Name,
			Lname:    user.Lname,
			Pname:    user.Pname,
			Email:    user.Email,
			Phone:    user.Phone,
			HasPhoto: user.HasPhoto,
		}

		userDTO := dto.NewUserDto(dm, s.storage)
		s.usersByEmail[user.Email] = userDTO
		s.usersByUUID[user.UUID] = userDTO

		if user.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = user.UpdatedAt
		}

		federationUsers[user.FederationUUID] = append(federationUsers[user.FederationUUID], userDTO)
	}

	s.lastUpdatedAt["users"] = lastUpdatedAt

	for k := range s.federationUsers {
		uniqUsers := lo.Uniq(federationUsers[k])
		sort.SliceStable(uniqUsers, func(i, j int) bool {
			return uniqUsers[i].Lname > uniqUsers[j].Lname
		})

		federationUsers[k] = uniqUsers
	}

	s.federationUsers = federationUsers

	return nil
}

func (s *Service) SyncFederations() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchFederations(s.lastUpdatedAt["federations"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)
	for _, i := range items {
		if i.DeletedAt != nil {
			delete(s.federationByUUID, i.UUID)
		} else {
			s.federationByUUID[i.UUID] = dto.FederationDTO{
				UUID: i.UUID,
				Name: i.Name,
			}
		}

		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}
	s.lastUpdatedAt["federations"] = lastUpdatedAt

	return nil
}

func (s *Service) SyncCompanies() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchCompanies(s.lastUpdatedAt["companies"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)
	for _, i := range items {
		if i.DeletedAt != nil {
			delete(s.companiesByUUID, i.UUID)
		} else {
			s.companiesByUUID[i.UUID] = dto.CompanyDTO{
				UUID:           i.UUID,
				Name:           i.Name,
				FederationUUID: i.FederationUUID,
			}
		}

		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}
	s.lastUpdatedAt["companies"] = lastUpdatedAt

	return nil
}

func (s *Service) SyncProjects() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchProjects(s.lastUpdatedAt["projects"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)
	for _, i := range items {

		if i.DeletedAt != nil {
			delete(s.projectsByUUID, i.UUID)
		} else {
			statusGraph := domain.NewStatusGraph("0")
			graph := statusGraph.Graph
			if len(i.StatusGraph) > 5 {
				statusGraph, err = domain.NewStatusGraphFromJSON(i.StatusGraph)
				if err != nil {
					logrus.Error(err)
				} else {
					graph = statusGraph.Graph
				}
			}

			// todo: refactor
			var options dto.ProjectOptionsDTO
			err = json.Unmarshal([]byte(i.Options), &options)
			if err != nil {
				logrus.Error("Error on unmarshal json: ", err)
				continue
			}

			s.projectsByUUID[i.UUID] = dto.ProjectDTO{
				UUID:           i.UUID,
				Name:           i.Name,
				CompanyUUID:    i.CompanyUUID,
				FederationUUID: i.FederationUUID,
				StatusGraph:    &graph,
				Options:        &options,
			}
		}

		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}
	s.lastUpdatedAt["projects"] = lastUpdatedAt

	return nil
}

func (s *Service) SyncCompanyFields() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchCompanyFields(s.lastUpdatedAt["company_fields"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)
	mp := s.companyFieldsByProjectUUID

	for _, i := range items {
		if i.DeletedAt == nil {
			if _, ok := mp[i.CompanyUUID]; !ok {
				mp[i.CompanyUUID] = []dto.CompanyFieldDTO{}
			}

			mp[i.CompanyUUID] = append(mp[i.CompanyUUID], dto.CompanyFieldDTO{
				Hash:         i.Hash,
				Name:         i.Name,
				DataType:     i.DataType,
				ProjectsUUID: i.ProjectsUUID,
			})

			mp[i.CompanyUUID] = lo.UniqBy(mp[i.CompanyUUID], func(i dto.CompanyFieldDTO) interface{} {
				return i.Hash
			})
		}

		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}

	s.companyFieldsByProjectUUID = mp
	s.lastUpdatedAt["company_fields"] = lastUpdatedAt

	return nil
}

func (s *Service) SyncProjectFields() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchProjectFields(s.lastUpdatedAt["project_fields"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)
	mp := s.projectFields

	for _, i := range items {
		if i.DeletedAt == nil {
			if _, ok := mp[i.ProjectUUID]; !ok {
				mp[i.ProjectUUID] = []dto.ProjectFieldDTO{}
			}

			mp[i.ProjectUUID] = append(mp[i.ProjectUUID], dto.ProjectFieldDTO{
				Hash:               i.Hash,
				Name:               i.Name,
				DataType:           i.DataType,
				RequiredOnStatuses: i.RequiredOnStatuses,
				ProjectUUID:        i.ProjectUUID,
			})

			mp[i.ProjectUUID] = lo.UniqBy(mp[i.ProjectUUID], func(i dto.ProjectFieldDTO) interface{} {
				return i.Hash
			})
		}

		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}

	s.projectFields = mp
	s.lastUpdatedAt["project_fields"] = lastUpdatedAt

	return nil
}

func (s *Service) SyncCatalogFields() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchCatalogFields(s.lastUpdatedAt["catalog_fields"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)

	mp := s.catalogFieldsByCatalogUUID

	for _, i := range items {
		if i.DeletedAt == nil {
			if _, ok := mp[i.CatalogUUID]; !ok {
				mp[i.CatalogUUID] = []dto.CatalogFieldDTO{}
			}

			mp[i.CatalogUUID] = lo.Uniq(append(mp[i.CatalogUUID], dto.CatalogFieldDTO{
				Hash:     i.Hash,
				Name:     i.Name,
				DataType: i.DataType,
			}))
		}

		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}

	s.catalogFieldsByCatalogUUID = mp
	s.lastUpdatedAt["catalog_fields"] = lastUpdatedAt

	return nil
}

func (s *Service) SyncUserFederation() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchUserFederation(s.lastUpdatedAt["user_federation"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)

	for _, i := range items {
		if i.DeletedAt != nil {
			s.usersFederations[i.UserUUID] = lo.Filter(s.usersFederations[i.UserUUID], func(uuid uuid.UUID, idx int) bool {
				return uuid != i.FederationUUID
			})
		} else {
			if _, ok := s.usersFederations[i.UserUUID]; !ok {
				s.usersFederations[i.UserUUID] = []uuid.UUID{}
			}

			s.usersFederations[i.UserUUID] = lo.Uniq(append(s.usersFederations[i.UserUUID], i.FederationUUID))
		}
		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}

	s.lastUpdatedAt["user_federation"] = lastUpdatedAt

	return nil
}

func (s *Service) SyncUserCompany() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchUserCompany(s.lastUpdatedAt["user_company"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)

	for _, i := range items {
		if i.DeletedAt != nil {
			s.usersCompanies[i.UserUUID] = lo.Filter(s.usersCompanies[i.UserUUID], func(uuid uuid.UUID, idx int) bool {
				return uuid != i.CompanyUUID
			})
		} else {
			if _, ok := s.usersCompanies[i.UserUUID]; !ok {
				s.usersCompanies[i.UserUUID] = []uuid.UUID{}
			}

			s.usersCompanies[i.UserUUID] = lo.Uniq(append(s.usersCompanies[i.UserUUID], i.CompanyUUID))
		}
		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}

	s.lastUpdatedAt["user_company"] = lastUpdatedAt

	return nil
}

func (s *Service) SyncCompanyPriorities() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchCompanyPriority(s.lastUpdatedAt["company_priorities"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)

	for _, i := range items {
		if i.DeletedAt != nil {
			s.companyPriorities[i.CompanyUUID] = lo.Filter(s.companyPriorities[i.CompanyUUID], func(cp dto.CompanyPriorityDTO, _ int) bool {
				return cp.UUID != i.UUID
			})
		} else {
			if _, ok := s.companyPriorities[i.CompanyUUID]; !ok {
				s.companyPriorities[i.CompanyUUID] = []dto.CompanyPriorityDTO{}
			}

			s.companyPriorities[i.CompanyUUID] = lo.Uniq(append(s.companyPriorities[i.CompanyUUID], dto.CompanyPriorityDTO{
				UUID:   i.UUID,
				Name:   i.Name,
				Number: i.Number,
				Color:  i.Color,
			}))
		}
		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}

	s.lastUpdatedAt["company_priorities"] = lastUpdatedAt

	return nil
}

//

func (s *Service) GetRandomUser() (*dto.UserDTO, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// It is slow, but ok.

	keys := lo.Keys(s.usersByEmail)
	rand := helpers.RandomNumber(0, len(keys))

	user := s.usersByEmail[keys[rand]]

	return &user, false
}

func (s *Service) FindUser(email string) (*dto.UserDTO, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.usersByEmail) == 0 {
		logrus.Error("users not synced")
		return &dto.UserDTO{}, false
	}

	if user, ok := s.usersByEmail[email]; ok {
		return &user, true
	}

	return &dto.UserDTO{}, false
}

func (s *Service) FindUserByUUID(uid uuid.UUID) (*dto.UserDTO, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.usersByUUID) == 0 {
		logrus.Error("users by uuid not synced")
		return &dto.UserDTO{}, false
	}

	if user, ok := s.usersByUUID[uid]; ok {
		return &user, true
	}

	return &dto.UserDTO{}, false
}

func (s *Service) FindFederation(uid uuid.UUID) (*dto.FederationDTO, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.federationByUUID) == 0 {
		logrus.Error("federations not synced")
		return &dto.FederationDTO{}, false
	}

	if i, ok := s.federationByUUID[uid]; ok {
		return &i, true
	}

	return &dto.FederationDTO{}, false
}

func (s *Service) FindCompany(uid uuid.UUID) (*dto.CompanyDTO, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.companiesByUUID) == 0 {
		logrus.Error("companies not synced")
		return &dto.CompanyDTO{}, false
	}

	if i, ok := s.companiesByUUID[uid]; ok {
		return &i, true
	}

	return &dto.CompanyDTO{}, false
}

func (s *Service) FindTag(uid uuid.UUID) (*dto.TagDTO, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.tagsByUUID) == 0 {
		logrus.Error("tags not synced")
		return &dto.TagDTO{}, false
	}

	if i, ok := s.tagsByUUID[uid]; ok {
		return &i, true
	}

	return &dto.TagDTO{}, false
}

func (s *Service) FindProject(uid uuid.UUID) (*dto.ProjectDTO, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.projectsByUUID) == 0 {
		logrus.Error("projects not synced")
		return &dto.ProjectDTO{}, false
	}

	if i, ok := s.projectsByUUID[uid]; ok {
		return &i, true
	}

	return &dto.ProjectDTO{}, false
}

func (s *Service) GetRandomCompany() (*dto.CompanyDTO, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.companiesByUUID) == 0 {
		logrus.Error("companies not synced")
		return &dto.CompanyDTO{}, false
	}

	companiiesCount := len(s.companiesByUUID)
	randomIdx := helpers.RandomNumber(0, companiiesCount)
	k := 0
	for _, company := range s.companiesByUUID {
		if randomIdx == k {
			return &company, true
		}

		k++
	}

	return &dto.CompanyDTO{}, false
}

func (s *Service) FindCompanyFields(companyUUID uuid.UUID) (dtos []dto.CompanyFieldDTO, found bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.companyFieldsByProjectUUID) == 0 {
		logrus.Error("company fields fields not synced")
		return dtos, found
	}

	if i, ok := s.companyFieldsByProjectUUID[companyUUID]; ok {
		found = true
		return i, found
	}

	return dtos, found
}

func (s *Service) FindCompanyPriorities(companyUUID uuid.UUID, number int) (*dto.CompanyPriorityDTO, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.companyPriorities) == 0 {
		logrus.Error("сompany priorities not synced")
		return &dto.CompanyPriorityDTO{}, false
	}

	if _, ok := s.companyPriorities[companyUUID]; ok {
		for _, cp := range s.companyPriorities[companyUUID] {
			if cp.Number == number {
				return &cp, true
			}
		}
	}

	return &dto.CompanyPriorityDTO{}, false
}

func (s *Service) FindProjectFields(projectUUID uuid.UUID) (dtos []dto.ProjectFieldDTO, found bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.projectFields) == 0 {
		logrus.Error("project fields fields not synced")
		return dtos, found
	}

	if i, ok := s.projectFields[projectUUID]; ok {
		found = true
		return i, found
	}

	return dtos, found
}

func (s *Service) FindCatalogFields(catalogUUID uuid.UUID) (dtos []dto.CatalogFieldDTO, found bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.catalogFieldsByCatalogUUID) == 0 {
		logrus.Error("catalogs fields not synced")
		return dtos, found
	}

	if i, ok := s.catalogFieldsByCatalogUUID[catalogUUID]; ok {
		found = true
		return i, found
	}

	return dtos, found
}

func (s *Service) GetUserCompanies(userUUID uuid.UUID) (res []uuid.UUID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.usersCompanies) == 0 {
		logrus.Error("user companies not synced")
		return res
	}

	if v, ok := s.usersCompanies[userUUID]; ok {
		return v
	}

	return res
}

func (s *Service) GetUserFederatons(userUUID uuid.UUID) (res []uuid.UUID) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.usersFederations) == 0 {
		logrus.Error("user federations not synced")
		return res
	}

	if v, ok := s.usersFederations[userUUID]; ok {
		return v
	}

	return res
}

type MatchItem[T any] struct {
	Match int
	Item  T
}

func Contain(s string, fields []interface{}) int {
	strs := strings.Split(s, " ")
	match := 0

	for _, subStr := range strs {
		for _, f := range fields {
			if strings.ToLower(fmt.Sprint(f)) == subStr {
				match += 5
			}
		}
	}

	for _, subStr := range strs {
		for _, f := range fields {
			if strings.HasPrefix(strings.ToLower(fmt.Sprint(f)), subStr) {
				match += 3

				if len(fmt.Sprint(f))-len(subStr) <= 5 {
					match += len(fmt.Sprint(f)) - len(subStr)
				}
			}
		}
	}

	for _, subStr := range strs {
		for _, f := range fields {
			if strings.Contains(strings.ToLower(fmt.Sprint(f)), subStr) {
				match++
			}
		}
	}

	return match
}

func (s *Service) SearchUsers(search domain.SearchUser) (dms []domain.User, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	str := strings.ToLower(search.Search)

	found := 0
	items := []MatchItem[domain.User]{}

	federationUsers := s.federationUsers[search.FederationUUID]

	for _, user := range federationUsers {
		strs := strings.Split(str, " ")
		match := 0

		for _, subStr := range strs {
			match += Contain(subStr, []interface{}{user.Name, user.Lname, user.Pname, user.Email, user.Phone})
		}

		if match >= len(strs) {
			items = append(items, MatchItem[domain.User]{
				Match: match,
				Item: domain.User{
					UUID:     user.UUID,
					Name:     user.Name,
					Lname:    user.Lname,
					Pname:    user.Pname,
					Phone:    user.Phone,
					Email:    user.Email,
					HasPhoto: user.HasPhoto,
				},
			})

			found++
			if found >= 50 {
				break
			}
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Match > items[j].Match
	})

	dms = lo.Map(items, func(item MatchItem[domain.User], index int) domain.User {
		return item.Item
	})

	return dms, nil
}

func (s *Service) SyncCompanyTags() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	items, err := s.repo.FetchCompanyTags(s.lastUpdatedAt["tags"])
	if err != nil {
		return err
	}

	lastUpdatedAt := time.Now().AddDate(-10, 0, 0)
	for _, i := range items {

		if i.DeletedAt != nil {
			delete(s.tagsByUUID, i.UUID)
		} else {
			s.tagsByUUID[i.UUID] = dto.TagDTO{
				UUID:  i.UUID,
				Name:  i.Name,
				Color: i.Color,
			}
		}

		if i.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = i.UpdatedAt
		}
	}
	s.lastUpdatedAt["tags"] = lastUpdatedAt

	return nil
}

func (s *Service) SyncAll() {
	wg := sync.WaitGroup{}
	wg.Add(11)

	go func() {
		defer wg.Done()
		err := s.SyncUsers()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncUsers done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncCompanyTags()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncCompanyTags done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncFederations()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncFederations done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncCompanies()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncCompanies done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncProjects()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncProjects done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncCompanyFields()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncCompanyFields done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncProjectFields()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncProjectFields done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncCatalogFields()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncCatalogFields done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncUserFederation()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncUserFederation done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncUserCompany()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncUserCompany done")
	}()

	go func() {
		defer wg.Done()
		err := s.SyncCompanyPriorities()
		if err != nil {
			logrus.Error(err)
		}
		logrus.Debug("SyncCompanyPriorities done")
	}()

	wg.Wait()
}
