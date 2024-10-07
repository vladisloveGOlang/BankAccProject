package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/samber/lo"
)

type CommentDTO struct {
	UUID uuid.UUID `json:"uuid"`

	Comment      string     `json:"comment"`
	ReplyUUID    *uuid.UUID `json:"reply_uuid,omitempty"`
	ReplyComment *string    `json:"reply_comment,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	CreatedBy *UserDTO      `json:"created_by"`
	People    []UserLikeDTO `json:"people,omitempty"`
	Likes     []UserLikeDTO `json:"likes,omitempty"`

	Uploads []UploadDTO `json:"files,omitempty"`

	Pin bool `json:"pin"`
}

func NewCommentDTO(dm domain.Comment, dict IDict, s3 IStorage) CommentDTO {
	createdBy, _ := dict.FindUser(dm.CreatedBy)

	uploads := lo.Map(dm.Files, func(file domain.File, i int) UploadDTO {
		return NewUploadDTO(file.UUID, file.Name, file.Ext, file.Size, file.URL)
	})

	likes := lo.Map(dm.UserLikes, func(user domain.UserLike, _ int) UserLikeDTO {
		return UserLikeDTO{
			UnixAt: user.CreatedAt,
			User:   NewUserShotDto(user.User, s3),
		}
	})

	people := lo.Map(dm.PeopleAdded, func(user domain.UserLike, _ int) UserLikeDTO {
		return UserLikeDTO{
			UnixAt: user.CreatedAt,
			User:   NewUserShotDto(user.User, s3),
		}
	})

	return CommentDTO{
		UUID:         dm.UUID,
		Comment:      dm.Comment,
		CreatedBy:    createdBy,
		ReplyUUID:    dm.ReplyUUID,
		ReplyComment: dm.ReplyComment,

		CreatedAt: dm.CreatedAt,
		UpdatedAt: dm.UpdatedAt,

		Uploads: uploads,

		Likes: likes,

		People: people,

		Pin: dm.Pin,
	}
}

func (d *CommentDTO) InPeople(userUUID uuid.UUID) (UserLikeDTO, bool) {
	f, ok := lo.Find(d.People, func(p UserLikeDTO) bool {
		return p.User.UUID == userUUID
	})

	return f, ok
}
