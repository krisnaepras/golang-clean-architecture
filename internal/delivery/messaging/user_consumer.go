package messaging

import (
	"encoding/json"
	"golang-clean-architecture/internal/model"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type UserConsumer struct {
	Log *logrus.Logger
}

func NewUserConsumer(log *logrus.Logger) *UserConsumer {
	return &UserConsumer{
		Log: log,
	}
}

func (c UserConsumer) Consume(message *sarama.ConsumerMessage) error {
	event := new(model.UserRegisteredEvent)
	if err := json.Unmarshal(message.Value, event); err != nil {
		c.Log.WithError(err).Error("error unmarshalling UserRegistered event")
		return err
	}

	// TODO process event
	c.Log.Infof("Received topic users with event: %v from partition %d", event, message.Partition)
	return nil
}
