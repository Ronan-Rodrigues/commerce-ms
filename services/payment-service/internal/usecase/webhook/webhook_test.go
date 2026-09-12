package webhook_test

import (
	"context"
	"testing"

	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/usecase/webhook"
)

// MockTransactionRepository
type MockTransactionRepository struct {
	transactions []*entity.Transaction
}

func (m *MockTransactionRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	m.transactions = append(m.transactions, tx)
	return nil
}
func (m *MockTransactionRepository) FindByOrderID(ctx context.Context, orderID string) ([]*entity.Transaction, error) {
	return nil, nil
}
func (m *MockTransactionRepository) GetFinancialSummary(ctx context.Context) (*entity.FinancialSummary, error) {
	return nil, nil
}
func (m *MockTransactionRepository) ListTransactions(ctx context.Context, limit, offset int) ([]*entity.Transaction, error) {
	return m.transactions, nil
}

// MockOrderService
type MockOrderService struct {
	updatedOrders map[string]string // orderID -> status
}

func (m *MockOrderService) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	m.updatedOrders[orderID] = status
	return nil
}

// MockPublisher
type MockPublisher struct {
	publishedChannels []string
}

func (m *MockPublisher) Publish(ctx context.Context, channel string, event interface{}) error {
	m.publishedChannels = append(m.publishedChannels, channel)
	return nil
}

// MockHMACValidator
type MockHMACValidator struct{}

func (m *MockHMACValidator) ValidateSignature(payload []byte, signature, secret string) bool {
	return signature == "valid-sha256-signature"
}

func TestProcessWebhookUseCase_ValidPayment(t *testing.T) {
	txRepo := &MockTransactionRepository{}
	orderSvc := &MockOrderService{updatedOrders: make(map[string]string)}
	pub := &MockPublisher{}
	hmac := &MockHMACValidator{}

	uc := webhook.NewProcessWebhookUseCase(txRepo, orderSvc, pub, hmac, "my-secret")

	rawPayload := []byte(`{"event":"payment_intent.succeeded"}`)
	payload := webhook.WebhookPayload{
		Event:                "payment_intent.succeeded",
		OrderID:              "ord-999",
		Amount:               15000,
		PaymentMethod:        "pix",
		GatewayTransactionID: "gw-tx-123456",
	}

	err := uc.Execute(context.Background(), rawPayload, "valid-sha256-signature", payload)
	if err != nil {
		t.Fatalf("esperava sucesso no processamento do webhook, obteve: %v", err)
	}

	// Verifica transação gravada
	if len(txRepo.transactions) != 1 {
		t.Fatalf("esperava 1 transação registrada, obteve %d", len(txRepo.transactions))
	}
	if txRepo.transactions[0].Status != entity.StatusSucceeded {
		t.Errorf("status da transação incorreto: %s", txRepo.transactions[0].Status)
	}

	// Verifica pedido atualizado para "pago"
	if orderSvc.updatedOrders["ord-999"] != "pago" {
		t.Errorf("status do pedido deveria ser 'pago', obteve: %s", orderSvc.updatedOrders["ord-999"])
	}

	// Verifica evento payment.confirmado publicado
	if len(pub.publishedChannels) != 1 || pub.publishedChannels[0] != "payment.confirmado" {
		t.Errorf("esperava evento payment.confirmado, obteve: %v", pub.publishedChannels)
	}
}

func TestProcessWebhookUseCase_InvalidSignature(t *testing.T) {
	txRepo := &MockTransactionRepository{}
	orderSvc := &MockOrderService{updatedOrders: make(map[string]string)}
	pub := &MockPublisher{}
	hmac := &MockHMACValidator{}

	uc := webhook.NewProcessWebhookUseCase(txRepo, orderSvc, pub, hmac, "my-secret")

	rawPayload := []byte(`{"event":"payment_intent.succeeded"}`)
	payload := webhook.WebhookPayload{
		Event:   "payment_intent.succeeded",
		OrderID: "ord-999",
		Amount:  15000,
	}

	// Assinatura inválida/falsificada
	err := uc.Execute(context.Background(), rawPayload, "assinatura-falsificada", payload)
	if err != webhook.ErrInvalidWebhookSignature {
		t.Errorf("esperava ErrInvalidWebhookSignature, obteve: %v", err)
	}

	// Nenhuma transação pode ter sido gravada
	if len(txRepo.transactions) != 0 {
		t.Errorf("nenhuma transação deveria ser gravada para webhook com assinatura inválida")
	}
}
