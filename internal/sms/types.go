package sms

import (
	"net/http"
)

type Service struct {
	APIURL string
	HTTP   *http.Client
	Debug  bool

	repo *Repository
}

type Response struct {
	Status    int               `json:"status"`
	Ids       []string          `json:"id"`
	Cost      float32           `json:"cost"`
	Count     int               `json:"count"`
	Balance   float32           `json:"balance"`
	Limit     int               `json:"limit"`
	LimitSent int               `json:"limit_sent"`
	Senders   []string          `json:"senders"`
	Stoplist  map[string]string `json:"stoplist"`
	Callbacks []string          `json:"callbacks"`
}
