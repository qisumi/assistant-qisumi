package task

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTx 支持在事务中生成一个带 Tx 的 repo
func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

func (r *Repository) GetTaskWithSteps(ctx context.Context, userID, taskID uint64) (*Task, error) {
	var t Task
	err := r.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index ASC")
		}).
		Where("id = ? AND user_id = ?", taskID, userID).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) InsertTaskWithSteps(ctx context.Context, t *Task) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) ListTasks(ctx context.Context, userID uint64) ([]Task, error) {
	var tasks []Task
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// ApplyUpdateTaskFields 动态更新 tasks
func (r *Repository) ApplyUpdateTaskFields(
	ctx context.Context,
	userID, taskID uint64,
	fields UpdateTaskFields,
) error {
	updates, err := buildTaskUpdateMap(fields)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Model(&Task{}).
		Where("id = ? AND user_id = ?", taskID, userID).
		Updates(updates).Error
}

// ApplyUpdateStepFields 动态更新 task_steps 中的一行
func (r *Repository) ApplyUpdateStepFields(
	ctx context.Context,
	userID, taskID, stepID uint64,
	fields UpdateStepFields,
) error {
	updates, err := buildStepUpdateMap(fields)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}

	// 加上 user_id 保护：用子查询限制
	subQuery := r.db.
		Select("id").
		Table("tasks").
		Where("id = ? AND user_id = ?", taskID, userID)

	return r.db.WithContext(ctx).
		Model(&TaskStep{}).
		Where("id = ? AND task_id IN (?)", stepID, subQuery).
		Updates(updates).Error
}

// AddStep 添加新步骤
func (r *Repository) AddStep(ctx context.Context, step *TaskStep) error {
	return r.db.WithContext(ctx).Create(step).Error
}

// AddSteps 添加多个新步骤
func (r *Repository) AddSteps(ctx context.Context, steps []TaskStep) error {
	if len(steps) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&steps).Error
}

// AddDependency 添加任务依赖
func (r *Repository) AddDependency(ctx context.Context, dep *TaskDependency) error {
	return r.db.WithContext(ctx).Create(dep).Error
}

// AddDependencies 添加多个任务依赖
func (r *Repository) AddDependencies(ctx context.Context, dependencies []TaskDependency) error {
	if len(dependencies) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&dependencies).Error
}

// MarkTasksFocusToday 标记任务为今日重点
func (r *Repository) MarkTasksFocusToday(ctx context.Context, userID uint64, taskIDs []uint64) error {
	// TODO: 实现标记今日重点任务的逻辑
	return nil
}

// 把 UpdateTaskFields 转成 GORM Updates 使用s的 map
func buildTaskUpdateMap(fields UpdateTaskFields) (map[string]any, error) {
	updates := make(map[string]any)

	if fields.Title != nil {
		updates["title"] = *fields.Title
	}
	if fields.Description != nil {
		updates["description"] = *fields.Description
	}
	if fields.Status != nil {
		updates["status"] = *fields.Status
	}
	if fields.Priority != nil {
		updates["priority"] = *fields.Priority
	}
	if fields.DueAt != nil {
		if *fields.DueAt == "" {
			updates["due_at"] = nil
		} else {
			t, err := time.Parse(time.RFC3339, *fields.DueAt)
			if err != nil {
				return nil, err
			}
			updates["due_at"] = t
		}
	}

	return updates, nil
}

func buildStepUpdateMap(fields UpdateStepFields) (map[string]any, error) {
	updates := make(map[string]any)

	if fields.Title != nil {
		updates["title"] = *fields.Title
	}
	if fields.Detail != nil {
		updates["detail"] = *fields.Detail
	}
	if fields.Status != nil {
		updates["status"] = *fields.Status
	}
	if fields.BlockingReason != nil {
		updates["blocking_reason"] = *fields.BlockingReason
	}
	if fields.EstimateMin != nil {
		updates["estimate_minutes"] = *fields.EstimateMin
	}
	if fields.OrderIndex != nil {
		updates["order_index"] = *fields.OrderIndex
	}
	if fields.PlannedStart != nil {
		if *fields.PlannedStart == "" {
			updates["planned_start"] = nil
		} else {
			t, err := time.Parse(time.RFC3339, *fields.PlannedStart)
			if err != nil {
				return nil, err
			}
			updates["planned_start"] = t
		}
	}
	if fields.PlannedEnd != nil {
		if *fields.PlannedEnd == "" {
			updates["planned_end"] = nil
		} else {
			t, err := time.Parse(time.RFC3339, *fields.PlannedEnd)
			if err != nil {
				return nil, err
			}
			updates["planned_end"] = t
		}
	}

	return updates, nil
}
