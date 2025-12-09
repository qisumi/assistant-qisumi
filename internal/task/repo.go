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

// UpdateTask 更新任务字段
func (r *Repository) UpdateTask(ctx context.Context, tx *sql.Tx, taskID uint64, fields map[string]interface{}) error {
	// 构建UPDATE语句
	query := `UPDATE tasks SET `
	var args []interface{}
	var setClauses []string

	// 添加更新字段
	for key, value := range fields {
		switch key {
		case "title", "description", "status", "priority":
			setClauses = append(setClauses, key+" = ?")
			args = append(args, value)
		case "due_at":
			setClauses = append(setClauses, key+" = ?")
			args = append(args, value)
		}
	}

	if len(setClauses) == 0 {
		return nil // 没有需要更新的字段
	}

	query += " " + joinStrings(setClauses, ", ") + " WHERE id = ?"
	args = append(args, taskID)

	// 执行更新
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = r.db.ExecContext(ctx, query, args...)
	}

	return err
}

// UpdateStep 更新步骤字段
func (r *Repository) UpdateStep(ctx context.Context, tx *sql.Tx, taskID, stepID uint64, fields map[string]interface{}) error {
	// 构建UPDATE语句
	query := `UPDATE task_steps SET `
	var args []interface{}
	var setClauses []string

	// 添加更新字段
	for key, value := range fields {
		switch key {
		case "title", "detail", "status", "blocking_reason":
			setClauses = append(setClauses, key+" = ?")
			args = append(args, value)
		case "estimate_minutes":
			setClauses = append(setClauses, key+" = ?")
			args = append(args, value)
		case "order_index":
			setClauses = append(setClauses, key+" = ?")
			args = append(args, value)
		case "planned_start", "planned_end":
			setClauses = append(setClauses, key+" = ?")
			args = append(args, value)
		}
	}

	if len(setClauses) == 0 {
		return nil // 没有需要更新的字段
	}

	query += " " + joinStrings(setClauses, ", ") + " WHERE id = ? AND task_id = ?"
	args = append(args, stepID, taskID)

	// 执行更新
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = r.db.ExecContext(ctx, query, args...)
	}

	return err
}

// joinStrings 连接字符串切片
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// AddStep 添加新步骤
func (r *Repository) AddStep(ctx context.Context, tx *sql.Tx, step *Step) error {
	query := `INSERT INTO task_steps (task_id, order_index, title, detail, status, blocking_reason, estimate_minutes, planned_start, planned_end) 
	         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var err error
	var result sql.Result
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, step.TaskID, step.OrderIndex, step.Title, step.Detail, step.Status, step.BlockingReason, step.EstimateMin, step.PlannedStart, step.PlannedEnd)
	} else {
		result, err = r.db.ExecContext(ctx, query, step.TaskID, step.OrderIndex, step.Title, step.Detail, step.Status, step.BlockingReason, step.EstimateMin, step.PlannedStart, step.PlannedEnd)
	}

	if err != nil {
		return err
	}

	stepID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	step.ID = uint64(stepID)

	return nil
}

// AddSteps 添加多个新步骤
func (r *Repository) AddSteps(ctx context.Context, tx *sql.Tx, steps []Step) error {
	for _, step := range steps {
		if err := r.AddStep(ctx, tx, &step); err != nil {
			return err
		}
	}
	return nil
}

// AddDependency 添加任务依赖
func (r *Repository) AddDependency(ctx context.Context, tx *sql.Tx, dep *Dependency) error {
	query := `INSERT INTO task_dependencies (predecessor_task_id, predecessor_step_id, successor_task_id, successor_step_id, dependency_condition, action) 
	         VALUES (?, ?, ?, ?, ?, ?)`

	var err error
	var result sql.Result
	if tx != nil {
		result, err = tx.ExecContext(ctx, query, dep.PredecessorTaskID, dep.PredecessorStepID, dep.SuccessorTaskID, dep.SuccessorStepID, dep.DependencyCondition, dep.Action)
	} else {
		result, err = r.db.ExecContext(ctx, query, dep.PredecessorTaskID, dep.PredecessorStepID, dep.SuccessorTaskID, dep.SuccessorStepID, dep.DependencyCondition, dep.Action)
	}

	if err != nil {
		return err
	}

	depID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	dep.ID = uint64(depID)

	return nil
}

// AddDependencies 添加多个任务依赖
func (r *Repository) AddDependencies(ctx context.Context, tx *sql.Tx, dependencies []Dependency) error {
	for _, dep := range dependencies {
		if err := r.AddDependency(ctx, tx, &dep); err != nil {
			return err
		}
	}
	return nil
}

// MarkTasksFocusToday 标记任务为今日重点（示例实现，实际需要根据业务需求调整）
func (r *Repository) MarkTasksFocusToday(ctx context.Context, tx *sql.Tx, userID uint64, taskIDs []uint64) error {
	// TODO: 实现标记今日重点任务的逻辑
	// 这需要根据实际业务需求来实现，可能需要一个专门的表来存储今日重点任务
	return nil
}
