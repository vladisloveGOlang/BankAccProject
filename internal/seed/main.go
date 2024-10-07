package seed

import (
	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/app"
	"github.com/krisch/crm-backend/internal/configs"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/sirupsen/logrus"
)

func Seed(opt *configs.Configs) {
	a, err := app.InitApp(helpers.FakeName(), opt.DB_CREDS, true, opt.REDIS_CREDS)
	if err != nil {
		logrus.Fatal(err)
	}

	companiesCatalog := domain.NewCatalog(
		"companies",
		uuid.MustParse("a837020f-ddc3-414c-afbf-9a88a4dc35c2"),
		uuid.MustParse("3700a983-d2d6-4925-97e2-29c1ff101d34"),
		"nightsong@oviovi.site",
		uuid.MustParse("e1cdc61c-2be0-42a8-9b51-135205ecdbb9"),
	)

	usersCatalog := domain.NewCatalog(
		"users",
		uuid.MustParse("a837020f-ddc3-414c-afbf-9a88a4dc35c2"),
		uuid.MustParse("3700a983-d2d6-4925-97e2-29c1ff101d34"),
		"nightsong@oviovi.site",
		uuid.MustParse("e1cdc61c-2be0-42a8-9b51-135205ecdbb9"),
	)

	if err = a.CatalogService.CreateCatalog(companiesCatalog); err != nil {
		logrus.Fatal(err)
	}

	if err = a.CatalogService.CreateCatalog(usersCatalog); err != nil {
		logrus.Fatal(err)
	}

	if _, err = a.CatalogService.CreateCatalogField(domain.NewCatalogFiled(
		"name",
		"name",
		domain.String,
		nil,
		usersCatalog.UUID,
		"nightsong@oviovi.site",
	)); err != nil {
		logrus.Fatal(err)
	}

	if _, err = a.CatalogService.CreateCatalogField(domain.NewCatalogFiled(
		"owner",
		"owner",
		domain.Data,
		helpers.Ptr(usersCatalog.UUID),
		companiesCatalog.UUID,
		"nightsong@oviovi.site",
	)); err != nil {
		logrus.Fatal(err)
	}

	// Data

	a.SyncDictionaries()

	mp := make(map[string]interface{})
	mp["name"] = "nighty night"

	userData := domain.CatalogData{
		UUID:           uuid.New(),
		FederationUUID: usersCatalog.FederationUUID,
		CompanyUUID:    usersCatalog.CompanyUUID,
		CatalogUUID:    usersCatalog.UUID,
		Fields:         make(map[string]interface{}),
		RawFields:      mp,
		Entities:       make(map[string]interface{}),

		CreatedBy:     "nightsong@oviovi.site",
		CreatedByUUID: uuid.MustParse("e1cdc61c-2be0-42a8-9b51-135205ecdbb9"),
	}

	if _, err = a.CatalogService.AddData(userData); err != nil {
		logrus.Fatal(err)
	}

	//

	mp = make(map[string]interface{})
	mp["owner"] = userData.UUID.String()

	catalogData2 := domain.CatalogData{
		UUID:           uuid.New(),
		FederationUUID: companiesCatalog.FederationUUID,
		CompanyUUID:    companiesCatalog.CompanyUUID,
		CatalogUUID:    companiesCatalog.UUID,
		Fields:         make(map[string]interface{}),
		RawFields:      mp,
		Entities:       make(map[string]interface{}),

		CreatedBy:     "nightsong@oviovi.site",
		CreatedByUUID: uuid.MustParse("e1cdc61c-2be0-42a8-9b51-135205ecdbb9"),
	}

	_, err = a.CatalogService.AddData(catalogData2)
	if err != nil {
		logrus.Fatal(err)
	}

	print(".")

	items, t, err := a.CatalogService.GetData(
		dto.CatalogSearchDTO{
			CatalogUUID: companiesCatalog.UUID,
			Offset:      helpers.Ptr(0),
			Limit:       helpers.Ptr(10),
		},
	)
	if err != nil {
		logrus.Fatal(err)
	}

	println(t)
	println(len(items))
	println(items[0].UUID.String())
}
