package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Ronan-Rodrigues/commerce-ms/services/notify-service/internal/usecase/sse"
)

type NotifyHandler struct {
	broker *sse.Broker
}

func NewNotifyHandler(broker *sse.Broker) *NotifyHandler {
	return &NotifyHandler{broker: broker}
}

// StreamEvents GET /events — Conexão SSE contínua com o frontend
func (h *NotifyHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	// 1. Verifica se o response writer suporta flushing contínuo
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming SSE não suportado pelo cliente", http.StatusBadRequest)
		return
	}

	// 2. Headers obrigatórios do protocolo SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 3. Registra o cliente no broker
	clientChan := h.broker.Subscribe()
	defer h.broker.Unsubscribe(clientChan)

	// Envia mensagem inicial de boas-vindas para abrir o canal
	fmt.Fprintf(w, "event: connected\ndata: {\"message\": \"SSE stream estabelecido com sucesso\"}\n\n")
	flusher.Flush()

	log.Printf("[SSE] Novo cliente conectado. Total conectados: %d", h.broker.TotalConnectedClients())

	// 4. Loop de envio de eventos em tempo real
	for {
		select {
		case <-r.Context().Done():
			// Cliente fechou o navegador ou desconectou
			log.Println("[SSE] Cliente desconectou.")
			return
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			w.Write(msg)
			flusher.Flush() // Envia imediatamente pela conexão TCP aberta
		}
	}
}

type wahaConfirmRequest struct {
	OrderID string `json:"order_id"`
	Action  string `json:"action"` // "confirmar" ou "cancelar"
}

// HandleConfirmation POST /webhook/confirmar — Chamado pelo n8n quando usuário responde no WhatsApp
func (h *NotifyHandler) HandleConfirmation(w http.ResponseWriter, r *http.Request) {
	var req wahaConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	log.Printf("[WAHA-CALLBACK] Confirmação recebida do WhatsApp para o pedido %s (Ação: %s)", req.OrderID, req.Action)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "received",
		"message": "resposta do WhatsApp processada",
	})
}

func (h *NotifyHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":            "ok",
		"service":           "notify-service",
		"connected_clients": h.broker.TotalConnectedClients(),
	})
}
