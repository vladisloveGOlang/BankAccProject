package emails

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"regexp"
	"strings"

	"github.com/krisch/crm-backend/internal/configs"
	"github.com/sirupsen/logrus"
)

type IEmailsService interface {
	SendEmail(to []string, message IMessage) error
}

type Emails struct {
	from     string
	password string
	smtpHost string
	smtpPort string

	enable bool

	repo *EmailRepository
}

func New(from, password, smtpHost, smtpPort string, enable bool, r *EmailRepository) IEmailsService {
	return &Emails{
		from:     from,
		password: password,
		smtpHost: smtpHost,
		smtpPort: smtpPort,

		enable: enable,

		repo: r,
	}
}

func NewFromCreds(conf *configs.Configs, r *EmailRepository) (IEmailsService, error) {
	// parse string: smtp://johndoe@yandex.ru:xxxxxx:smtp.yandex.ru:465
	pattern := regexp.MustCompile(`smtp://(?P<password>[^:]+):(?P<user>[^:]+):(?P<host>[^:]+):(?P<port>[^:]+)`)
	sub := pattern.FindStringSubmatch(conf.SMTP_CREDS)

	if len(sub) != 5 {
		return &Emails{}, fmt.Errorf("invalid smtp connection string")
	}

	password := sub[1]
	user := sub[2]
	host := sub[3]
	port := sub[4]

	return New(user, password, host, port, conf.SMTP_ENABLE, r), nil
}

func (e *Emails) SendEmail(to []string, message IMessage) error {
	if e.enable {

		header := ""
		header += fmt.Sprintf("From: %s\r\n", e.from)

		header += fmt.Sprintf("To: %s\r\n", strings.Join(to, ";"))

		subject := "Subject: " + message.GetSubject() + "\n"
		mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		body := message.GetBody() + "\n"
		msg := []byte(header + subject + mime + body)

		// Create authentication
		auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)

		logrus.Debug(e.from, e.password, e.smtpHost, e.smtpHost+":"+e.smtpPort, msg)

		tlsConfig := &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
			ServerName:         e.smtpHost,
		}

		conn, err := tls.Dial("tcp", e.smtpHost+":"+e.smtpPort, tlsConfig)
		if err != nil {
			return err
		}

		client, err := smtp.NewClient(conn, e.smtpHost)
		if err != nil {
			return err
		}

		// step 1: Use Auth
		err = client.Auth(auth)
		if err != nil {
			return err
		}

		// step 2: add all from and to
		err = client.Mail(e.from)
		if err != nil {
			return err
		}

		for _, k := range to {
			logrus.Debug("sending to: ", k)
			err = client.Rcpt(k)
			if err != nil {
				return err
			}
		}

		// Data
		w, err := client.Data()
		if err != nil {
			return err
		}

		_, err = w.Write(msg)
		if err != nil {
			return err
		}

		err = w.Close()
		if err != nil {
			return err
		}

		err = client.Quit()
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"module":  "emails",
			"to":      to,
			"from":    e.from,
			"subject": message.GetSubject(),
		}).Debug("Email was sent to: ", to)
	}

	_, err := e.repo.StoreEmail(message)
	if err != nil {
		return err
	}

	return nil
}
