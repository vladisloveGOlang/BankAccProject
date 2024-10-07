package domain

import (
	"github.com/google/uuid"
)

type Creator struct {
	UUID  uuid.UUID
	Email string
}

type IUser interface {
	GetEmail() string
	GetUUID() uuid.UUID
}

func NewCreator(uid uuid.UUID, email string) *Creator {
	return &Creator{
		UUID:  uid,
		Email: email,
	}
}

func NewCreatorFromUser(user IUser) Creator {
	return Creator{
		UUID:  user.GetUUID(),
		Email: user.GetEmail(),
	}
}
