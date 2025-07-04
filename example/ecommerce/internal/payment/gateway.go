package payment

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type Gateway struct {
	apiKey string
}

func NewGateway() *Gateway {
	return &Gateway{
		apiKey: "test-api-key",
	}
}

func (g *Gateway) ProcessCharge(ctx context.Context, req *ChargeRequest) (*PaymentResult, error) {
	if err := g.validateChargeRequest(req); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Duration(100+rand.Intn(400)) * time.Millisecond):
	}

	if rand.Float64() < 0.05 {
		return nil, errors.New("payment gateway timeout")
	}

	if req.Amount > 10000 {
		return nil, errors.New("amount exceeds maximum limit")
	}

	transactionID := g.generateTransactionID()
	
	result := &PaymentResult{
		TransactionID: transactionID,
		Status:        "approved",
		Amount:        req.Amount,
		Timestamp:     time.Now(),
	}

	return result, nil
}

func (g *Gateway) RefundTransaction(ctx context.Context, transactionID string) error {
	if transactionID == "" {
		return errors.New("transaction ID is required")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(200+rand.Intn(300)) * time.Millisecond):
	}

	if rand.Float64() < 0.02 {
		return errors.New("refund failed: transaction not found")
	}

	return nil
}

func (g *Gateway) validateChargeRequest(req *ChargeRequest) error {
	if req.Amount <= 0 {
		return errors.New("charge amount must be positive")
	}

	if req.CustomerID == "" {
		return errors.New("customer ID is required")
	}

	if req.OrderID == "" {
		return errors.New("order ID is required")
	}

	supportedCurrencies := []string{"USD", "EUR", "GBP"}
	supported := false
	for _, curr := range supportedCurrencies {
		if req.Currency == curr {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported currency: %s", req.Currency)
	}

	return nil
}

func (g *Gateway) generateTransactionID() string {
	timestamp := time.Now().UnixNano()
	random := rand.Int63n(999999)
	return fmt.Sprintf("txn_%d_%06d", timestamp, random)
}