package task

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetTaskWithSteps(ctx context.Context, userID, taskID uint64) (*Task, error) {
	// 查询task基本信息
	var t Task
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, title, description, status, priority, due_at, created_at, updated_at 
		 FROM tasks WHERE id = ? AND user_id = ?`,
		taskID, userID).
		Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueAt, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// 查询task_steps
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, task_id, order_index, title, detail, status, blocking_reason, estimate_minutes, planned_start, planned_end, created_at, updated_at 
		 FROM task_steps WHERE task_id = ? ORDER BY order_index ASC`,
		taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []Step
	for rows.Next() {
		var s Step
		if err := rows.Scan(&s.ID, &s.TaskID, &s.OrderIndex, &s.Title, &s.Detail, &s.Status, &s.BlockingReason, &s.EstimateMin, &s.PlannedStart, &s.PlannedEnd, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	t.Steps = steps

	return &t, nil
}

func (r *Repository) InsertTaskWithSteps(ctx context.Context, t *Task) error {
	// 开始事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 插入task
	result, err := tx.ExecContext(ctx,
		`INSERT INTO tasks (user_id, title, description, status, priority, due_at) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.UserID, t.Title, t.Description, t.Status, t.Priority, t.DueAt)
	if err != nil {
		return err
	}

	// 获取task_id
	taskID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = uint64(taskID)

	// 插入steps
	for i, step := range t.Steps {
		step.TaskID = t.ID
		step.OrderIndex = i
		result, err := tx.ExecContext(ctx,
			`INSERT INTO task_steps (task_id, order_index, title, detail, status, blocking_reason, estimate_minutes, planned_start, planned_end) 
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			step.TaskID, step.OrderIndex, step.Title, step.Detail, step.Status, step.BlockingReason, step.EstimateMin, step.PlannedStart, step.PlannedEnd)
		if err != nil {
			return err
		}
		stepID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		t.Steps[i].ID = uint64(stepID)
	}

	return nil
}

func (r *Repository) ListTasks(ctx context.Context, userID uint64) ([]Task, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, title, description, status, priority, due_at, created_at, updated_at 
		 FROM tasks WHERE user_id = ? ORDER BY created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}