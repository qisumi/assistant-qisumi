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
		Where("user_id = ? AND status != ?", userID, "done").
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// ListCompletedTasks 获取用户已完成的任务列表
func (r *Repository) ListCompletedTasks(ctx context.Context, userID uint64) ([]Task, error) {
	var tasks []Task
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "done").
		Order("updated_at DESC").
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
	// 如果需要更新 status，需要先查询当前状态来决定是否自动设置 completedAt
	if fields.Status != nil && fields.CompletedAt == nil {
		var currentStep TaskStep
		subQuery := r.db.
			Select("id").
			Table("tasks").
			Where("id = ? AND user_id = ?", taskID, userID)

		err := r.db.WithContext(ctx).
			Where("id = ? AND task_id IN (?)", stepID, subQuery).
			First(&currentStep).Error
		if err != nil {
			return err
		}

		// 自动处理 completedAt：当状态变为 done 时设置为当前时间，否则清除
		now := r.db.NowFunc()
		if *fields.Status == "done" && currentStep.Status != "done" {
			// 状态从非 done 变为 done，设置 completedAt
			completedAtStr := now.Format(time.RFC3339)
			fields.CompletedAt = &completedAtStr
		} else if *fields.Status != "done" && currentStep.Status == "done" {
			// 状态从 done 变为非 done，清除 completedAt
			emptyStr := ""
			fields.CompletedAt = &emptyStr
		}
	}

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

// DeleteStep 删除步骤
func (r *Repository) DeleteStep(ctx context.Context, userID, taskID, stepID uint64) error {
	// 使用子查询验证任务属于该用户
	subQuery := r.db.
		Select("id").
		Table("tasks").
		Where("id = ? AND user_id = ?", taskID, userID)

	result := r.db.WithContext(ctx).
		Where("id = ? AND task_id IN (?)", stepID, subQuery).
		Delete(&TaskStep{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

// GetTaskDependencies 获取与指定任务相关的所有依赖关系
// 返回该任务作为前置或后置的所有依赖记录
func (r *Repository) GetTaskDependencies(ctx context.Context, userID, taskID uint64) ([]TaskDependency, error) {
	var deps []TaskDependency
	err := r.db.WithContext(ctx).
		Where("(predecessor_task_id = ? OR successor_task_id = ?) AND "+
			"(predecessor_task_id IN (SELECT id FROM tasks WHERE user_id = ?) OR "+
			"successor_task_id IN (SELECT id FROM tasks WHERE user_id = ?))",
			taskID, taskID, userID, userID).
		Find(&deps).Error
	return deps, err
}

// GetAllUserDependencies 获取用户所有任务的依赖关系
// 用于全局视图或Executor判断依赖关系
func (r *Repository) GetAllUserDependencies(ctx context.Context, userID uint64) ([]TaskDependency, error) {
	var deps []TaskDependency
	err := r.db.WithContext(ctx).
		Where("predecessor_task_id IN (SELECT id FROM tasks WHERE user_id = ?) OR "+
			"successor_task_id IN (SELECT id FROM tasks WHERE user_id = ?)",
			userID, userID).
		Find(&deps).Error
	return deps, err
}

// MarkTasksFocusToday 标记任务为今日重点
func (r *Repository) MarkTasksFocusToday(ctx context.Context, userID uint64, taskIDs []uint64) error {
	if len(taskIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&Task{}).
		Where("id IN ? AND user_id = ?", taskIDs, userID).
		Update("is_focus_today", true).Error
}

// 把 UpdateTaskFields 转成 GORM Updates 使用s的 map
func buildTaskUpdateMap(fields UpdateTaskFields) (map[string]any, error) {
	updates := make(map[string]any)

	setIfNotNil(updates, "title", fields.Title)
	setIfNotNil(updates, "description", fields.Description)
	setIfNotNil(updates, "status", fields.Status)
	setIfNotNil(updates, "priority", fields.Priority)
	setIfNotNil(updates, "is_focus_today", fields.IsFocusToday)
	
	if err := setFlexibleTimeField(updates, "due_at", fields.DueAt); err != nil {
		return nil, err
	}
	if err := setRFC3339TimeField(updates, "completed_at", fields.CompletedAt); err != nil {
		return nil, err
	}

	return updates, nil
}

func buildStepUpdateMap(fields UpdateStepFields) (map[string]any, error) {
	updates := make(map[string]any)

	setIfNotNil(updates, "title", fields.Title)
	setIfNotNil(updates, "detail", fields.Detail)
	setIfNotNil(updates, "status", fields.Status)
	setIfNotNil(updates, "blocking_reason", fields.BlockingReason)
	setIfNotNil(updates, "estimate_minutes", fields.EstimateMin)
	setIfNotNil(updates, "order_index", fields.OrderIndex)
	
	if err := setFlexibleTimeField(updates, "planned_start", fields.PlannedStart); err != nil {
		return nil, err
	}
	if err := setFlexibleTimeField(updates, "planned_end", fields.PlannedEnd); err != nil {
		return nil, err
	}
	if err := setRFC3339TimeField(updates, "completed_at", fields.CompletedAt); err != nil {
		return nil, err
	}

	return updates, nil
}

// setIfNotNil 如果值非 nil，则设置到 map 中
func setIfNotNil[T any](m map[string]any, key string, value *T) {
	if value != nil {
		m[key] = *value
	}
}

// setFlexibleTimeField 设置 FlexibleTime 字段
func setFlexibleTimeField(m map[string]any, key string, value *string) error {
	if value == nil {
		return nil
	}
	if *value == "" {
		m[key] = nil
		return nil
	}
	ft, err := ParseFlexibleTime(*value)
	if err != nil {
		return err
	}
	m[key] = ft.ToTime()
	return nil
}

// setRFC3339TimeField 设置 RFC3339 时间字段
func setRFC3339TimeField(m map[string]any, key string, value *string) error {
	if value == nil {
		return nil
	}
	if *value == "" {
		m[key] = nil
		return nil
	}
	t, err := ParseRFC3339(*value)
	if err != nil {
		return err
	}
	m[key] = t
	return nil
}

// DeleteTask 删除任务及其关联数据
func (r *Repository) DeleteTask(ctx context.Context, userID, taskID uint64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 删除任务步骤
		if err := tx.Where("task_id = ?", taskID).Delete(&TaskStep{}).Error; err != nil {
			return err
		}

		// 2. 删除任务依赖（作为前置或后置任务）
		if err := tx.Where("predecessor_task_id = ? OR successor_task_id = ?", taskID, taskID).Delete(&TaskDependency{}).Error; err != nil {
			return err
		}

		// 3. 删除关联的会话
		var sessionIDs []uint64
		if err := tx.Table("sessions").Where("task_id = ? AND user_id = ?", taskID, userID).Pluck("id", &sessionIDs).Error; err != nil {
			return err
		}

		// 4. 删除会话的消息
		if len(sessionIDs) > 0 {
			if err := tx.Table("messages").Where("session_id IN ?", sessionIDs).Delete(nil).Error; err != nil {
				return err
			}
		}

		// 5. 删除会话
		if err := tx.Table("sessions").Where("task_id = ? AND user_id = ?", taskID, userID).Delete(nil).Error; err != nil {
			return err
		}

		// 6. 删除任务本身（带 user_id 验证）
		result := tx.Where("id = ? AND user_id = ?", taskID, userID).Delete(&Task{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}
