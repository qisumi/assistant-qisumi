package dependency

import (
	"context"

	"assistant-qisumi/internal/session"
	"assistant-qisumi/internal/task"

	"gorm.io/gorm"
)

type Service struct {
	db          *gorm.DB
	taskRepo    *task.Repository
	sessionRepo *session.Repository
}

func NewService(db *gorm.DB, taskRepo *task.Repository, sessionRepo *session.Repository) *Service {
	return &Service{
		db:          db,
		taskRepo:    taskRepo,
		sessionRepo: sessionRepo,
	}
}

// OnTaskOrStepDone predecessorStepID 为 nil 表示「整个任务完成」的触发
func (s *Service) OnTaskOrStepDone(
	ctx context.Context,
	predecessorTaskID uint64,
	predecessorStepID *uint64,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var deps []task.TaskDependency

		q := tx.Where("predecessor_task_id = ?", predecessorTaskID)
		if predecessorStepID != nil {
			q = q.Where("predecessor_step_id = ?", *predecessorStepID).
				Where("dependency_condition = ?", "step_done")
		} else {
			// 任务完成触发：predecessor_step_id IS NULL AND condition = 'task_done'
			q = q.Where("predecessor_step_id IS NULL").
				Where("dependency_condition = ?", "task_done")
		}

		if err := q.Find(&deps).Error; err != nil {
			return err
		}
		if len(deps) == 0 {
			return nil
		}

		// 针对每条依赖执行对应动作
		for _, d := range deps {
			switch d.Action {
			case "unlock_step":
				if d.SuccessorStepID == nil {
					// 没有具体步骤就跳过
					continue
				}
				// 仅在 locked 状态下改为 todo，避免覆盖用户手动状态
				if err := tx.Model(&task.TaskStep{}).
					Where("id = ? AND task_id = ? AND status = ?", *d.SuccessorStepID, d.SuccessorTaskID, "locked").
					Update("status", "todo").Error; err != nil {
					return err
				}

			case "set_task_todo":
				// 如果任务不是 done，就把状态设置为 todo
				if err := tx.Model(&task.Task{}).
					Where("id = ? AND status != ?", d.SuccessorTaskID, "done").
					Update("status", "todo").Error; err != nil {
					return err
				}

			case "notify_only":
				// TODO: 写入系统消息
			}
		}

		return nil
	})
}
