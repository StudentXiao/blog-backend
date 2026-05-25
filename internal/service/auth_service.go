package service

import (
	"blog-backend/internal/models"
	"blog-backend/internal/repositorty"
	"blog-backend/pkg/jwt"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repositorty.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: &repositorty.UserRepository{},
	}
}

func (s *AuthService) Register(username, email, password string) (*models.User, string, error) {
	// Check if user exists
	existingUser, _ := s.userRepo.FindByEmail(email)
	if existingUser != nil {
		return nil, "", errors.New("email already registered")
	}

	existingUser, _ = s.userRepo.FindByUsername(username)
	if existingUser != nil {
		return nil, "", errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", err
	}

	// Generate token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}
