package service

import (
	"errors"
	"time"

	"github.com/zealot/managing-up/apps/api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
)

type UserRepository interface {
	GetUserByUsername(username string) (models.User, bool)
	GetUserByID(id string) (models.User, bool)
	CreateUser(user models.User) error
}

type AuthService struct {
	repo UserRepository
}

func NewAuthService(repo UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

type LoginRequest struct {
	Username string
	Password string
}

type AuthResult struct {
	User  models.User
	Token string
}

func (s *AuthService) Login(req LoginRequest) (AuthResult, error) {
	user, ok := s.repo.GetUserByUsername(req.Username)
	if !ok {
		return AuthResult{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	return AuthResult{User: user}, nil
}

func (s *AuthService) GetUserByID(id string) (models.User, error) {
	user, ok := s.repo.GetUserByID(id)
	if !ok {
		return models.User{}, ErrUserNotFound
	}
	return user, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *AuthService) EnsureAdminUser(username, password string) error {
	_, ok := s.repo.GetUserByUsername(username)
	if ok {
		return nil
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	user := models.User{
		ID:           "user_admin",
		Username:     username,
		PasswordHash: hash,
		Role:         "admin",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	return s.repo.CreateUser(user)
}
