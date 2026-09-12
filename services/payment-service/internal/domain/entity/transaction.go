package entity

import (
	"errors"
	"time"
)

type TransactionStatus string

const (
	StatusSucceeded TransactionStatus = "succeeded"
	StatusFailed    TransactionStatus = "failed"
	StatusRefunded  TransactionStatus = "refunded"
)

var (
	ErrInvalidAmount    = errors.New("o valor da transação deve ser maior que zero")
	ErrInvalidOrderID   = errors.New("id de pedido obrigatório")
	ErrInvalidGatewayID = errors.New("id de transação do gateway obrigatório")
)

type Transaction struct {
	ID                   string            `json:"id"`
	OrderID              string            `json:"order_id"`
	Amount               int64             `json:"amount"` // em centavos
	Status               TransactionStatus `json:"status"`
	PaymentMethod        string            `json:"payment_method"`
	GatewayTransactionID string            `json:"gateway_transaction_id"`
	CreatedAt            time.Time         `json:"created_at"`
}

func NewTransaction(
	id, orderID string,
	amount int64,
	status TransactionStatus,
	paymentMethod, gatewayTxID string,
) (*Transaction, error) {
	if id == "" || orderID == "" {
		return nil, ErrInvalidOrderID
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if gatewayTxID == "" {
		return nil, ErrInvalidGatewayID
	}
	if paymentMethod == "" {
		paymentMethod = "pix"
	}

	return &Transaction{
		ID:                   id,
		OrderID:              orderID,
		Amount:               amount,
		Status:               status,
		PaymentMethod:        paymentMethod,
		GatewayTransactionID: gatewayTxID,
		CreatedAt:            time.Now().UTC(),
	}, nil
}

type FinancialSummary struct {
	TotalRevenue  int64 `json:"total_revenue"`  // faturamento bruto em centavos
	TotalCount    int   `json:"total_count"`    // total de transações registradas
	SuccessCount  int   `json:"success_count"`  // pedidos pagos
	FailedCount   int   `json:"failed_count"`   // pagamentos recusados
	AverageTicket int64 `json:"average_ticket"` // ticket médio em centavos
}
