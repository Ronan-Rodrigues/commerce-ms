package notificar

import (
	"context"
	"log"

	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/domain/event"
	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/usecase/sse"
)

type N8NClient interface {
	DispatchWebhook(ctx context.Context, eventType string, payload interface{}) error
}

type NotificationOrchestrator struct {
	broker    *sse.Broker
	n8nClient N8NClient
}

func NewNotificationOrchestrator(broker *sse.Broker, n8nClient N8NClient) *NotificationOrchestrator {
	return &NotificationOrchestrator{
		broker:    broker,
		n8nClient: n8nClient,
	}
}

// ProcessIncomingEvent processa eventos do Redis Pub/Sub e roteia para SSE e n8n
func (o *NotificationOrchestrator) ProcessIncomingEvent(ctx context.Context, eventType string, rawPayload interface{}) {
	evt := event.NewNotificationEvent(eventType, rawPayload)

	// 1. Transmite via SSE para os navegadores conectados
	o.broker.BroadcastEvent(evt)
	log.Printf("[NOTIFY-SSE] Evento '%s' transmitido via SSE.", eventType)

	// 2. Dispara webhook para o n8n para automações de WhatsApp/WAHA
	if o.n8nClient != nil {
		go func() {
			if err := o.n8nClient.DispatchWebhook(context.Background(), eventType, rawPayload); err != nil {
				log.Printf("[NOTIFY-N8N] Aviso: falha ao enviar webhook ao n8n: %v", err)
			} else {
				log.Printf("[NOTIFY-N8N] Webhook do evento '%s' despachado com sucesso ao n8n.", eventType)
			}
		}()
	}
}
