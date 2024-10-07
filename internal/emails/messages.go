package emails

import (
	"bytes"
	"text/template"

	_ "embed"
)

type IMessage interface {
	GetSubject() string
	GetBody() string
}

type Message struct {
	subject string
	body    string
}

func (m Message) GetSubject() string {
	return m.subject
}

func (m Message) GetBody() string {
	return m.body
}

// Confirmation email template
//
//go:embed confirmation.html
var confirmationTmpl string

//go:embed reset.html
var resetTmpl string

func NewConfirmationMessage(code string) (IMessage, error) {
	templateData := struct {
		Code string
	}{
		Code: code,
	}

	body, err := parseTemplate("confirmation", confirmationTmpl, templateData)

	if err == nil {
		return Message{
			subject: "Подтверждение профиля",
			body:    body,
		}, nil
	}

	return Message{}, nil
}

func NewResetMessage(code string) (IMessage, error) {
	templateData := struct {
		Code string
	}{
		Code: code,
	}

	body, err := parseTemplate("reset", resetTmpl, templateData)

	if err == nil {
		return Message{
			subject: "Сброс пароля",
			body:    body,
		}, nil
	}

	return Message{}, nil
}

func parseTemplate(name, templateString string, data interface{}) (string, error) {
	t, err := template.New(name).Parse(templateString)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
