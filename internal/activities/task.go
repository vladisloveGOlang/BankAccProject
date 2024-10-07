package activities

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/krisch/crm-backend/domain"
	"github.com/krisch/crm-backend/dto"
	"github.com/krisch/crm-backend/internal/helpers"
)

func (s *Service) TaskWasChangedActivity(creator domain.Creator, taskUID uuid.UUID, field string, oldVal, newVal interface{}) (*Activity, error) {
	ActivityMeta := dto.ActivityTaskFieldDTO{
		Old:  oldVal,
		New:  newVal,
		Name: field,
	}

	mp, err := helpers.StructToMap(ActivityMeta)
	if err != nil {
		return nil, err
	}

	act := &Activity{
		UUID:          uuid.New(),
		EntityUUID:    taskUID,
		EntityType:    "task",
		Description:   fmt.Sprint(domain.ActivityTaskField),
		CreatedByUUID: creator.UUID,
		CreatedBy:     creator.Email,
		Type:          domain.ActivityTaskField,
		Meta:          mp,
	}

	if reflect.DeepEqual(mp["old"], mp["new"]) {
		return act, nil
	}

	err = s.CreateActivity(act)
	if err != nil {
		return nil, err
	}

	return act, nil
}

func (s *Service) TaskWasChangedStatusActivity(creator domain.Creator, taskUID uuid.UUID, oldVal, newVal dto.ProjectStatusDTOs) (*Activity, error) {
	ActivityMeta := dto.ActivityTaskStatusDTO{
		Old: oldVal.Number,
		New: newVal.Number,

		OldStatus: oldVal,
		NewStatus: newVal,
	}

	mp, err := helpers.StructToMap(ActivityMeta)
	if err != nil {
		return nil, err
	}

	act := &Activity{
		UUID:          uuid.New(),
		EntityUUID:    taskUID,
		EntityType:    "task",
		Description:   fmt.Sprint(domain.ActivityTaskStatus),
		CreatedByUUID: creator.UUID,
		CreatedBy:     creator.Email,
		Type:          domain.ActivityTaskStatus,
		Meta:          mp,
	}

	if reflect.DeepEqual(oldVal, newVal) {
		return act, nil
	}

	err = s.CreateActivity(act)
	if err != nil {
		return nil, err
	}

	return act, nil
}

func (s *Service) TaskWasChangedArrayActivity(creator domain.Creator, taskUID uuid.UUID, name string, oldVal, newVal []interface{}) (*Activity, error) {
	add := helpers.FindNewElements(oldVal, newVal)
	remove := helpers.FindRemovedElements(oldVal, newVal)

	ActivityMeta := dto.ActivityTaskFieldArrayDTO{
		Name:   name,
		Old:    oldVal,
		New:    newVal,
		Add:    add,
		Remove: remove,
	}

	mp, err := helpers.StructToMap(ActivityMeta)
	if err != nil {
		return nil, err
	}

	act := &Activity{
		UUID:          uuid.New(),
		EntityUUID:    taskUID,
		EntityType:    "task",
		Description:   fmt.Sprint(domain.ActivityTaskFieldArray),
		CreatedByUUID: creator.UUID,
		CreatedBy:     creator.Email,
		Type:          domain.ActivityTaskFieldArray,
		Meta:          mp,
	}

	if reflect.DeepEqual(mp["old"], mp["new"]) {
		return act, nil
	}

	err = s.CreateActivity(act)
	if err != nil {
		return nil, err
	}

	return act, nil
}

func (s *Service) TaskWasChangedTeamActivity(creator domain.Creator, taskUID uuid.UUID, name string, oldVal, newVal []dto.UserDTO) (*Activity, error) {
	add := helpers.FindNewElements(oldVal, newVal)
	remove := helpers.FindRemovedElements(oldVal, newVal)

	ActivityMeta := dto.ActivityTaskTeamDTO{
		Add:    add,
		Remove: remove,
		Name:   name,
	}

	mp, err := helpers.StructToMap(ActivityMeta)
	if err != nil {
		return nil, err
	}

	act := &Activity{
		UUID:          uuid.New(),
		EntityUUID:    taskUID,
		EntityType:    "task",
		Description:   fmt.Sprint(domain.ActivityTaskTeamArray),
		CreatedByUUID: creator.UUID,
		CreatedBy:     creator.Email,
		Type:          domain.ActivityTaskTeamArray,
		Meta:          mp,
	}

	if reflect.DeepEqual(oldVal, newVal) {
		return act, nil
	}

	err = s.CreateActivity(act)
	if err != nil {
		return nil, err
	}

	return act, nil
}

func (s *Service) TaskWasDeleted(creator domain.Creator, taskUID uuid.UUID, name string) (*Activity, error) {
	ActivityMeta := dto.ActivityTaskWasDeletedDTO{
		Name: name,
	}

	mp, err := helpers.StructToMap(ActivityMeta)
	if err != nil {
		return nil, err
	}

	act := &Activity{
		UUID:          uuid.New(),
		EntityUUID:    taskUID,
		EntityType:    "task",
		Description:   fmt.Sprint(domain.ActivityTaskWasDeleted),
		CreatedByUUID: creator.UUID,
		CreatedBy:     creator.Email,
		Type:          domain.ActivityTaskWasDeleted,
		Meta:          mp,
	}

	err = s.CreateActivity(act)
	if err != nil {
		return nil, err
	}

	return act, nil
}

func (s *Service) TaskFileWasDeleted(creator domain.Creator, taskUUID uuid.UUID, file domain.File) (*Activity, error) {
	ActivityMeta := dto.ActivityTaskFileWasDeletedDTO{
		Name: file.Name,
		Ext:  file.Ext,
		Size: file.Size,
	}

	mp, err := helpers.StructToMap(ActivityMeta)
	if err != nil {
		return nil, err
	}

	act := &Activity{
		UUID:          uuid.New(),
		EntityUUID:    taskUUID,
		EntityType:    "task",
		Description:   fmt.Sprint(domain.ActivityTaskFileWasDeleted),
		CreatedByUUID: creator.UUID,
		CreatedBy:     creator.Email,
		Type:          domain.ActivityTaskFileWasDeleted,
		Meta:          mp,
	}

	err = s.CreateActivity(act)
	if err != nil {
		return nil, err
	}

	return act, nil
}
