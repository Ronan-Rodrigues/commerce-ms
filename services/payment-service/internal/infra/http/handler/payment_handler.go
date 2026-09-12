package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/domain/repository"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/usecase/financeiro"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/usecase/webhook"
)

type PaymentHandler struct {
	webhookUC *webhook.ProcessWebhookUseCase
	summaryUC *financeiro.FinancialSummaryUseCase
	txRepo    repository.TransactionRepository
}

func NewPaymentHandler(
	webhookUC *webhook.ProcessWebhookUseCase,
	summaryUC *financeiro.FinancialSummaryUseCase,
	txRepo repository.TransactionRepository,
) *PaymentHandler {
	return &PaymentHandler{
		webhookUC: webhookUC,
		summaryUC: summaryUC,
		txRepo:    txRepo,
	}
}

func renderJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func renderError(w http.ResponseWriter, status int, message string) {
	renderJSON(w, status, map[string]string{"error": message})
}

// HandleWebhook POST /payments/webhook
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get("X-Signature-SHA256")
	if signature == "" {
		renderError(w, http.StatusUnauthorized, "header X-Signature-SHA256 ausente")
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		renderError(w, http.StatusBadRequest, "erro ao ler corpo da requisição")
		return
	}

	var payload webhook.WebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		renderError(w, http.StatusBadRequest, "payload inválido")
		return
	}

	err = h.webhookUC.Execute(r.Context(), bodyBytes, signature, payload)
	if err != nil {
		if errors.Is(err, webhook.ErrInvalidWebhookSignature) {
			renderError(w, http.StatusUnauthorized, "assinatura de webhook inválida")
			return
		}
		renderError(w, http.StatusInternalServerError, "erro ao processar webhook")
		return
	}

	renderJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

// GetSummary GET /financial/summary
func (h *PaymentHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.summaryUC.Execute(r.Context())
	if err != nil {
		renderError(w, http.StatusInternalServerError, "erro ao obter resumo financeiro")
		return
	}
	renderJSON(w, http.StatusOK, summary)
}

// ListTransactions GET /financial/transactions
func (h *PaymentHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	transactions, err := h.txRepo.ListTransactions(r.Context(), limit, offset)
	if err != nil {
		renderError(w, http.StatusInternalServerError, "erro ao listar transações")
		return
	}
	renderJSON(w, http.StatusOK, transactions)
}

func (h *PaymentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	renderJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "payment-service",
	})
}
