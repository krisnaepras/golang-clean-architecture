package messaging

import (
	"golang-clean-architecture/internal/model"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type UserProducer struct {
	Producer[*model.UserRegisteredEvent]
}

func NewUserProducer(producer sarama.SyncProducer, log *logrus.Logger) *UserProducer {
	return &UserProducer{
		Producer: Producer[*model.UserRegisteredEvent]{
			Producer: producer,
			Topic:    "users",
			Log:      log,
		},
	}
}
