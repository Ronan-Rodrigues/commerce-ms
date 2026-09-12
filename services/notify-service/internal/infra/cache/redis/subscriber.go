package redis

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/usecase/notificar"
	"github.com/redis/go-redis/v9"
)

type RedisSubscriber struct {
	client       *redis.Client
	orchestrator *notificar.NotificationOrchestrator
	channels     []string
}

func NewRedisSubscriber(
	client *redis.Client,
	orchestrator *notificar.NotificationOrchestrator,
	channels []string,
) *RedisSubscriber {
	return &RedisSubscriber{
		client:       client,
		orchestrator: orchestrator,
		channels:     channels,
	}
}

// StartListening escuta os canais do Redis Pub/Sub em segundo plano
func (s *RedisSubscriber) StartListening(ctx context.Context) {
	pubsub := s.client.Subscribe(ctx, s.channels...)
	defer pubsub.Close()

	ch := pubsub.Channel()
	log.Printf("[REDIS-PUBSUB] Inscrito nos canais: %v", s.channels)

	for {
		select {
		case <-ctx.Done():
			log.Println("[REDIS-PUBSUB] Encerrando subscrição nos canais.")
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			// Decodifica JSON genérico para repassar
			var rawPayload interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &rawPayload); err != nil {
				rawPayload = msg.Payload
			}

			s.orchestrator.ProcessIncomingEvent(ctx, msg.Channel, rawPayload)
		}
	}
}
