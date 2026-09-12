package sse_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/domain/event"
	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/usecase/sse"
)

func TestSSEBroker_Broadcast(t *testing.T) {
	broker := sse.NewBroker()
	go broker.Start()

	// 1. Conecta cliente 1
	client1 := broker.Subscribe()
	defer broker.Unsubscribe(client1)

	// Dá tempo para o loop registrar
	time.Sleep(10 * time.Millisecond)

	if broker.TotalConnectedClients() != 1 {
		t.Errorf("esperava 1 cliente conectado, obteve %d", broker.TotalConnectedClients())
	}

	// 2. Transmite evento
	evt := event.NewNotificationEvent("payment.confirmado", map[string]string{
		"order_id": "ord-100",
		"status":   "pago",
	})
	broker.BroadcastEvent(evt)

	// 3. Cliente 1 deve receber o evento no formato SSE
	select {
	case msg := <-client1:
		strMsg := string(msg)
		if !strings.Contains(strMsg, "event: payment.confirmado") {
			t.Errorf("tipo do evento incorreto na mensagem SSE: %s", strMsg)
		}
		if !strings.Contains(strMsg, `"order_id":"ord-100"`) {
			t.Errorf("payload incorreto na mensagem SSE: %s", strMsg)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout aguardando mensagem no canal SSE do cliente")
	}
}

func TestSSEBroker_Unsubscribe(t *testing.T) {
	broker := sse.NewBroker()
	go broker.Start()

	client := broker.Subscribe()
	time.Sleep(10 * time.Millisecond)

	broker.Unsubscribe(client)
	time.Sleep(10 * time.Millisecond)

	if broker.TotalConnectedClients() != 0 {
		t.Errorf("esperava 0 clientes conectados após unsubscribe, obteve %d", broker.TotalConnectedClients())
	}
}
