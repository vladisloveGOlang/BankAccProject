package comments

import (
	"context"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/cache"
	"github.com/krisch/crm-backend/internal/helpers"
	"github.com/krisch/crm-backend/pkg/postgres"
	"github.com/krisch/crm-backend/pkg/redis"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type Repository struct {
	gorm      *postgres.GDB
	rds       *redis.RDS
	histogram *prometheus.HistogramVec
	cache     *cache.Service
}

func NewRepository(db *postgres.GDB, rds *redis.RDS, metrics *helpers.MetricsCounters, cs *cache.Service) *Repository {
	return &Repository{
		gorm:      db,
		rds:       rds,
		histogram: metrics.RepoHistogram,
		cache:     cs,
	}
}

func (r *Repository) storeTime(name string, time *helpers.Time) {
	func() { r.histogram.WithLabelValues(name).Observe(time.Secondsf()) }()
}

func tm() *helpers.Time {
	return helpers.NewTime()
}

func (r *Repository) CreateComment(ctx context.Context, cmnt domain.Comment) (err error) {
	defer r.storeTime("CreateComment", tm())

	err = r.gorm.DB.Transaction(func(tx *gorm.DB) error {
		orm := &Comment{
			UUID: cmnt.UUID,

			ReplyUUID: cmnt.ReplyUUID,
			TaskUUID:  cmnt.TaskUUID,
			Comment:   cmnt.Comment,

			CreatedBy: cmnt.CreatedBy,

			CreatedAt: cmnt.CreatedAt,

			People: cmnt.People,

			Likes: Persons{},
		}

		err := tx.Create(&orm).Error
		if err != nil {
			return err
		}

		err = tx.Exec("update tasks set comments_total = comments_total + 1 where uuid = ?", orm.TaskUUID).Error

		return err
	})

	if err == nil {
		go r.cache.ClearTask(ctx, cmnt.TaskUUID)
	}

	return err
}

func (r *Repository) UpdateComment(ctx context.Context, cmnt domain.Comment) (err error) {
	defer r.storeTime("UpdateComment", tm())

	err = r.gorm.DB.Transaction(func(tx *gorm.DB) error {
		orm := Comment{
			UUID:      cmnt.UUID,
			ReplyUUID: cmnt.ReplyUUID,
			Comment:   cmnt.Comment,
			People:    cmnt.People,
			Likes:     cmnt.Likes,
		}

		err := tx.
			Updates(orm).
			Where("uuid = ?", cmnt.UUID).
			Error
		if err != nil {
			return err
		}

		err = tx.Exec("update tasks set updated_at = now() where uuid = ?", cmnt.TaskUUID).Error

		return err
	})

	if err == nil {
		go r.cache.ClearTask(ctx, cmnt.TaskUUID)
	}

	return err
}

func (r *Repository) DeleteComment(ctx context.Context, taskUUID, uid uuid.UUID) (err error) {
	defer r.storeTime("DeleteComment", tm())

	err = r.gorm.DB.Transaction(func(tx *gorm.DB) error {
		orm := Comment{}

		r.gorm.DB.
			Where("uuid = ?", uid.String()).
			Where("deleted_at IS NULL").
			First(&orm)

		if orm.UUID == uuid.Nil {
			return dto.NotFoundErr("комментарий не найден")
		}

		res := r.gorm.DB.
			Model(orm).
			Where("uuid = ?", uid.String()).
			Update("deleted_at", "now()")

		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			return dto.NotFoundErr("комментарий не найден")
		}

		err = r.gorm.DB.Exec("update tasks set comments_total = comments_total - 1 where uuid = ?", orm.TaskUUID).Error

		return err
	})

	if err == nil {
		go r.cache.ClearTask(ctx, taskUUID)
	}

	return err
}

