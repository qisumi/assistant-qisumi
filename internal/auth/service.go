package auth

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	jwt *JWTManager
}

func NewService(db *gorm.DB, jwt *JWTManager) *Service {
	return &Service{db: db, jwt: jwt}
}

func (s *Service) Register(ctx context.Context, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := &User{
		Email:        email,
		PasswordHash: string(hash),
	}
	return s.db.WithContext(ctx).Create(user).Error
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	var u User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return s.jwt.GenerateToken(u.ID)
}
