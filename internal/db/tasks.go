package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func GetTaskDataByType(type_s string) (*model.TaskItem, error) {
	task := model.TaskItem{Key: type_s}
	if err := db.Where(task).First(&task).Error; err != nil {
		return nil, errors.Wrapf(err, "failed find task")
	}
	return &task, nil
}

func UpdateTaskData(t *model.TaskItem) error {
	return errors.WithStack(db.Model(&model.TaskItem{}).Where("key = ?", t.Key).Update("persist_data", t.PersistData).Error)
}

func CreateTaskData(t *model.TaskItem) error {
	return errors.WithStack(db.Create(t).Error)
}

func GetTaskDataFunc(type_s string, enabled bool) func() ([]byte, error) {
	if !enabled {
		return nil
	}
	task, err := GetTaskDataByType(type_s)
	if err != nil {
		return nil
	}
	return func() ([]byte, error) {
		return []byte(task.PersistData), nil
	}
}

func UpdateTaskDataFunc(type_s string, enabled bool) func([]byte) error {
	if !enabled {
		return nil
	}
	return func(data []byte) error {
		s := string(data)
		if s == "null" || s == "" {
			s = "[]"
		}
		return UpdateTaskData(&model.TaskItem{Key: type_s, PersistData: s})
	}
}
