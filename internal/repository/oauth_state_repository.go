package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OauthStateRepository struct {
	Repository[entity.OauthState]
	Log *logrus.Logger
}

func NewOauthStateRepository(log *logrus.Logger) *OauthStateRepository {
	return &OauthStateRepository{Log: log}
}

func (r *OauthStateRepository) FindByState(db *gorm.DB, s *entity.OauthState, state string) error {
	return db.Where("state = ? AND expired_at > now()", state).First(s).Error
}
