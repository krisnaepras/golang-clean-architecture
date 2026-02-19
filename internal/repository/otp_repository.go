package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OtpRepository struct {
	Repository[entity.Otp]
	Log *logrus.Logger
}

func NewOtpRepository(log *logrus.Logger) *OtpRepository {
	return &OtpRepository{Log: log}
}

// FindActiveByDestinationAndPurpose mencari OTP yang belum diverifikasi dan belum expired.
func (r *OtpRepository) FindActiveByDestinationAndPurpose(db *gorm.DB, otp *entity.Otp, destination, purpose string) error {
	return db.Where(
		"destination = ? AND purpose = ? AND verified_at IS NULL AND expired_at > now()",
		destination, purpose,
	).Order("created_at DESC").First(otp).Error
}
