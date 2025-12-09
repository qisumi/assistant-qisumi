package db

import (
	"database/sql"
	"fmt"

	"assistant-qisumi/internal/config"

	_ "github.com/go-sql-driver/mysql" // 确保已执行 go get github.com/go-sql-driver/mysql
)

// InitDB 初始化数据库连接
func InitDB(cfg config.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)

	// 执行数据库迁移
	if err := migrateDB(db); err != nil {
		return nil, err
	}

	return db, nil
}

// migrateDB 执行数据库迁移，创建所有必要的表
func migrateDB(db *sql.DB) error {
	// 创建 users 表
	usersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
	  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	  email VARCHAR(255) NOT NULL UNIQUE,
	  display_name VARCHAR(255),
	  password_hash VARCHAR(255) NOT NULL,
	  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	if _, err := db.Exec(usersTableSQL); err != nil {
		return fmt.Errorf("create users table: %v", err)
	}

	// 创建 user_llm_settings 表
	llmSettingsTableSQL := `
	CREATE TABLE IF NOT EXISTS user_llm_settings (
	  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	  user_id BIGINT UNSIGNED NOT NULL,
	  base_url VARCHAR(512) NOT NULL,
	  api_key_enc TEXT NOT NULL,
	  model VARCHAR(255) NOT NULL,
	  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	  CONSTRAINT fk_llm_user FOREIGN KEY (user_id) REFERENCES users(id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	if _, err := db.Exec(llmSettingsTableSQL); err != nil {
		return fmt.Errorf("create user_llm_settings table: %v", err)
	}

	// 创建 tasks 表
	tasksTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
	  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	  user_id BIGINT UNSIGNED NOT NULL,
	  title VARCHAR(255) NOT NULL,
	  description TEXT,
	  status ENUM('todo','in_progress','done','cancelled') NOT NULL DEFAULT 'todo',
	  priority ENUM('low','medium','high') DEFAULT 'medium',
	  due_at DATETIME NULL,
	  created_from TEXT,
	  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	  CONSTRAINT fk_tasks_user FOREIGN KEY (user_id) REFERENCES users(id),
	  INDEX idx_tasks_user (user_id),
	  INDEX idx_tasks_user_status (user_id, status),
	  INDEX idx_tasks_user_due (user_id, due_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	if _, err := db.Exec(tasksTableSQL); err != nil {
		return fmt.Errorf("create tasks table: %v", err)
	}

	// 创建 task_steps 表
	taskStepsTableSQL := `
	CREATE TABLE IF NOT EXISTS task_steps (
	  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	  task_id BIGINT UNSIGNED NOT NULL,
	  order_index INT NOT NULL DEFAULT 0,
	  title VARCHAR(255) NOT NULL,
	  detail TEXT,
	  status ENUM('locked','todo','in_progress','done','blocked') NOT NULL DEFAULT 'todo',
	  blocking_reason TEXT,
	  estimate_minutes INT NULL,
	  planned_start DATETIME NULL,
	  planned_end DATETIME NULL,
	  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	  CONSTRAINT fk_steps_task FOREIGN KEY (task_id) REFERENCES tasks(id),
	  INDEX idx_steps_task (task_id),
	  INDEX idx_steps_task_order (task_id, order_index),
	  INDEX idx_steps_schedule (task_id, planned_start, planned_end)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	if _, err := db.Exec(taskStepsTableSQL); err != nil {
		return fmt.Errorf("create task_steps table: %v", err)
	}

	// 创建 task_dependencies 表
	taskDependenciesTableSQL := `
	CREATE TABLE IF NOT EXISTS task_dependencies (
	  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	  predecessor_task_id BIGINT UNSIGNED NOT NULL,
	  predecessor_step_id BIGINT UNSIGNED NULL,
	  successor_task_id BIGINT UNSIGNED NOT NULL,
	  successor_step_id BIGINT UNSIGNED NULL,
	  condition ENUM('task_done','step_done') NOT NULL,
	  action ENUM('unlock_step','set_task_todo','notify_only') NOT NULL DEFAULT 'unlock_step',
	  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  CONSTRAINT fk_dep_pre_task FOREIGN KEY (predecessor_task_id) REFERENCES tasks(id),
	  CONSTRAINT fk_dep_pre_step FOREIGN KEY (predecessor_step_id) REFERENCES task_steps(id),
	  CONSTRAINT fk_dep_suc_task FOREIGN KEY (successor_task_id) REFERENCES tasks(id),
	  CONSTRAINT fk_dep_suc_step FOREIGN KEY (successor_step_id) REFERENCES task_steps(id),
	  INDEX idx_dep_pre (predecessor_task_id, predecessor_step_id),
	  INDEX idx_dep_suc (successor_task_id, successor_step_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	if _, err := db.Exec(taskDependenciesTableSQL); err != nil {
		return fmt.Errorf("create task_dependencies table: %v", err)
	}

	// 创建 sessions 表
	sessionsTableSQL := `
	CREATE TABLE IF NOT EXISTS sessions (
	  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	  user_id BIGINT UNSIGNED NOT NULL,
	  task_id BIGINT UNSIGNED NULL,
	  type ENUM('task','global') NOT NULL DEFAULT 'task',
	  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id),
	  CONSTRAINT fk_sessions_task FOREIGN KEY (task_id) REFERENCES tasks(id),
	  INDEX idx_sessions_user (user_id),
	  INDEX idx_sessions_task (task_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	if _, err := db.Exec(sessionsTableSQL); err != nil {
		return fmt.Errorf("create sessions table: %v", err)
	}

	// 创建 messages 表
	messagesTableSQL := `
	CREATE TABLE IF NOT EXISTS messages (
	  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
	  session_id BIGINT UNSIGNED NOT NULL,
	  role ENUM('user','assistant','system') NOT NULL,
	  agent_name VARCHAR(64) NULL,
	  content TEXT NOT NULL,
	  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	  CONSTRAINT fk_messages_session FOREIGN KEY (session_id) REFERENCES sessions(id),
	  INDEX idx_messages_session (session_id, created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	if _, err := db.Exec(messagesTableSQL); err != nil {
		return fmt.Errorf("create messages table: %v", err)
	}

	return nil
}
