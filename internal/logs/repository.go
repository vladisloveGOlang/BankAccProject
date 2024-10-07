package logs

import (
	"github.com/krisch/crm-backend/pkg/postgres"
)

type ILogRepository interface {
	InsertLog(l Log) error
}

type LogRepository struct {
	gorm *postgres.GDB
}

func NewLogRepository(db *postgres.GDB) ILogRepository {
	return &LogRepository{
		gorm: db,
	}
}

// ----------------------------

func (r *LogRepository) InsertLog(l Log) error {
	return r.gorm.DB.Create(&l).Error
}
