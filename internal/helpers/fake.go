package helpers

import (
	"github.com/brianvoe/gofakeit/v6"
)

func FakeSentence(length int) string {
	return gofakeit.Sentence(length)[0 : length-1]
}

func FakeString(l uint) string {
	return gofakeit.LetterN(l)
}

func FakeEmail() string {
	return gofakeit.Email()
}

func FakeName() string {
	return gofakeit.Name()
}

func FakeFName() string {
	return gofakeit.FirstName()
}

func FakePName() string {
	return gofakeit.MiddleName()
}

func FakeLName() string {
	return gofakeit.LastName()
}

func FakePhone() int {
	return RandomNumber(79000000000, 79299999999)
}

func FakeAddress() string {
	return gofakeit.Address().Address
}

func FakeTag() string {
	tags := []string{"frontend", "backend", "fullstack", "devops", "design", "marketing", "management", "hr", "sales", "support", "qa", "analytics", "seo", "smm", "copywriting", "content", "product", "other"}

	_, r := RandomFromSlice(tags)

	return r
}

func FakeEmails(min, max int) []string {
	res := []string{}

	for i := min; i < RandomNumber(min, max); i++ {
		res = append(res, FakeEmail())
	}

	return res
}
