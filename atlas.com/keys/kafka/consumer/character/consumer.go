package character

import (
	consumer2 "atlas-keys/kafka/consumer"
	"atlas-keys/kafka/message/character"
	"atlas-keys/key"
	"context"
	"github.com/Chronicle20/atlas-kafka/consumer"
	"github.com/Chronicle20/atlas-kafka/handler"
	"github.com/Chronicle20/atlas-kafka/message"
	"github.com/Chronicle20/atlas-kafka/topic"
	"github.com/Chronicle20/atlas-model/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitConsumers(l logrus.FieldLogger) func(func(config consumer.Config, decorators ...model.Decorator[consumer.Config])) func(consumerGroupId string) {
	return func(rf func(config consumer.Config, decorators ...model.Decorator[consumer.Config])) func(consumerGroupId string) {
		return func(consumerGroupId string) {
			rf(consumer2.NewConfig(l)("character_created")(character.EnvEventTopicStatus)(consumerGroupId), consumer.SetHeaderParsers(consumer.SpanHeaderParser, consumer.TenantHeaderParser))
		}
	}
}

func InitHandlers(l logrus.FieldLogger) func(db *gorm.DB) func(rf func(topic string, handler handler.Handler) (string, error)) {
	return func(db *gorm.DB) func(rf func(topic string, handler handler.Handler) (string, error)) {
		return func(rf func(topic string, handler handler.Handler) (string, error)) {
			var t string
			t, _ = topic.EnvProvider(l)(character.EnvEventTopicStatus)()
			_, _ = rf(t, message.AdaptHandler(message.PersistentConfig(handleStatusEventCreated(db))))
			_, _ = rf(t, message.AdaptHandler(message.PersistentConfig(handleStatusEventDeleted(db))))
		}
	}
}

func handleStatusEventCreated(db *gorm.DB) func(l logrus.FieldLogger, ctx context.Context, e character.StatusEvent[character.CreatedStatusBody]) {
	return func(l logrus.FieldLogger, ctx context.Context, e character.StatusEvent[character.CreatedStatusBody]) {
		if e.Type != character.StatusEventTypeCreated {
			return
		}

		err := key.CreateDefault(l)(ctx)(db)(e.CharacterId)
		if err != nil {
			l.WithError(err).Errorf("Unable to create default keymapping for character [%d].", e.CharacterId)
		}
	}
}

func handleStatusEventDeleted(db *gorm.DB) message.Handler[character.StatusEvent[character.DeletedStatusEventBody]] {
	return func(l logrus.FieldLogger, ctx context.Context, e character.StatusEvent[character.DeletedStatusEventBody]) {
		if e.Type != character.StatusEventTypeDeleted {
			return
		}

		err := key.Delete(l)(ctx)(db)(e.CharacterId)
		if err != nil {
			l.WithError(err).Errorf("Unable to delete for character [%d].", e.CharacterId)
		}
	}
}
