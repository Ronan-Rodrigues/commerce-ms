package financeiro

import (
	"context"

	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/domain/repository"
)

type FinancialSummaryUseCase struct {
	txRepo repository.TransactionRepository
}

func NewFinancialSummaryUseCase(txRepo repository.TransactionRepository) *FinancialSummaryUseCase {
	return &FinancialSummaryUseCase{txRepo: txRepo}
}

func (uc *FinancialSummaryUseCase) Execute(ctx context.Context) (*entity.FinancialSummary, error) {
	return uc.txRepo.GetFinancialSummary(ctx)
}
