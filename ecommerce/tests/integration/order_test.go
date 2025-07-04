package integration

import (
	"context"
	"testing"
	"time"

	"github.com/example/ecommerce/internal/inventory"
	"github.com/example/ecommerce/internal/notification"
	"github.com/example/ecommerce/internal/order"
	"github.com/example/ecommerce/internal/payment"
	"github.com/example/ecommerce/pkg/models"
)

func TestCreateOrder_Success(t *testing.T) {
	inventorySvc := inventory.NewService()
	notificationSvc := notification.NewService()
	paymentGateway := payment.NewGateway()
	paymentSvc := payment.NewService(paymentGateway, inventorySvc, notificationSvc)
	orderSvc := order.NewService(paymentSvc, inventorySvc)

	ctx := context.Background()
	req := order.CreateOrderRequest{
		UserID: "USR-TEST-001",
		Items: []models.OrderItem{
			{
				ProductID: "PROD-001",
				Quantity:  2,
				Price:     1299.99,
			},
			{
				ProductID: "PROD-002",
				Quantity:  1,
				Price:     49.99,
			},
		},
		PaymentMethod: "credit_card",
		ShippingAddr: models.Address{
			Street:  "123 Test St",
			City:    "Test City",
			State:   "TS",
			ZipCode: "12345",
			Country: "US",
		},
	}

	order, err := orderSvc.CreateOrder(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	if order.ID == "" {
		t.Error("Order ID should not be empty")
	}

	if order.Status != models.OrderStatusPaid {
		t.Errorf("Expected order status to be %s, got %s", models.OrderStatusPaid, order.Status)
	}

	if order.Total <= 0 {
		t.Error("Order total should be greater than 0")
	}

	time.Sleep(100 * time.Millisecond)
}

func TestCreateOrder_InsufficientInventory(t *testing.T) {
	inventorySvc := inventory.NewService()
	notificationSvc := notification.NewService()
	paymentGateway := payment.NewGateway()
	paymentSvc := payment.NewService(paymentGateway, inventorySvc, notificationSvc)
	orderSvc := order.NewService(paymentSvc, inventorySvc)

	ctx := context.Background()
	req := order.CreateOrderRequest{
		UserID: "USR-TEST-002",
		Items: []models.OrderItem{
			{
				ProductID: "PROD-001",
				Quantity:  1000,
				Price:     1299.99,
			},
		},
		PaymentMethod: "credit_card",
		ShippingAddr: models.Address{
			Street:  "123 Test St",
			City:    "Test City",
			State:   "TS",
			ZipCode: "12345",
			Country: "US",
		},
	}

	_, err := orderSvc.CreateOrder(ctx, req)
	if err == nil {
		t.Fatal("Expected error for insufficient inventory, got nil")
	}

	if err.Error() != "payment processing failed: insufficient inventory for order" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestCreateOrder_InvalidPaymentMethod(t *testing.T) {
	inventorySvc := inventory.NewService()
	notificationSvc := notification.NewService()
	paymentGateway := payment.NewGateway()
	paymentSvc := payment.NewService(paymentGateway, inventorySvc, notificationSvc)
	orderSvc := order.NewService(paymentSvc, inventorySvc)

	ctx := context.Background()
	req := order.CreateOrderRequest{
		UserID: "USR-TEST-003",
		Items: []models.OrderItem{
			{
				ProductID: "PROD-001",
				Quantity:  1,
				Price:     1299.99,
			},
		},
		PaymentMethod: "invalid_method",
		ShippingAddr: models.Address{
			Street:  "123 Test St",
			City:    "Test City",
			State:   "TS",
			ZipCode: "12345",
			Country: "US",
		},
	}

	_, err := orderSvc.CreateOrder(ctx, req)
	if err == nil {
		t.Fatal("Expected error for invalid payment method, got nil")
	}
}