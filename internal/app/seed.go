package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

func (a *App) Seed(ch chan string, usersCount, projectsCount, cores, tasksCountPerCore, batchSize int) error {
	ctx := context.Background()

	customers := []domain.User{}
	for i := 0; i < usersCount; i++ {
		email := fmt.Sprintf("nightsong-%v@oviovi.site", i)

		user := &domain.User{
			UUID:     uuid.New(),
			Name:     helpers.FakeName(),
			Lname:    helpers.FakeLName(),
			Pname:    helpers.FakePName(),
			Email:    email,
			Password: "secret",
			Phone:    helpers.FakePhone(),
			IsValid:  true,
			Provider: 1,
			Meta:     "{}",
		}

		_, _, err := a.ProfileService.CreateUser(user, false)
		if err != nil {
			logrus.Warn(err)

			userFromDB, err := a.ProfileService.GetUserByEmail(ctx, email, "uuid", "password", "email", "name", "lname", "pname", "is_valid")
			if err != nil {
				return err
			}
			customers = append(customers, userFromDB)

		} else {
			customers = append(customers, *user)
		}

		ch <- fmt.Sprintf("user (%v): %s", i, user.Email)
	}

	fdrn := domain.NewFederation(helpers.FakeName(), customers[0].Email, customers[0].UUID)

	err := a.FederationService.CreateFederation(fdrn)
	if err != nil {
		return err
	}

	for i, user := range customers {
		if i == 0 {
			continue
		}

		fu := domain.NewFederationUser(fdrn.UUID, user.UUID)

		err := a.FederationService.AddUser(*fu)
		if err != nil {
			return err
		}
	}

	company := &domain.Company{
		UUID:           uuid.New(),
		Name:           helpers.FakeName(),
		FederationUUID: fdrn.UUID,
		CreatedBy:      customers[0].Email,
		CreatedByUUID:  customers[0].UUID,
	}

	err = a.FederationService.CreateCompany(company, true)
	if err != nil {
		return err
	}

	// Catalogs

	catalogDnms, err := a.CatalogService.GetCatalogsByCompany(company.UUID)
	if err != nil {
		return err
	}

	usersCatalog := lo.Filter(catalogDnms, func(item domain.Catalog, index int) bool {
		return item.Name == "users"
	})[0]

	companiesCatalog := lo.Filter(catalogDnms, func(item domain.Catalog, index int) bool {
		return item.Name == "companies"
	})[0]

	//

	i := 100
	for i > 0 {
		i--

		newCompanyUUID := uuid.New()

		mp := make(map[string]interface{})
		mp["name"] = "my company"

		catalogData2 := domain.CatalogData{
			UUID:           newCompanyUUID,
			FederationUUID: companiesCatalog.FederationUUID,
			CompanyUUID:    companiesCatalog.CompanyUUID,
			CatalogUUID:    companiesCatalog.UUID,
			Fields:         make(map[string]interface{}),
			RawFields:      mp,
			Entities:       make(map[string]interface{}),

			CreatedBy:     customers[0].Email,
			CreatedByUUID: customers[0].UUID,
		}

		_, err = a.CatalogService.AddData(catalogData2)
		if err != nil {
			logrus.Fatal(err)
		}

		//
		j := 100
		for j > 0 {
			j--
			mp := make(map[string]interface{})
			mp["name"] = helpers.FakeName()
			mp["lname"] = helpers.FakeLName()
			mp["phone"] = fmt.Sprint(helpers.FakePhone())
			mp["company_uuid"] = newCompanyUUID.String()

			userData := domain.CatalogData{
				UUID:           uuid.New(),
				FederationUUID: usersCatalog.FederationUUID,
				CompanyUUID:    usersCatalog.CompanyUUID,
				CatalogUUID:    usersCatalog.UUID,
				Fields:         make(map[string]interface{}),
				RawFields:      mp,
				Entities:       make(map[string]interface{}),

				CreatedBy:     customers[0].Email,
				CreatedByUUID: customers[0].UUID,
			}

			if _, err = a.CatalogService.AddData(userData); err != nil {
				logrus.Fatal(err)
			}
		}
	}

	//

	for i, user := range customers {
		if i == 0 {
			continue
		}

		cu := domain.NewCompanyUser(fdrn.UUID, company.UUID, user.UUID)
		err := a.FederationService.AddUserToCompany(*cu)
		if err != nil {
			return err
		}
	}

	projects := []domain.Project{}
	for i := 0; i < projectsCount; i++ {

		project := &domain.Project{
			UUID:           uuid.New(),
			Name:           helpers.FakeName(),
			CompanyUUID:    company.UUID,
			FederationUUID: fdrn.UUID,
			CreatedBy:      customers[0].Email,
		}

		err := a.FederationService.CreateProgect(project)
		if err != nil {
			return err
		}

		projects = append(projects, *project)

	}

	for i := 0; i < projectsCount; i++ {

		_, err = a.FederationService.CreateCompanyField(&domain.CompanyField{
			UUID:        uuid.New(),
			Name:        "name",
			CompanyUUID: company.UUID,
			DataType:    domain.Integer,
		})
		if err != nil {
			return err
		}

		_, err = a.FederationService.CreateCompanyField(&domain.CompanyField{
			UUID:        uuid.New(),
			Name:        "name",
			CompanyUUID: company.UUID,
			DataType:    domain.String,
		})
		if err != nil {
			return err
		}

		_, err = a.FederationService.CreateCompanyField(&domain.CompanyField{
			UUID:        uuid.New(),
			Name:        "name",
			CompanyUUID: company.UUID,
			DataType:    domain.Bool,
		})
		if err != nil {
			return err
		}
	}

	v := atomic.NewInt64(0)

	for k := 1; k <= cores; k++ {
		go func() {
			batch := []domain.Task{}

			for i := 1; i <= tasksCountPerCore; i++ {

				mp := make(map[string]interface{})
				for i := 1; i < helpers.RandomNumber(0, 20); i++ {
					mp["a"] = helpers.RandomNumber(0, 20)
					mp["b"] = helpers.FakeSentence(4)
					mp["c"] = i%3 == 0
				}

				epicNumbers := helpers.RandomNumber(0, 6)
				epicNumbers2 := helpers.RandomNumber(0, 3)

				teammates := []string{}
				for i := 0; i < helpers.RandomNumber(3, 10); i++ {

					_, tm := helpers.RandomFromSlice(customers)
					teammates = append(teammates, tm.Email)
				}

				_, pr := helpers.RandomFromSlice(projects)
				projectUUID := pr.UUID

				tUUID := uuid.New()

				_, rb := helpers.RandomFromSlice(teammates)
				_, wb := helpers.RandomPartFromSlice(teammates)
				_, cw := helpers.RandomPartFromSlice(teammates)
				t := &domain.Task{
					UUID:           tUUID,
					Name:           helpers.FakeName(),
					ProjectUUID:    projectUUID,
					IsEpic:         epicNumbers > 0,
					FederationUUID: fdrn.UUID,
					CreatedBy:      teammates[0],
					ImplementBy:    teammates[1],
					ResponsibleBy:  rb,
					WatchBy:        wb,
					CoWorkersBy:    cw,
					Fields:         mp,

					Path: []string{tUUID.String()},
				}

				batch = append(batch, *t)
				v.Inc()

				for j := 1; j < epicNumbers; j++ {
					t2UUID := uuid.New()

					_, rb := helpers.RandomFromSlice(teammates)
					_, wb := helpers.RandomPartFromSlice(teammates)
					_, cw := helpers.RandomPartFromSlice(teammates)

					t2 := &domain.Task{
						UUID:           t2UUID,
						Name:           helpers.FakeName(),
						IsEpic:         false,
						ProjectUUID:    projectUUID,
						FederationUUID: fdrn.UUID,
						CreatedBy:      teammates[0],
						ImplementBy:    teammates[1],
						ResponsibleBy:  rb,
						WatchBy:        wb,
						CoWorkersBy:    cw,
						Fields:         mp,

						Path: []string{tUUID.String(), t2UUID.String()},
					}

					batch = append(batch, *t2)
					v.Inc()

					for j2 := 1; j2 < epicNumbers2; j2++ {
						t3UUID := uuid.New()

						_, rb := helpers.RandomFromSlice(teammates)
						_, wb := helpers.RandomPartFromSlice(teammates)
						_, cw := helpers.RandomPartFromSlice(teammates)

						t3 := &domain.Task{
							UUID:           uuid.New(),
							Name:           helpers.FakeName(),
							IsEpic:         false,
							ProjectUUID:    projectUUID,
							FederationUUID: fdrn.UUID,
							CreatedBy:      teammates[0],
							ImplementBy:    teammates[1],
							ResponsibleBy:  rb,
							WatchBy:        wb,
							CoWorkersBy:    cw,
							Fields:         mp,

							Path: []string{tUUID.String(), t2UUID.String(), t3UUID.String()},
						}

						batch = append(batch, *t3)
						v.Inc()
					}

				}

				if len(batch) > batchSize {
					err := a.TaskService.CreateTaskBatch("", batch)
					if err != nil {
						logrus.Error(err)
					}

					batch = []domain.Task{}
				}
			}

			err := a.TaskService.CreateTaskBatch("", batch)
			if err != nil {
				logrus.Error(err)
			}
		}()
	}

	df := 0
	go func() {
		for {
			time.Sleep(time.Second * 1)
			df++

			logrus.Infof("%v task count: %v (%v) ", df, v.Load(), cores*tasksCountPerCore)

			vInt := v.Load()
			ch <- fmt.Sprintf("%d", vInt)

			if vInt >= int64(cores*tasksCountPerCore) {
				close(ch)
				return
			}

		}
	}()

	return nil
}

