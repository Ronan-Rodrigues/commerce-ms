package redis

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/entity"
	"github.com/redis/go-redis/v9"
)

type AbandonedCartEvent struct {
	UserID      string             `json:"user_id"`
	Items       []*entity.CartItem `json:"items"`
	TotalAmount int64              `json:"total_amount"`
	AbandonedAt time.Time          `json:"abandoned_at"`
}

type KeyspaceListener struct {
	client   *redis.Client
	cartRepo *RedisCartRepository
}

func NewKeyspaceListener(client *redis.Client, cartRepo *RedisCartRepository) *KeyspaceListener {
	return &KeyspaceListener{
		client:   client,
		cartRepo: cartRepo,
	}
}

// StartListening escuta eventos de expiração de chave no Redis em segundo plano
func (l *KeyspaceListener) StartListening(ctx context.Context) {
	// Subscrição no canal de chaves expiradas do banco 0
	pubsub := l.client.Subscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	ch := pubsub.Channel()
	log.Println("[REDIS-KEYSPACE] Ouvinte de chaves expiradas ativado (detecção de carrinho abandonado).")

	for {
		select {
		case <-ctx.Done():
			log.Println("[REDIS-KEYSPACE] Encerrando ouvinte de chaves.")
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			// msg.Payload contém a chave expirada, ex: "cart:usr-123"
			expiredKey := msg.Payload
			if strings.HasPrefix(expiredKey, "cart:") {
				userID := strings.TrimPrefix(expiredKey, "cart:")
				l.handleCartAbandonment(ctx, userID)
			}
		}
	}
}

func (l *KeyspaceListener) handleCartAbandonment(ctx context.Context, userID string) {
	log.Printf("[CARRINHO-ABANDONADO] TTL expirou para o usuário: %s", userID)

	// Busca snapshot dos itens que o usuário deixou no carrinho
	snapshot, err := l.cartRepo.GetSnapshot(ctx, userID)
	if err != nil || snapshot == nil || len(snapshot.Items) == 0 {
		return
	}

	event := AbandonedCartEvent{
		UserID:      userID,
		Items:       snapshot.Items,
		TotalAmount: snapshot.TotalAmount,
		AbandonedAt: time.Now().UTC(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("[ERRO] Falha ao serializar evento de abandono: %v", err)
		return
	}

	// Publica no Pub/Sub para o notify-service / n8n
	if err := l.client.Publish(ctx, "cart.abandonado", payload).Err(); err != nil {
		log.Printf("[ERRO] Falha ao publicar evento cart.abandonado: %v", err)
		return
	}

	log.Printf("[EVENTO] cart.abandonado publicado para usuário %s com total de R$ %.2f", userID, float64(snapshot.TotalAmount)/100.0)
}
