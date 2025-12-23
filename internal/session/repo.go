package session

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

func (r *Repository) GetSession(ctx context.Context, sessionID uint64) (*Session, error) {
	var s Session
	err := r.db.WithContext(ctx).First(&s, sessionID).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) CreateMessage(ctx context.Context, m *Message) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) ListRecentMessages(ctx context.Context, sessionID uint64, limit int) ([]Message, error) {
	var messages []Message
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 反转顺序，使最早的消息在前面
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetTaskSessionOrCreate: 针对某个 user + task 找到一个 task session，没有就创建。
func (r *Repository) GetTaskSessionOrCreate(ctx context.Context, userID, taskID uint64) (*Session, error) {
	var sess Session
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND task_id = ? AND type = 'task'", userID, taskID).
		Limit(1).
		Find(&sess)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		return &sess, nil
	}

	sess = Session{
		UserID: userID,
		TaskID: &taskID,
		Type:   "task",
	}
	if err := r.db.WithContext(ctx).Create(&sess).Error; err != nil {
		return nil, err
	}
	return &sess, nil
}

// CreateSystemMessage 在指定 session 中插入 system 消息
func (r *Repository) CreateSystemMessage(ctx context.Context, sessionID uint64, agentName *string, content string) error {
	msg := Message{
		SessionID: sessionID,
		Role:      "system",
		AgentName: agentName,
		Content:   content,
	}
	return r.db.WithContext(ctx).Create(&msg).Error
}

// CreateSystemMessageForTask: 针对某个任务（以及用户）写一条系统消息。
// 用于依赖解锁、系统通知等场景。
func (r *Repository) CreateSystemMessageForTask(
	ctx context.Context,
	userID, taskID uint64,
	content string,
) error {
	sess, err := r.GetTaskSessionOrCreate(ctx, userID, taskID)
	if err != nil {
		return err
	}
	systemAgent := "system"
	return r.CreateSystemMessage(ctx, sess.ID, &systemAgent, content)
}

// GetGlobalSessionOrCreate: 获取或创建全局会话
func (r *Repository) GetGlobalSessionOrCreate(ctx context.Context, userID uint64) (*Session, error) {
	var sess Session
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND type = 'global'", userID).
		Limit(1).
		Find(&sess)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		return &sess, nil
	}

	sess = Session{
		UserID: userID,
		Type:   "global",
	}
	if err := r.db.WithContext(ctx).Create(&sess).Error; err != nil {
		return nil, err
	}
	return &sess, nil
}

// ClearMessages 清空指定 session 的所有消息
func (r *Repository) ClearMessages(ctx context.Context, sessionID uint64) error {
	return r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Delete(&Message{}).Error
}
