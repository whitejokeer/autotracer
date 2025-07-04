package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/example/ecommerce/internal/inventory"
	"github.com/example/ecommerce/internal/payment"
	"github.com/example/ecommerce/pkg/models"
)

type Service struct {
	paymentSvc  *payment.Service
	inventorySvc *inventory.Service
	repository   *Repository
}

func NewService(paymentSvc *payment.Service, inventorySvc *inventory.Service) *Service {
	return &Service{
		paymentSvc:   paymentSvc,
		inventorySvc: inventorySvc,
		repository:   NewRepository(),
	}
}

type CreateOrderRequest struct {
	UserID        string              `json:"user_id"`
	Items         []models.OrderItem  `json:"items"`
	PaymentMethod string              `json:"payment_method"`
	ShippingAddr  models.Address      `json:"shipping_address"`
}

func (s *Service) CreateOrder(ctx context.Context, req CreateOrderRequest) (*models.Order, error) {
	if err := s.validateOrderRequest(req); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	order := &models.Order{
		ID:            s.generateOrderID(),
		UserID:        req.UserID,
		Items:         req.Items,
		PaymentMethod: req.PaymentMethod,
		ShippingAddr:  req.ShippingAddr,
		Status:        models.OrderStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	total, err := s.calculateOrderTotal(ctx, order.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate order total: %w", err)
	}
	order.Total = total

	if err := s.repository.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	paymentResult, err := s.paymentSvc.ProcessOrderPayment(ctx, order)
	if err != nil {
		order.Status = models.OrderStatusCancelled
		s.repository.Update(ctx, order)
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	order.Status = models.OrderStatusPaid
	if err := s.repository.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	go s.initiateOrderFulfillment(context.Background(), order, paymentResult.TransactionID)

	return order, nil
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	if orderID == "" {
		return nil, errors.New("order ID is required")
	}

	order, err := s.repository.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	return order, nil
}

func (s *Service) CancelOrder(ctx context.Context, orderID string, userID string) error {
	order, err := s.repository.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if order.UserID != userID {
		return errors.New("unauthorized: cannot cancel another user's order")
	}

	if order.Status != models.OrderStatusPending && order.Status != models.OrderStatusPaid {
		return fmt.Errorf("cannot cancel order with status: %s", order.Status)
	}

	order.Status = models.OrderStatusCancelled
	order.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	if order.Status == models.OrderStatusPaid {
		go s.processOrderRefund(context.Background(), order)
	}

	return nil
}

func (s *Service) validateOrderRequest(req CreateOrderRequest) error {
	if req.UserID == "" {
		return errors.New("user ID is required")
	}

	if len(req.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return fmt.Errorf("invalid quantity for product %s", item.ProductID)
		}
		if item.Price <= 0 {
			return fmt.Errorf("invalid price for product %s", item.ProductID)
		}
	}

	if req.ShippingAddr.Street == "" || req.ShippingAddr.City == "" {
		return errors.New("incomplete shipping address")
	}

	return nil
}

func (s *Service) calculateOrderTotal(ctx context.Context, items []models.OrderItem) (float64, error) {
	var total float64
	
	for i, item := range items {
		product, err := s.inventorySvc.GetProduct(ctx, item.ProductID)
		if err != nil {
			return 0, fmt.Errorf("product not found: %s", item.ProductID)
		}

		items[i].ProductName = product.Name
		items[i].Price = product.Price
		items[i].Subtotal = product.Price * float64(item.Quantity)
		total += items[i].Subtotal
	}

	return total, nil
}

func (s *Service) initiateOrderFulfillment(ctx context.Context, order *models.Order, transactionID string) {
	time.Sleep(2 * time.Second)

	order.Status = models.OrderStatusProcessing
	order.UpdatedAt = time.Now()
	
	if err := s.repository.Update(ctx, order); err != nil {
		fmt.Printf("Failed to update order status to processing: %v\n", err)
	}
}

func (s *Service) processOrderRefund(ctx context.Context, order *models.Order) {
	fmt.Printf("Processing refund for cancelled order %s\n", order.ID)
}

func (s *Service) generateOrderID() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("ORD-%d-%04d", timestamp, time.Now().Nanosecond()%10000)
}