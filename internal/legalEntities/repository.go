package legalEntities

import (
	"github.com/krisch/crm-backend/pkg/postgres"
)

type bankAccounts interface {
	GetAllBankAccounts() ([]bankAcc, error)
	CreateBankAccount(a bankAcc) (Ans, error)
	DeleteBankAccount(a bankAcc) (Ans, error)
	UpdateBankAccount(a bankAcc) (Ans, error)
}

type legalEntitiRepository struct {
	db *postgres.GDB
}

func (r *legalEntitiRepository) GetAllBankAccounts() ([]bankAcc, error) {
	var accList []bankAcc
	err := r.db.DB.Find(&accList).Error
	return accList, err
}

func (r *legalEntitiRepository) CreateBankAccount(a bankAcc) (Ans, error) {
	result := r.db.DB.Create(&a)
	if result.Error != nil {

		ans := Ans{
			ChangedCount: int(result.RowsAffected),
			CorrAcc:      a.corAcc,
		}
		return ans, result.Error
	}
	ans := Ans{
		ChangedCount: int(result.RowsAffected),
		CorrAcc:      a.corAcc,
	}
	return ans, result.Error
}

func (r *legalEntitiRepository) DeleteBankAccount(a bankAcc) (Ans, error) {
	result := r.db.DB.Delete(a)
	var ans Ans
	if result.RowsAffected == 0 {
		ans.ChangedCount = int(result.RowsAffected)
		ans.CorrAcc = a.corAcc
	}
	ans.ChangedCount = int(result.RowsAffected)
	ans.CorrAcc = a.corAcc

	return ans, result.Error
}

func (r *legalEntitiRepository) UpdateBankAccount(a bankAcc) (Ans, error) {
	result := r.db.DB.Model(&a).Omit("curAcc").Updates(map[string]interface{}{
		"name": a.name, "address": a.address,
		"curAcc": a.curAcc, "corAcc": a.corAcc, "currency": a.currency, "comment": a.comment,
	})
	var ans Ans
	ans.ChangedCount = int(result.RowsAffected)
	ans.CorrAcc = a.corAcc

	return ans, result.Error
}
