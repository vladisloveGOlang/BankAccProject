package sms

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
)

var codeStatus = map[int]string{
	-1:  "Not found",
	100: "Success",
	101: "The messege is passed to operator",
	102: "The message sent (in transit)",
	103: "The message was delivered",
	104: "Cannot be delivered: Time of life expired",
	105: "Cannot be delivered: deleted by operator",
	106: "Cannot be delivered: phone failure",
	107: "Cannot be delivered: unknown reason",
	108: "Cannot be delivered: rejected",
	130: "Cannot be delivered: Daily message limit on this number was exceeded",
	131: "Cannot be delivered: Same messages limit on this phone number in a minute was exceeded",
	132: "Cannot be delivered: Same messages limit on this phone number in a day was exceeded",
	200: "Wrong apiId",
	201: "Too low balance",
	202: "Wrong recipient",
	203: "The message has no text",
	204: "Sender name did not approve with administration",
	205: "The message is too long (more than 8 sms)",
	206: "Daily message limit exceeded",
	207: "On this phone number (or one of them) must not send the messages, or you indicated more than 100 phone numbers",
	208: "Wrong time value",
	209: "You added this phone number (or one of them) in the stop-list",
	210: "You must use a POST, not a GET",
	211: "Method not found",
	212: "Text of message must be in UTF-8",
	220: "The service is not available now, try again later",
	230: "Daily message limit on this number was exceeded",
	231: "Same messages limit on this phone number in a minute was exceeded",
	232: "Same messages limit on this phone number in a day was exceeded",
	300: "Wrong token (maybe it was expired or your IP was changed)",
	301: "Wrong password, or user is not exist",
	302: "User was authorized, but account is not activate",
	901: "Wrong Url (should begin with 'HTTP://')",
	902: "Callback is not defined",
}

var (
	errInternal   = errors.New("internal error")
	errNoResponse = errors.New("something went wrong")
)

func New(repo *Repository) *Service {
	return NewWithHTTP(&http.Client{}, repo)
}

func NewWithHTTP(client *http.Client, repo *Repository) *Service {
	c := &Service{
		APIURL: "https://sms.ru",
		HTTP:   client,

		repo: repo,
	}

	return c
}

func NewCompanySms(to, text, from string, senderUUID uuid.UUID, senderEmail string, company *dto.CompanyDTO) *domain.Sms {
	return &domain.Sms{
		UUID:           uuid.New(),
		FederationUUID: company.FederationUUID,
		CompanyUUID:    company.UUID,
		CreatedByUUID:  senderUUID,
		CreatedBy:      senderEmail,
		To:             to,
		Text:           text,
		From:           from,
	}
}

func NewSms(to, text string) *domain.Sms {
	return &domain.Sms{
		To:   to,
		Text: text,
	}
}

func NewMulti(sms ...*domain.Sms) *domain.Sms {
	arr := make(map[string]string)
	for _, o := range sms {
		arr[o.To] = o.Text
	}

	return &domain.Sms{
		Multi: arr,
	}
}

func (c *Service) makeRequest(endpoint, id string, params url.Values) (Response, []string, error) {
	params.Set("api_id", id)
	aPIURL := c.APIURL + endpoint + "?" + params.Encode()

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, aPIURL, http.NoBody)
	if err != nil {
		return Response{}, nil, err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Response{}, nil, err
	}
	defer resp.Body.Close()

	sc := bufio.NewScanner(resp.Body)
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	if err := sc.Err(); err != nil {
		return Response{}, nil, errInternal
	}

	if len(lines) == 0 {
		return Response{}, nil, errNoResponse
	}

	status, err := strconv.Atoi(lines[0])
	if err != nil {
		return Response{}, nil, errInternal
	}

	if status >= 200 {
		msg := fmt.Sprintf("Code: %d; Status: %s", status, codeStatus[status])
		return Response{}, nil, errors.New(msg)
	}

	res := Response{Status: status}
	return res, lines, nil
}

func (c *Service) StoreSms(s *domain.Sms) error {
	return c.repo.Create(s)
}

func (c *Service) SmsSend(id string, p *domain.Sms) (Response, error) {
	params := url.Values{}

	params.Set("from", p.From)

	params.Set("to", p.To)
	if len(p.Multi) > 0 {
		for to, text := range p.Multi {
			key := fmt.Sprintf("multi[%s]", to)
			params.Add(key, text)
		}
	} else {
		params.Set("to", p.To)
		params.Set("text", p.Text)
	}

	if len(p.From) > 0 {
		params.Set("from", p.From)
	}

	if p.PartnerID > 0 {
		val := strconv.Itoa(p.PartnerID)
		params.Set("partner_id", val)
	}

	if p.Test {
		params.Set("test", "1")
	}

	if p.Time.After(time.Now()) {
		val := strconv.FormatInt(p.Time.Unix(), 10)
		params.Set("time", val)
	}

	if p.Translit {
		params.Set("translit", "1")
	}

	res, lines, err := c.makeRequest("/sms/send", id, params)
	if err != nil {
		return Response{}, err
	}

	var ids []string
	re := regexp.MustCompile("^balance=")

	for i := 1; i < len(lines); i++ {
		isBalance := re.MatchString(lines[i])

		if isBalance {
			str := re.ReplaceAllString(lines[i], "")
			balance, err := strconv.ParseFloat(str, 32)
			if err != nil {
				return Response{}, errInternal
			}
			res.Balance = float32(balance)
		} else {
			ids = append(ids, lines[i])
		}
	}

	res.Ids = ids
	return res, nil
}

