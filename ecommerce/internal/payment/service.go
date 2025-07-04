package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/example/ecommerce/internal/inventory"
	"github.com/example/ecommerce/internal/notification"
	"github.com/example/ecommerce/pkg/models"
)

type Service struct {
	gateway    *Gateway
	inventory  *inventory.Service
	notifier   *notification.Service
}

func NewService(gateway *Gateway, inv *inventory.Service, notifier *notification.Service) *Service {
	return &Service{
		gateway:   gateway,
		inventory: inv,
		notifier:  notifier,
	}
}

func (s *Service) ProcessOrderPayment(ctx context.Context, order *models.Order) (*PaymentResult, error) {
	if err := s.validatePaymentRequest(order); err != nil {
		return nil, fmt.Errorf("payment validation failed: %w", err)
	}

	available, err := s.checkInventoryForOrder(ctx, order.Items)
	if err != nil {
		return nil, fmt.Errorf("inventory check failed: %w", err)
	}

	if !available {
		return nil, errors.New("insufficient inventory for order")
	}

	result, err := s.processPaymentTransaction(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("payment transaction failed: %w", err)
	}

	if err := s.updateInventoryAfterPayment(ctx, order.Items); err != nil {
		go s.initiatePaymentRollback(context.Background(), result.TransactionID)
		return nil, fmt.Errorf("inventory update failed: %w", err)
	}

	go s.sendPaymentNotification(context.Background(), order, result)

	return result, nil
}

func (s *Service) validatePaymentRequest(order *models.Order) error {
	if order.Total <= 0 {
		return errors.New("order total must be positive")
	}

	if len(order.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	validMethods := []string{"credit_card", "debit_card", "paypal", "bank_transfer"}
	valid := false
	for _, method := range validMethods {
		if order.PaymentMethod == method {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid payment method: %s", order.PaymentMethod)
	}

	return nil
}

func (s *Service) checkInventoryForOrder(ctx context.Context, items []models.OrderItem) (bool, error) {
	for _, item := range items {
		available, err := s.inventory.CheckAvailability(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return false, err
		}
		if !available {
			return false, nil
		}
	}
	return true, nil
}

func (s *Service) processPaymentTransaction(ctx context.Context, order *models.Order) (*PaymentResult, error) {
	fees := s.calculateProcessingFees(order.Total, order.PaymentMethod)
	totalWithFees := order.Total + fees

	request := &ChargeRequest{
		Amount:        totalWithFees,
		Currency:      "USD",
		CustomerID:    order.UserID,
		OrderID:       order.ID,
		PaymentMethod: order.PaymentMethod,
		Description:   fmt.Sprintf("Order #%s", order.ID),
	}

	result, err := s.gateway.ProcessCharge(ctx, request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) calculateProcessingFees(amount float64, method string) float64 {
	switch method {
	case "credit_card":
		return amount * 0.029 + 0.30
	case "debit_card":
		return amount * 0.025 + 0.25
	case "paypal":
		return amount * 0.034 + 0.30
	case "bank_transfer":
		return amount * 0.01
	default:
		return 0
	}
}

func (s *Service) updateInventoryAfterPayment(ctx context.Context, items []models.OrderItem) error {
	for _, item := range items {
		if err := s.inventory.DeductStock(ctx, item.ProductID, item.Quantity); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) initiatePaymentRollback(ctx context.Context, transactionID string) {
	if err := s.gateway.RefundTransaction(ctx, transactionID); err != nil {
		fmt.Printf("CRITICAL: Failed to rollback payment %s: %v\n", transactionID, err)
	}
}

func (s *Service) sendPaymentNotification(ctx context.Context, order *models.Order, payment *PaymentResult) {
	notification := &notification.Message{
		UserID:  order.UserID,
		Type:    "payment_confirmation",
		Subject: fmt.Sprintf("Payment confirmed for Order #%s", order.ID),
		Content: fmt.Sprintf("Your payment of $%.2f has been processed successfully.", payment.Amount),
		Data: map[string]interface{}{
			"order_id":       order.ID,
			"transaction_id": payment.TransactionID,
			"amount":         payment.Amount,
		},
	}

	if err := s.notifier.Send(ctx, notification); err != nil {
		fmt.Printf("Failed to send payment notification: %v\n", err)
	}
}

type PaymentResult struct {
	TransactionID string
	Status        string
	Amount        float64
	Fees          float64
	Timestamp     time.Time
}

type ChargeRequest struct {
	Amount        float64
	Currency      string
	CustomerID    string
	OrderID       string
	PaymentMethod string
	Description   string
}