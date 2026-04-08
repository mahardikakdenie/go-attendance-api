package service

import (
	"errors"
	"fmt"
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
	Logout(token string) error
	GetMe(token string) (model.UserResponse, error)
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
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return model.User{}, errors.New("name, email, password required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}

	// Generate EmployeeID: FT-001, FT-002, dst.
	count, err := s.repo.CountByTenantID(req.TenantID)
	if err != nil {
		return model.User{}, err
	}
	employeeID := fmt.Sprintf("FT-%03d", count+1)

	user := model.User{
		Name:        req.Name,
		Email:       req.Email,
		Password:    string(hashedPassword),
		RoleID:      req.RoleID,
		TenantID:    req.TenantID,
		EmployeeID:  employeeID,
		Department:  req.Department,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
	}

	if err := s.repo.Create(&user); err != nil {
		return model.User{}, errors.New("failed to create user, email might already exist")
	}

	return user, nil
}

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

	exp := time.Now().Add(24 * time.Hour)

	var roleName string
	if user.Role != nil {
		roleName = user.Role.Name
	}

	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"tenant_id": user.TenantID,
		"role":      roleName,
		"exp":       exp.Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", model.UserResponse{}, errors.New("failed to generate token")
	}

	err = s.repo.SaveToken(&model.Token{
		UserID:    user.ID,
		Token:     tokenString,
		IsRevoked: false,
	})
	if err != nil {
		return "", model.UserResponse{}, errors.New("failed to store token")
	}

	userResponse := model.UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		TenantID: user.TenantID,
	}

	if user.Role != nil {
		userResponse.Role = &model.RoleResponse{
			ID:   user.Role.ID,
			Name: user.Role.Name,
		}
	}

	return tokenString, userResponse, nil
}

func (s *authService) Logout(token string) error {
	if token == "" {
		return nil
	}
	return s.repo.RevokeToken(token)
}

func (s *authService) GetMe(token string) (model.UserResponse, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return model.UserResponse{}, errors.New("JWT secret not configured")
	}

	isRevoked, err := s.repo.IsTokenRevoked(token)
	if err != nil {
		return model.UserResponse{}, errors.New("invalid token")
	}
	if isRevoked {
		return model.UserResponse{}, errors.New("token revoked")
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !parsedToken.Valid {
		return model.UserResponse{}, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return model.UserResponse{}, errors.New("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return model.UserResponse{}, errors.New("invalid token payload")
	}

	userID := uint(userIDFloat)

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return model.UserResponse{}, errors.New("user not found")
	}

	var tenantResponse *model.TenantResponse
	if user.Tenant != nil {
		tenantResponse = &model.TenantResponse{
			ID:   user.Tenant.ID,
			Name: user.Tenant.Name,
		}
	}

	userResponse := model.UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		TenantID: user.TenantID,
		Tenant:   tenantResponse,
	}

	if user.Role != nil {
		userResponse.Role = &model.RoleResponse{
			ID:   user.Role.ID,
			Name: user.Role.Name,
		}
	}

	return userResponse, nil
}
