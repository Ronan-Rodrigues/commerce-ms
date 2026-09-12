package webhook

import (
	"context"
	"errors"

	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/domain/repository"
	"github.com/google/uuid"
)

var (
	ErrInvalidWebhookSignature = errors.New("assinatura do webhook de pagamento inválida")
)

type WebhookPayload struct {
	Event                string `json:"event"` // "payment_intent.succeeded" ou "payment_intent.failed"
	OrderID              string `json:"order_id"`
	Amount               int64  `json:"amount"` // em centavos
	PaymentMethod        string `json:"payment_method"`
	GatewayTransactionID string `json:"gateway_transaction_id"`
}

type PaymentConfirmedEvent struct {
	OrderID              string `json:"order_id"`
	Amount               int64  `json:"amount"`
	PaymentMethod        string `json:"payment_method"`
	GatewayTransactionID string `json:"gateway_transaction_id"`
	Status               string `json:"status"`
}

type ProcessWebhookUseCase struct {
	txRepo        repository.TransactionRepository
	orderSvc      repository.OrderService
	publisher     repository.EventPublisher
	hmacValidator repository.HMACValidator
	webhookSecret string
}

func NewProcessWebhookUseCase(
	txRepo repository.TransactionRepository,
	orderSvc repository.OrderService,
	publisher repository.EventPublisher,
	hmacValidator repository.HMACValidator,
	webhookSecret string,
) *ProcessWebhookUseCase {
	return &ProcessWebhookUseCase{
		txRepo:        txRepo,
		orderSvc:      orderSvc,
		publisher:     publisher,
		hmacValidator: hmacValidator,
		webhookSecret: webhookSecret,
	}
}

func (uc *ProcessWebhookUseCase) Execute(ctx context.Context, rawBody []byte, signature string, payload WebhookPayload) error {
	// 1. Validação Criptográfica de Assinatura HMAC SHA256
	if !uc.hmacValidator.ValidateSignature(rawBody, signature, uc.webhookSecret) {
		return ErrInvalidWebhookSignature
	}

	// 2. Determina o status da transação
	var status entity.TransactionStatus
	var orderStatus string

	switch payload.Event {
	case "payment_intent.succeeded":
		status = entity.StatusSucceeded
		orderStatus = "pago"
	default:
		status = entity.StatusFailed
		orderStatus = "cancelado"
	}

	// 3. Registra a transação
	tx, err := entity.NewTransaction(
		uuid.NewString(),
		payload.OrderID,
		payload.Amount,
		status,
		payload.PaymentMethod,
		payload.GatewayTransactionID,
	)
	if err != nil {
		return err
	}

	if err := uc.txRepo.Create(ctx, tx); err != nil {
		return err
	}

	// 4. Notifica o order-service
	_ = uc.orderSvc.UpdateOrderStatus(ctx, payload.OrderID, orderStatus)

	// 5. Se aprovado, publica no Redis Pub/Sub para o notify-service / n8n
	if status == entity.StatusSucceeded {
		event := PaymentConfirmedEvent{
			OrderID:              payload.OrderID,
			Amount:               payload.Amount,
			PaymentMethod:        payload.PaymentMethod,
			GatewayTransactionID: payload.GatewayTransactionID,
			Status:               string(status),
		}
		_ = uc.publisher.Publish(ctx, "payment.confirmado", event)
	}

	return nil
}
