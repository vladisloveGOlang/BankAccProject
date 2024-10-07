package helpers

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ru_translator "github.com/go-playground/validator/v10/translations/ru"
	"github.com/sirupsen/logrus"
)

//

// Validations.
var validate22 *validator.Validate = validator.New()

func ValidateEmail(s string) error {
	err := validate22.Var(s, "required,email")
	if err != nil {
		return fmt.Errorf("[email:%s] email is not valid", s)
	}

	return nil
}

func ValidateOptionalEmail(s string) error {
	err := validate22.Var(s, "email")
	if err != nil {
		return fmt.Errorf("[email:%s] email is not valid", s)
	}

	return nil
}

func ValidateColor(color string) error {
	if len(color) != 7 {
		return fmt.Errorf("[color:%s] цвет должен быть #000000 (#xxxxxx)", color)
	}

	if color[0] != '#' {
		return fmt.Errorf("[color:%s] цвет должен начинаться с # (#xxxxxx)", color)
	}

	if _, err := strconv.ParseInt(color[1:], 16, 64); err != nil {
		return fmt.Errorf("[color:%s] цвет может содержать только цифры #000000 (#xxxxxx)", color)
	}

	return nil
}

func HTTPSValidation(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return strings.HasPrefix(value, "https://")
}

func TrimValidation(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return !(strings.HasPrefix(value, " ") || strings.HasSuffix(value, " "))
}

func ColorValidation(fl validator.FieldLevel) bool {
	color := fl.Field().String()

	if len(color) != 7 {
		return false
	}

	if color[0] != '#' {
		return false
	}

	if _, err := strconv.ParseInt(color[1:], 16, 64); err != nil {
		return false
	}

	return true
}

func NameValidation(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	rgxp := regexp.MustCompile(`^[\p{L}0-9\s\(\)-_]*$`)
	if !rgxp.MatchString(value) {
		panic(errors.New("название должно содержать только буквы, цифры и пробелы"))
	}

	return !(strings.HasPrefix(value, " ") || strings.HasSuffix(value, " "))
}

func ValidateLegalEntityField(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if !regexp.MustCompile(`^\d+$`).MatchString(value) {
		return false
	}
	return !regexp.MustCompile(`^0{2,}`).MatchString(value)
}

func OptionalEmailValidation(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true
	}

	err := validate22.Var(value, "required,email")
	return err == nil
}

func ValidationStruct[T any](str T, fields ...string) ([]string, bool) {
	russian := ru.New()

	uni := ut.New(russian, russian)

	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	trans, found := uni.GetTranslator("ru")
	if !found {
		logrus.Error("translator not found")
	}

	validate := validator.New()

	err := validate.RegisterValidation("is_https", HTTPSValidation)
	if err != nil {
		logrus.Error(err)
	}

	err = validate.RegisterValidation("trim", TrimValidation)
	if err != nil {
		logrus.Error(err)
	}

	err = validate.RegisterTranslation("trim", trans, func(ut ut.Translator) error {
		return ut.Add("trim", "нужно убрать пробелы", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T("trim", fe.Field(), fe.Param())
		if err != nil {
			logrus.Error(err)
		}

		return t
	})
	if err != nil {
		logrus.Error(err)
	}

	// Color
	err = validate.RegisterValidation("color", ColorValidation)
	if err != nil {
		logrus.Error(err)
	}

	err = validate.RegisterTranslation("color", trans, func(ut ut.Translator) error {
		return ut.Add("color", "цвет должен быть #000000 - #999999", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T("color", fe.Field(), fe.Param())
		if err != nil {
			logrus.Error(err)
		}

		return t
	})
	if err != nil {
		logrus.Error(err)
	}

	// Name
	err = validate.RegisterValidation("name", NameValidation)
	if err != nil {
		logrus.Error(err)
	}

	err = validate.RegisterTranslation("name", trans, func(ut ut.Translator) error {
		return ut.Add("name", "название а-Я и цифры 0-9", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T("name", fe.Field(), fe.Param())
		if err != nil {
			logrus.Error(err)
		}

		return t
	})
	if err != nil {
		logrus.Error(err)
	}

	// Legal entity
	err = validate.RegisterValidation("legal_entity_field", ValidateLegalEntityField)
	if err != nil {
		logrus.Error(err)
	}

	err = validate.RegisterTranslation("legal_entity_field", trans, func(ut ut.Translator) error {
		return ut.Add("legal_entity", "{0} должно быть числом и начинаться с не более чем одного нуля", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T("legal_entity", fe.Field(), fe.Param())
		if err != nil {
			logrus.Error(err)
		}

		return t
	})
	if err != nil {
		logrus.Error(err)
	}

	err = ru_translator.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		logrus.Error(err)
	}

	// Registrer email
	err = validate.RegisterTranslation("email", trans, func(ut ut.Translator) error {
		return ut.Add("email", "Почта должна быть действительной", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T("email", fe.Field(), fe.Param())
		if err != nil {
			logrus.Error(err)
		}

		return t
	})
	if err != nil {
		logrus.Error(err)
	}

	// Registrer optional email
	err = validate.RegisterValidation("optional_email", OptionalEmailValidation)
	if err != nil {
		logrus.Error(err)
	}

	err = validate.RegisterTranslation("optional_email", trans, func(ut ut.Translator) error {
		return ut.Add("optional_email", "Почта должна быть действительной или пустой", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T("optional_email", fe.Field(), fe.Param())
		if err != nil {
			logrus.Error(err)
		}

		return t
	})
	if err != nil {
		logrus.Error(err)
	}

	//
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("ru"), ",", 2)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}
		return name
	})

	//
	if len(fields) > 0 {
		err = validate.StructPartial(str, fields...)
	} else {
		err = validate.Struct(str)
	}

	if err != nil {
		var errs validator.ValidationErrors

		if errors.As(err, &errs) {
			ers := Map(errs, func(e validator.FieldError, _ int) string {
				return e.Translate(trans)
			})

			return ers, false
		}
	}

	return []string{}, true
}