func (r *Repository) GetTaskComments(uid uuid.UUID, commentUID *uuid.UUID) (dms []domain.Comment, err error) {
	defer r.storeTime("GetTaskComments", tm())

	orm := []Comment{}
	q := r.gorm.DB.
		Model(orm).
		Select("comments.uuid, comments.comment, comments.created_by, comments.reply_uuid, comments.task_uuid, comments.people, comments.created_at, comments.updated_at, comments.likes, comments.pin, c.comment as reply_comment").
		Where("comments.task_uuid = ?", uid).
		Where("comments.deleted_at IS NULL").
		Joins("LEFT JOIN comments c ON c.uuid = comments.reply_uuid").
		Order("comments.pin DESC, comments.created_at DESC").
		Limit(300)

	if commentUID != nil {
		q = q.Where("comments.uuid = ?", commentUID)
	}

	err = q.Find(&orm).Error

	for _, o := range orm {
		dms = append(dms, domain.Comment{
			UUID:         o.UUID,
			Comment:      o.Comment,
			CreatedBy:    o.CreatedBy,
			ReplyUUID:    o.ReplyUUID,
			ReplyComment: o.ReplyComment,
			TaskUUID:     o.TaskUUID,
			People:       o.People,
			CreatedAt:    o.CreatedAt,
			UpdatedAt:    o.UpdatedAt,
			Likes:        o.Likes,
			Pin:          o.Pin,
		})
	}

	return dms, err
}

func (r *Repository) GetTaskComment(commentUID uuid.UUID) (dms domain.Comment, err error) {
	defer r.storeTime("GetTaskComments", tm())

	orm := Comment{}
	err = r.gorm.DB.
		Model(orm).
		Select("comments.uuid, comments.comment, comments.created_by, comments.reply_uuid, comments.task_uuid, comments.people, comments.created_at, comments.updated_at, comments.likes, comments.pin, c.comment as reply_comment").
		Where("comments.deleted_at IS NULL").
		Joins("LEFT JOIN comments c ON c.uuid = comments.reply_uuid").
		Order("comments.pin DESC, comments.created_at DESC").
		Where("comments.uuid = ?", commentUID).
		Limit(1).
		First(&orm).
		Error

	return domain.Comment{
		UUID:         orm.UUID,
		Comment:      orm.Comment,
		CreatedBy:    orm.CreatedBy,
		ReplyUUID:    orm.ReplyUUID,
		ReplyComment: orm.ReplyComment,
		TaskUUID:     orm.TaskUUID,
		People:       orm.People,
		CreatedAt:    orm.CreatedAt,
		UpdatedAt:    orm.UpdatedAt,
		Likes:        orm.Likes,
		Pin:          orm.Pin,
	}, err
}

func (r *Repository) GetCommentText(uid uuid.UUID) (msg string, err error) {
	defer r.storeTime("GetCommentText", tm())

	orm := Comment{}
	res := r.gorm.DB.
		Model(orm).
		Select("comments.uuid, comments.comment").
		Where("comments.uuid = ?", uid).
		Where("comments.deleted_at IS NULL").
		First(&orm)

	if res.RowsAffected == 0 {
		return "", dto.NotFoundErr("комментарий не найден")
	}

	return orm.Comment, res.Error
}

func (r *Repository) GetComment(uid uuid.UUID) (comment Comment, err error) {
	defer r.storeTime("GetComment", tm())

	orm := Comment{}
	res := r.gorm.DB.
		Model(orm).
		Where("comments.uuid = ?", uid).
		Where("comments.deleted_at IS NULL").
		First(&orm)

	if res.RowsAffected == 0 {
		return comment, dto.NotFoundErr("комментарий не найден")
	}

	return orm, res.Error
}

func (r *Repository) PatchCommentLikes(uid uuid.UUID, likes map[string]int64) (err error) {
	defer r.storeTime("PatchCommentLikes", tm())

	res := r.gorm.DB.
		Model(&Comment{}).
		Where("comments.uuid = ?", uid).
		Where("comments.deleted_at IS NULL").
		Update("likes", likes).
		Update("updated_at", "now()")

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("комментарий не найден")
	}

	return res.Error
}

func (r *Repository) PatchCommentPin(uid uuid.UUID) (err error) {
	defer r.storeTime("PatchCommentPin", tm())

	res := r.gorm.DB.Exec("UPDATE comments SET pin = NOT pin, updated_at = now() WHERE comments.uuid = ? AND comments.deleted_at IS NULL", uid)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return dto.NotFoundErr("комментарий не найден")
	}

	return res.Error
}
