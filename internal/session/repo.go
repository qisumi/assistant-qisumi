package session

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

func (r *Repository) GetSession(ctx context.Context, sessionID uint64) (*Session, error) {
	var s Session
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, task_id, type, created_at 
		 FROM sessions WHERE id = ?`,
		sessionID).
		Scan(&s.ID, &s.UserID, &s.TaskID, &s.Type, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) CreateMessage(ctx context.Context, m *Message) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO messages (session_id, role, agent_name, content) 
		 VALUES (?, ?, ?, ?)`,
		m.SessionID, m.Role, m.AgentName, m.Content)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	m.ID = uint64(id)
	return nil
}

func (r *Repository) ListRecentMessages(ctx context.Context, sessionID uint64, limit int) ([]Message, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, session_id, role, agent_name, content, created_at 
		 FROM messages WHERE session_id = ? 
		 ORDER BY created_at DESC 
		 LIMIT ?`,
		sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.AgentName, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	// 反转顺序，使最早的消息在前面
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}