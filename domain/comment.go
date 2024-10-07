package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/internal/helpers"
)

type Comment struct {
	UUID      uuid.UUID
	ID        uint
	Comment   string `validate:"required,trim,gte=2,lte=5000"  ru:"комментарий"`
	CreatedBy string `validate:"required,email,lte=100,gte=3"  ru:"автор (email)"`

	ReplyUUID    *uuid.UUID `validate:"uuid"  ru:"комментарий (uuid)"`
	ReplyComment *string

	TaskUUID uuid.UUID `validate:"required,uuid"  ru:"задача (uuid)"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	Files []File

	People      map[string]int64 `ru:"люди (email)"`
	PeopleAdded []UserLike

	Likes     map[string]int64
	UserLikes []UserLike

	Pin bool

	Meta map[string]interface{}
}

func NewComment(createdBy string, taskUUID, replyUUID uuid.UUID, emails []string, msg string) *Comment {
	people := make(map[string]int64)
	for _, p := range emails {
		people[p] = time.Now().UnixMicro()
	}

	comment := &Comment{
		UUID:      uuid.New(),
		CreatedBy: createdBy,
		TaskUUID:  taskUUID,
		ReplyUUID: &replyUUID,

		People:  people,
		Comment: msg,

		Likes:     make(map[string]int64),
		UserLikes: []UserLike{},
	}

	errs, ok := helpers.ValidationStruct(comment)
	if !ok {
		panic(errors.New(helpers.Join(errs, ", ")))
	}

	return comment
}
