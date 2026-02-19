package repository

import (
	"golang-clean-architecture/internal/entity"

	"github.com/sirupsen/logrus"
)

type OtpDeliveryRepository struct {
	Repository[entity.OtpDelivery]
	Log *logrus.Logger
}

func NewOtpDeliveryRepository(log *logrus.Logger) *OtpDeliveryRepository {
	return &OtpDeliveryRepository{Log: log}
}
