package aggregates

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/comments"
	"github.com/krisch/crm-backend/internal/dictionary"
	"github.com/krisch/crm-backend/internal/federation"
	"github.com/krisch/crm-backend/internal/profile"
	"github.com/krisch/crm-backend/internal/reminders"
	"github.com/krisch/crm-backend/internal/s3"
	"github.com/krisch/crm-backend/internal/task"
	"github.com/samber/lo"
)

type Service struct {
	dictionaryService *dictionary.Service
	ps                *profile.Service
	ts                *task.Service
	cs                *comments.Service
	s3ps              *s3.ServicePrivate
	rm                *reminders.Service
	federationService *federation.Service
}

func New(
	dictionaryService *dictionary.Service,
	ps *profile.Service,
	ts *task.Service,
	cs *comments.Service,
	s3ps *s3.ServicePrivate,
	rm *reminders.Service,
	federationService *federation.Service,
) *Service {
	return &Service{
		dictionaryService: dictionaryService,
		ps:                ps,
		ts:                ts,
		cs:                cs,
		s3ps:              s3ps,
		rm:                rm,
		federationService: federationService,
	}
}

func (s *Service) GetTaskWithFields(ctx context.Context, taskUUID uuid.UUID) (dt dto.TaskDTO, err error) {
	dm, err := s.ts.GetTaskGetTaskWithDeleted(ctx, taskUUID)
	if err != nil {
		return dt, err
	}

	taskComments, err := s.cs.GetTaskComments(dm.UUID, true, true)
	if err != nil {
		return dt, err
	}

	files, err := s.s3ps.GetTaskFiles(dm.UUID, true)
	if err != nil {
		return dt, err
	}

	taskReminders, err := s.rm.GetByTask(dm.UUID)
	if err != nil {
		return dt, err
	}

	taskDto := dto.NewTaskDTO(dm, taskComments, files, taskReminders, make(map[uuid.UUID]interface{}), s.dictionaryService, s.ps)

	return taskDto, nil
}

type StateDiff struct {
	NewComments  []dto.CommentDTO  `json:"new_comments"`
	NewLikes     int               `json:"new_likes"`
	NewMentions  int               `json:"new_mentions"`
	NewUploads   []dto.FileDTOs    `json:"new_uploads"`
	NewReminders []dto.ReminderDTO `json:"new_reminders"`

	UpdatedAt time.Time `json:"updated_at"`
}

func CompareState(taskDto dto.TaskDTO, userUUID uuid.UUID, fromTime time.Time) StateDiff {
	newMensions := 0
	newLikes := 0
	newComments := []dto.CommentDTO{}
	newUploads := []dto.FileDTOs{}
	newReminders := []dto.ReminderDTO{}

	score := taskDto.CreatedAt

	for _, c := range taskDto.Comments {
		if c.UpdatedAt.After(score) {
			score = c.UpdatedAt
		}

		if c.UpdatedAt.After(fromTime) {
			mentionedUser, ok := c.InPeople(userUUID)
			if ok {
				if time.Unix(mentionedUser.UnixAt, 0).After(fromTime) {
					newMensions++
				}
			}

			if c.CreatedBy.UUID != userUUID {
				newComments = append(newComments, c)
			}

			lo.Filter(c.Likes, func(l dto.UserLikeDTO, _ int) bool {
				if l.User.UUID != userUUID {
					newLikes++
					return true
				}

				return false
			})
		}
	}

	for _, f := range taskDto.Files {
		if f.CreatedAt.After(score) {
			score = f.CreatedAt
		}

		if f.CreatedAt.After(fromTime) {
			if f.CreatedBy.UUID != userUUID {
				newUploads = append(newUploads, f)
			}
		}
	}

	for _, r := range taskDto.Reminders {
		if r.UpdatedAt.After(score) {
			score = r.UpdatedAt
		}
		newReminders = append(newReminders, r)
	}

	if score.Before(taskDto.UpdatedAt) {
		score = taskDto.UpdatedAt
	}

	return StateDiff{
		NewComments:  newComments,
		NewMentions:  newMensions,
		NewLikes:     newLikes,
		NewUploads:   newUploads,
		NewReminders: newReminders,
		UpdatedAt:    score,
	}
}
