package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuthProviderRepository struct {
	Repository[entity.UserAuthProvider]
	Log *logrus.Logger
}

func NewAuthProviderRepository(log *logrus.Logger) *AuthProviderRepository {
	return &AuthProviderRepository{Log: log}
}

func (r *AuthProviderRepository) FindByProvider(db *gorm.DB, p *entity.UserAuthProvider, provider, providerUserID string) error {
	return db.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(p).Error
}

func (r *AuthProviderRepository) FindByUserAndProvider(db *gorm.DB, p *entity.UserAuthProvider, userID, provider string) error {
	return db.Where("user_id = ? AND provider = ?", userID, provider).First(p).Error
}