func (c *Service) SmsStatus(id string) (Response, error) {
	params := url.Values{}
	params.Set("id", id)

	res, _, err := c.makeRequest("/sms/status", id, params)
	if err != nil {
		return Response{}, err
	}

	return res, nil
}

func (c *Service) SmsCost(id string, p *domain.Sms) (Response, error) {
	params := url.Values{}
	params.Set("from", p.From)
	params.Set("to", p.To)
	params.Set("text", p.Text)
	if p.Translit {
		params.Set("translit", "1")
	}

	res, lines, err := c.makeRequest("/sms/cost", id, params)
	if err != nil {
		return Response{}, err
	}

	cost, err := strconv.ParseFloat(lines[1], 32)
	if err != nil {
		return Response{}, errInternal
	}

	count, err := strconv.Atoi(lines[2])
	if err != nil {
		return Response{}, errInternal
	}

	res.Cost = float32(cost)
	res.Count = count

	return res, nil
}

func (c *Service) MyBalance(id string) (Response, error) {
	res, lines, err := c.makeRequest("/my/balance", id, url.Values{})
	if err != nil {
		return Response{}, err
	}

	balance, err := strconv.ParseFloat(lines[1], 32)
	if err != nil {
		return Response{}, errInternal
	}

	res.Balance = float32(balance)
	return res, nil
}

// MyLimit checks the limit.
func (c *Service) MyLimit(id string) (Response, error) {
	res, lines, err := c.makeRequest("/my/limit", id, url.Values{})
	if err != nil {
		return Response{}, err
	}

	limit, err := strconv.Atoi(lines[1])
	if err != nil {
		return Response{}, errInternal
	}

	limitSent, err := strconv.Atoi(lines[2])
	if err != nil {
		return Response{}, errInternal
	}

	res.Limit = limit
	res.LimitSent = limitSent
	return res, nil
}

// MySenders receives the list of senders.
func (c *Service) MySenders(id string) (Response, error) {
	res, lines, err := c.makeRequest("/my/senders", id, url.Values{})
	if err != nil {
		return Response{}, err
	}

	var senders []string
	for i := 1; i < len(lines); i++ {
		senders = append(senders, lines[i])
	}

	res.Senders = senders
	return res, nil
}

// StoplistGet receives the stoplist.
func (c *Service) StoplistGet(id string) (Response, error) {
	res, lines, err := c.makeRequest("/stoplist/get", id, url.Values{})
	if err != nil {
		return Response{}, err
	}

	stoplist := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		re := regexp.MustCompile(";")
		str := re.Split(lines[i], 2)

		stoplist[str[0]] = str[1]
	}

	res.Stoplist = stoplist
	return res, nil
}

func (c *Service) StoplistAdd(id, phone, text string) (Response, error) {
	params := url.Values{}
	params.Set("stoplist_phone", phone)
	params.Set("stoplist_text", text)

	res, _, err := c.makeRequest("/stoplist/add", id, params)
	if err != nil {
		return Response{}, err
	}

	return res, nil
}

// StoplistDel will delete the phone number from stoplist
//
// phone is phone number.
func (c *Service) StoplistDel(id, phone string) (Response, error) {
	params := url.Values{}
	params.Set("stoplist_phone", phone)

	res, _, err := c.makeRequest("/stoplist/del", id, params)
	if err != nil {
		return Response{}, err
	}

	return res, nil
}

// CallbackGet receives the callbacks from service.
func (c *Service) CallbackGet(id string) (Response, error) {
	res, lines, err := c.makeRequest("/callback/get", id, url.Values{})
	if err != nil {
		return Response{}, err
	}

	var callbacks []string
	for i := 1; i < len(lines); i++ {
		callbacks = append(callbacks, lines[i])
	}

	res.Callbacks = callbacks
	return res, nil
}

func (c *Service) CallbackAdd(id, cbURL string) (Response, error) {
	params := url.Values{}
	params.Set("url", cbURL)

	res, lines, err := c.makeRequest("/callback/add", id, params)
	if err != nil {
		return Response{}, err
	}

	var callbacks []string
	for i := 1; i < len(lines); i++ {
		callbacks = append(callbacks, lines[i])
	}

	res.Callbacks = callbacks
	return res, nil
}

func (c *Service) CallbackDel(id, cbURL string) (Response, error) {
	params := url.Values{}
	params.Set("url", cbURL)

	res, lines, err := c.makeRequest("/callback/del", id, params)
	if err != nil {
		return Response{}, err
	}

	var callbacks []string
	for i := 1; i < len(lines); i++ {
		callbacks = append(callbacks, lines[i])
	}

	res.Callbacks = callbacks
	return res, nil
}

func (c *Service) GetSms(ctx context.Context, filter dto.SmsFilterDTO) ([]domain.Sms, int64, error) {
	return c.repo.GetSms(ctx, filter)
}
