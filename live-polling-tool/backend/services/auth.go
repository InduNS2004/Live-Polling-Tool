package services

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"live-polling-tool/backend/models"
	"live-polling-tool/backend/repositories"
	"strings"
	"time"
)

type AuthService struct {
	Repo   *repositories.Repo
	Secret string
}

func (s *AuthService) Register(ctx context.Context, email, password string) (models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if len(email) < 5 || len(password) < 8 {
		return models.User{}, errors.New("invalid credentials")
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if e != nil {
		return models.User{}, e
	}
	u := models.User{ID: uuid.NewString(), Email: email, PasswordHash: string(hash), CreatedAt: time.Now()}
	e = s.Repo.CreateUser(ctx, u)
	return u, e
}
func (s *AuthService) Login(ctx context.Context, email, password string) (models.User, string, error) {
	u, e := s.Repo.FindUser(ctx, strings.ToLower(strings.TrimSpace(email)))
	if e != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return models.User{}, "", errors.New("invalid email or password")
	}
	t, e := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": u.ID, "exp": time.Now().Add(24 * time.Hour).Unix()}).SignedString([]byte(s.Secret))
	return u, t, e
}
