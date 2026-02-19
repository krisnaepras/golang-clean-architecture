package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserRepository struct {
	Repository[entity.User]
	Log *logrus.Logger
}

func NewUserRepository(log *logrus.Logger) *UserRepository {
	return &UserRepository{Log: log}
}

func (r *UserRepository) FindByEmail(db *gorm.DB, user *entity.User, email string) error {
	return db.Where("email = ?", email).First(user).Error
}

// FindWithRoles mengambil user beserta daftar kode role-nya.
func (r *UserRepository) FindWithRoles(db *gorm.DB, userID string) (*entity.User, []string, error) {
	var user entity.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, nil, err
	}

	var roles []struct {
		Code string
	}
	err := db.Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Scan(&roles).Error
	if err != nil {
		return nil, nil, err
	}

	codes := make([]string, len(roles))
	for i, r := range roles {
		codes[i] = r.Code
	}
	return &user, codes, nil
}
