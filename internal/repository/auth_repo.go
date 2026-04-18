package repository

import (
	"time"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type AuthRepository interface {
	Create(user *model.User) error
	FindByEmail(email string) (model.User, error)
	FindByID(id uint) (model.User, error)
	SaveToken(token *model.Token) error
	RevokeToken(token string) error
	IsTokenRevoked(token string) (bool, error)
	CountByTenantID(tenantID uint) (int64, error)
	FindTokensByUserID(userID uint) ([]model.Token, error)
	UpdatePassword(userID uint, newPassword string) error
	SavePasswordReset(reset *model.PasswordReset) error
	FindPasswordResetByToken(token string) (model.PasswordReset, error)
	MarkPasswordResetUsed(token string) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{
		db: db,
	}
}

func (r *authRepository) FindTokensByUserID(userID uint) ([]model.Token, error) {
	var tokens []model.Token
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error
	return tokens, err
}

func (r *authRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) CountByTenantID(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

func (r *authRepository) FindByEmail(email string) (model.User, error) {
	var user model.User
	err := r.db.Preload("Role.Permissions").Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *authRepository) FindByID(id uint) (model.User, error) {
	var user model.User
	err := r.db.Preload("Tenant").Preload("Role.Permissions").First(&user, id).Error
	return user, err
}

func (r *authRepository) SaveToken(token *model.Token) error {
	return r.db.Create(token).Error
}

func (r *authRepository) RevokeToken(token string) error {
	return r.db.Model(&model.Token{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
		}).Error
}

func (r *authRepository) IsTokenRevoked(token string) (bool, error) {
	var t model.Token
	err := r.db.Where("token = ?", token).First(&t).Error
	if err != nil {
		return false, err
	}
	return t.IsRevoked, nil
}

func (r *authRepository) UpdatePassword(userID uint, newPassword string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password":             newPassword,
		"must_change_password": false,
	}).Error
}

func (r *authRepository) SavePasswordReset(reset *model.PasswordReset) error {
	return r.db.Create(reset).Error
}

func (r *authRepository) FindPasswordResetByToken(token string) (model.PasswordReset, error) {
	var reset model.PasswordReset
	err := r.db.Where("token = ? AND is_used = ? AND expires_at > ?", token, false, time.Now()).First(&reset).Error
	return reset, err
}

func (r *authRepository) MarkPasswordResetUsed(token string) error {
	return r.db.Model(&model.PasswordReset{}).Where("token = ?", token).Update("is_used", true).Error
}
