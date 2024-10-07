package postgres

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	//nolint
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/krisch/crm-backend/internal/helpers"
)

type PDB struct {
	DB *sql.DB
}

type Creds string

func NewPDB(creds Creds) (*PDB, error) {
	connStr, err := helpers.ConvertPostgresCreds(string(creds))
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &PDB{DB: db}, nil
}

func (p *PDB) Migrate(migratePath string) error {
	driver, err := postgres.WithInstance(p.DB, &postgres.Config{
		MigrationsTable: "migrations",
	})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file:"+migratePath,
		"postgres", driver)
	if err != nil {
		return err
	}

	return m.Up()
}
