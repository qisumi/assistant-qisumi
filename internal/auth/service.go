package auth

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uint64
	Email        string
	DisplayName  string
	PasswordHash string
}

type Service struct {
	db  *sql.DB
	jwt *JWTManager
}

func NewService(db *sql.DB, jwt *JWTManager) *Service {
	return &Service{db: db, jwt: jwt}
}

func (s *Service) Register(ctx context.Context, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		email, string(hash),
	)
	return err
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	var u User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, password_hash FROM users WHERE email = ?`, email).
		Scan(&u.ID, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return s.jwt.GenerateToken(u.ID)
}