func (a *App) SeedTasks(ctx context.Context, total int, projectUUID uuid.UUID, createdBy string, randomImplemented bool, commentsMax int) ([]domain.Task, error) {
	user, _ := a.DictionaryService.FindUser(createdBy)
	project, _ := a.DictionaryService.FindProject(projectUUID)
	companyFields, _ := a.DictionaryService.FindCompanyFields(projectUUID)

	batch := []domain.Task{}

	implementBy := user
	if randomImplemented {
		implementBy, _ = a.DictionaryService.GetRandomUser()
	}

	for i := 1; i <= total; i++ {

		mp := make(map[string]interface{})
		for _, f := range companyFields {
			if domain.FieldDataType(f.DataType) == domain.String {
				mp[f.Hash] = helpers.FakeString(30)
			}

			if domain.FieldDataType(f.DataType) == domain.Integer {
				mp[f.Hash] = helpers.RandomBigNumber()
			}

			if domain.FieldDataType(f.DataType) == domain.Float {
				mp[f.Hash] = float64(helpers.RandomNumber(0, 100)) + 3.1415
			}

			if domain.FieldDataType(f.DataType) == domain.Switch {
				mp[f.Hash] = i%2 == 0
			}

			if domain.FieldDataType(f.DataType) == domain.Array {
				mp[f.Hash] = []string{helpers.FakeName()}
			}

			if domain.FieldDataType(f.DataType) == domain.Data {
				mp[f.Hash] = helpers.DateNow()
			}

			if domain.FieldDataType(f.DataType) == domain.Text {
				mp[f.Hash] = helpers.FakeSentence(500)
			}
		}

		t, err := domain.NewTask(
			helpers.FakeName(),
			project.FederationUUID,
			project.CompanyUUID,
			project.UUID,
			user.Email,
			mp,
			[]string{},
			helpers.FakeSentence(500),
			[]string{},
			[]string{},
			implementBy.Email,
			implementBy.Email,
			5,
			nil,
			"",
			"",
			make(map[uuid.UUID][]string),
		)
		if err != nil {
			return batch, err
		}

		batch = append(batch, t)
	}

	err := a.TaskService.CreateTaskBatch("", batch)
	if err != nil {
		logrus.Error(err)
	}

	for _, task := range batch {
		for i := 1; i <= helpers.RandomNumber(0, commentsMax); i++ {

			rand := helpers.RandomNumber(0, 5)
			commentMentioned := []string{}
			if rand == 0 {
				u, _ := a.DictionaryService.GetRandomUser()
				commentMentioned = append(commentMentioned, u.Email)
			}

			comment := domain.NewComment(
				user.Email,
				task.UUID,
				uuid.Nil,
				commentMentioned,
				strings.Trim(helpers.FakeSentence(500), " "),
			)

			err := a.CommentService.CreateComment(ctx, *comment)
			if err != nil {
				logrus.Error(err)
			}
		}

		rand := helpers.RandomNumber(0, 20)
		if rand == 0 {
			imageName := fmt.Sprintf("photo-%v.jpg", helpers.RandomNumber(1, 5))

			_, err := a.S3PrivateService.UploadTaskFile(task.FederationUUID, task.UUID, imageName, "./"+imageName, user.UUID)
			if err != nil {
				return nil, err
			}

			// @todo: mv to service
			if len(task.People) > 0 {
				err = a.TaskService.TaskWasUpdatedOrCreated(task.UUID, task.People)
				if err != nil {
					logrus.Error("TaskWasUpdatedOrCreated error: ", err)
				}
			}

		}

	}

	return batch, err
}
