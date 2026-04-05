package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/KazikovAP/backend_course_start/hw2/internal/domain"
)

type AuthService struct {
	users     domain.UserRepository
	jwtSecret []byte
}

func NewAuthService(users domain.UserRepository, jwtSecret []byte) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", fmt.Errorf("invalid input")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.users.Create(domain.User{Username: username, Password: hash}); err != nil {
		return "", fmt.Errorf("user exists")
	}

	return s.generateToken(username)
}

func (s *AuthService) Login(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", fmt.Errorf("invalid input")
	}

	user, ok := s.users.FindByUsername(username)
	if !ok {
		return "", fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	return s.generateToken(username)
}

func (s *AuthService) ValidateToken(tokenStr string) bool {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	return err == nil && token.Valid
}

func (s *AuthService) generateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(s.jwtSecret)
}
