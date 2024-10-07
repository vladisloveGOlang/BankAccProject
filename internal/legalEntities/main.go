package legalEntities

type LegalEntitiesService struct {
	repo legalEntitiRepository
}

func createLegalEntitiesService(repo legalEntitiRepository) *LegalEntitiesService {
	return &LegalEntitiesService{repo: repo}
}

func (s *LegalEntitiesService) GetAllBankAccounts() ([]bankAcc, error) {
	return s.repo.GetAllBankAccounts()
}

func (s *LegalEntitiesService) CreateBankAccount(a bankAcc) (Ans, error) {
	return s.repo.CreateBankAccount(a)
}

func (s *LegalEntitiesService) DeleteBankAccount(a bankAcc) (Ans, error) {
	return s.DeleteBankAccount(a)
}

func (s *LegalEntitiesService) UpdateBankAccount(a bankAcc) (Ans, error) {
	return s.repo.UpdateBankAccount(a)
}
