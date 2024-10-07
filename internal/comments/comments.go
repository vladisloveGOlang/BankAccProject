package comments

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/activities"
	"github.com/krisch/crm-backend/internal/dictionary"
	"github.com/krisch/crm-backend/internal/s3"
	"github.com/samber/lo"
)

type Service struct {
	repo    *Repository
	dict    *dictionary.Service
	storage *s3.ServicePrivate
	act     *activities.Service
}

func New(repo *Repository, dict *dictionary.Service, storage *s3.ServicePrivate, act *activities.Service) *Service {
	return &Service{
		repo:    repo,
		dict:    dict,
		storage: storage,
		act:     act,
	}
}

func (s *Service) GetTaskComments(uid uuid.UUID, withFiles, withLikes bool) (dms []domain.Comment, err error) {
	dms, err = s.repo.GetTaskComments(uid, nil)
	if err != nil {
		return dms, err
	}

	// People
	for i, dm := range dms {
		emails := lo.Keys(dm.People)

		usersDTO, _ := s.dict.FindUsers(emails)
		dms[i].PeopleAdded = lo.Map(usersDTO, func(u dto.UserDTO, idx int) domain.UserLike {
			return domain.UserLike{
				User: domain.User{
					UUID:     u.UUID,
					Email:    u.Email,
					Name:     u.Name,
					Lname:    u.Lname,
					Pname:    u.Pname,
					HasPhoto: u.HasPhoto,
				},
				CreatedAt: dm.People[u.Email],
			}
		})
	}

	// Files
	if withFiles {
		for i, dm := range dms {
			files, err := s.storage.GetCommentFiles(dm.UUID, true)
			if err != nil {
				return dms, err
			}

			dms[i].Files = files
		}
	}

	// Likes
	if withLikes {
		for i, dm := range dms {
			emails := lo.Keys(dm.Likes)

			usersDTO, _ := s.dict.FindUsers(emails)

			dms[i].UserLikes = lo.Map(usersDTO, func(u dto.UserDTO, idx int) domain.UserLike {
				return domain.UserLike{
					User: domain.User{
						UUID:     u.UUID,
						Email:    u.Email,
						Name:     u.Name,
						Lname:    u.Lname,
						Pname:    u.Pname,
						HasPhoto: u.HasPhoto,
					},
					CreatedAt: dm.Likes[u.Email],
				}
			})
		}
	}

	return dms, err
}

func (s *Service) GetCommentsFiles(uid uuid.UUID) (files []domain.File, err error) {
	files, err = s.storage.GetCommentFiles(uid, true)
	if err != nil {
		return files, err
	}

	return files, err
}

func (s *Service) GetTaskCommentsFiles(uid uuid.UUID) (files []domain.File, err error) {
	dms, err := s.repo.GetTaskComments(uid, nil)
	if err != nil {
		return files, err
	}

	for _, dm := range dms {
		cf, err := s.storage.GetCommentFiles(dm.UUID, true)
		if err != nil {
			return files, err
		}

		files = append(files, cf...)
	}

	return files, err
}

func (s *Service) CheckCommentText(uid uuid.UUID) (str string, err error) {
	return s.repo.GetCommentText(uid)
}

func (s *Service) CreateComment(ctx context.Context, comment domain.Comment) (err error) {
	foundUsers, _ := s.dict.FindUsers(lo.Keys(comment.People))

	if len(foundUsers) != len(comment.People) {
		return dto.NotFoundErr("один из пользователей не найден")
	}

	err = s.repo.CreateComment(ctx, comment)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) UpdateComment(ctx context.Context, comment domain.Comment) (err error) {
	foundUsers, _ := s.dict.FindUsers(lo.Keys(comment.People))

	if len(foundUsers) != len(comment.People) {
		return dto.NotFoundErr("один из пользователей не найден")
	}

	err = s.repo.UpdateComment(ctx, comment)
	if err != nil {
		return err
	}

	return err
}

func (s *Service) GetComment(_ context.Context, commentUUID uuid.UUID) (dt domain.Comment, err error) {
	dt, err = s.repo.GetTaskComment(commentUUID)
	if err != nil {
		return dt, err
	}

	if dt.UUID == uuid.Nil {
		return dt, dto.NotFoundErr("комментарий не найден")
	}

	return dt, nil
}

func (s *Service) LikeComment(_ context.Context, commentUUID uuid.UUID, userEmail string) (dtos []dto.UserDTO, liked bool, err error) {
	comment, err := s.repo.GetComment(commentUUID)
	if err != nil {
		return dtos, liked, err
	}

	userEmails := comment.Likes

	_, ok := userEmails[userEmail]

	if !ok {
		liked = true
		userEmails[userEmail] = time.Now().UnixMicro()
	} else {
		liked = false
		delete(userEmails, userEmail)
	}

	err = s.repo.PatchCommentLikes(commentUUID, userEmails)

	dtos, _ = s.dict.FindUsers(lo.Keys(userEmails))

	return dtos, liked, err
}

func (s *Service) PinComment(_ context.Context, commentUUID uuid.UUID) (err error) {
	err = s.repo.PatchCommentPin(commentUUID)

	return err
}

func (s *Service) DeleteComment(ctx context.Context, taskUUID, commentUID uuid.UUID, _ *uuid.UUID) (err error) {
	err = s.repo.DeleteComment(ctx, taskUUID, commentUID)

	if err != nil {
		return err
	}

	files, err := s.GetCommentsFiles(commentUID)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = s.storage.Delete(file.UUID)
		if err != nil {
			return err
		}
	}

	return err
}

func (s *Service) DeleteCommentFile(commentUID, fileUID uuid.UUID) (err error) {
	files, err := s.GetCommentsFiles(commentUID)
	if err != nil {
		return err
	}
	file, f := lo.Find(files, func(f domain.File) bool {
		return f.UUID == fileUID
	})
	if !f {
		return dto.NotFoundErr("файл не найден")
	}

	err = s.storage.Delete(file.UUID)

	return err
}
