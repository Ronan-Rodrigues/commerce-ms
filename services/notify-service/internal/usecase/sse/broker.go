package sse

import (
	"sync"

	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/domain/event"
)

type Broker struct {
	clients    map[chan []byte]bool
	register   chan chan []byte
	unregister chan chan []byte
	broadcast  chan *event.NotificationEvent
	mu         sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		clients:    make(map[chan []byte]bool),
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
		broadcast:  make(chan *event.NotificationEvent, 100),
	}
}

// Start roda em background gerenciando o ciclo de vida das conexões SSE
func (b *Broker) Start() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = true
			b.mu.Unlock()

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client)
			}
			b.mu.Unlock()

		case evt := <-b.broadcast:
			msg, err := evt.FormatSSE()
			if err != nil {
				continue
			}

			b.mu.RLock()
			for client := range b.clients {
				select {
				case client <- msg:
				default:
					// Se o canal do cliente estiver cheio, pula para não travar os outros clientes
				}
			}
			b.mu.RUnlock()
		}
	}
}

func (b *Broker) Subscribe() chan []byte {
	client := make(chan []byte, 10)
	b.register <- client
	return client
}

func (b *Broker) Unsubscribe(client chan []byte) {
	b.unregister <- client
}

func (b *Broker) BroadcastEvent(evt *event.NotificationEvent) {
	b.broadcast <- evt
}

func (b *Broker) TotalConnectedClients() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
