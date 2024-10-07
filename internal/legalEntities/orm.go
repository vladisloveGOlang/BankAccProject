package legalEntities

// Банковский счет.
type bankAcc struct {
	name     string // Имя банка
	address  string // Юр. Ардес
	curAcc   string // рассчетный счет
	corAcc   string // корреспондентский счет
	currency string // Тип валюьы
	comment  string // Клментарий
}

type Ans struct {
	ChangedCount int    `json:"Changed,omitempty"`
	CorrAcc      string `json:"Corr.acc.,omitempty"`
}
