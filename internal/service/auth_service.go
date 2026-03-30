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
	Login(req model.LoginRequest) (model.LoginResponse, error)
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) AuthService {
	return &authService{
		repo: repo,
	}
}

func (s *authService) Register(req model.RegisterRequest) (model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	user := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(&user); err != nil {
		return model.User{}, errors.New("failed to create user, email might already exist")
	}

	return user, nil
}

func (s *authService) Login(req model.LoginRequest) (model.LoginResponse, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return model.LoginResponse{}, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return model.LoginResponse{}, errors.New("invalid email or password")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "supersecretkey"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return model.LoginResponse{}, errors.New("failed to generate token")
	}

	return model.LoginResponse{
		Token: tokenString,
		User:  user,
	}, nil
}
