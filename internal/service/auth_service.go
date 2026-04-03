package service

import (
	"errors"
	"os"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req model.RegisterRequest) (model.User, error)
	Login(req model.LoginRequest) (string, model.UserResponse, error)
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{
		repo: repo,
	}
}

// ======================
// REGISTER
// ======================
func (s *authService) Register(req model.RegisterRequest) (model.User, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return model.User{}, errors.New("name, email, password required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	user := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     model.RoleEmployee,
		TenantID: req.TenantID,
	}

	if err := s.repo.Create(&user); err != nil {
		return model.User{}, errors.New("failed to create user, email might already exist")
	}

	return user, nil
}

// ======================
// LOGIN (UPDATED)
// ======================
func (s *authService) Login(req model.LoginRequest) (string, model.UserResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return "", model.UserResponse{}, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", model.UserResponse{}, errors.New("invalid email or password")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", model.UserResponse{}, errors.New("JWT secret not configured")
	}

	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"tenant_id": user.TenantID,
		"role":      user.Role,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", model.UserResponse{}, errors.New("failed to generate token")
	}

	userResponse := model.UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
		TenantID: user.TenantID,
	}

	return tokenString, userResponse, nil
}
