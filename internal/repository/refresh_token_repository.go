package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	Repository[entity.RefreshToken]
	Log *logrus.Logger
}

func NewRefreshTokenRepository(log *logrus.Logger) *RefreshTokenRepository {
	return &RefreshTokenRepository{Log: log}
}

func (r *RefreshTokenRepository) FindByHash(db *gorm.DB, t *entity.RefreshToken, hash string) error {
	return db.Where("token_hash = ? AND is_revoked = false", hash).First(t).Error
}

func (r *RefreshTokenRepository) RevokeAllByUser(db *gorm.DB, userID string) error {
	return db.Model(&entity.RefreshToken{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Update("is_revoked", true).Error
}
