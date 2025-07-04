package payment

import (
    "context"
    "errors"
    "fmt"
    "time"
)

// Order representa una orden de compra
type Order struct {
    ID            string
    UserID        string
    Items         []OrderItem
    Total         float64
    PaymentMethod string
    Status        string
    CreatedAt     time.Time
}

type OrderItem struct {
    ProductID string
    Quantity  int
    Price     float64
}

// PaymentService maneja el procesamiento de pagos
type PaymentService struct {
    gateway    PaymentGateway
    inventory  InventoryService
    notifier   NotificationService
}

// ProcessOrder procesa una orden completa
func (s *PaymentService) ProcessOrder(ctx context.Context, order Order) error {
    // Validar orden
    if err := s.validateOrder(order); err != nil {
        return fmt.Errorf("order validation failed: %w", err)
    }
    
    // Verificar inventario
    available, err := s.checkInventoryAvailability(ctx, order.Items)
    if err != nil {
        return fmt.Errorf("inventory check failed: %w", err)
    }
    
    if !available {
        return errors.New("insufficient inventory")
    }
    
    // Procesar pago
    paymentResult, err := s.processPayment(ctx, order)
    if err != nil {
        return fmt.Errorf("payment processing failed: %w", err)
    }
    
    // Actualizar inventario
    if err := s.updateInventory(ctx, order.Items); err != nil {
        // Rollback payment if inventory update fails
        s.rollbackPayment(ctx, paymentResult.TransactionID)
        return fmt.Errorf("inventory update failed: %w", err)
    }
    
    // Enviar notificaciones
    go s.sendOrderConfirmation(context.Background(), order, paymentResult)
    
    return nil
}

// validateOrder valida las reglas de negocio de la orden
func (s *PaymentService) validateOrder(order Order) error {
    if order.Total <= 0 {
        return errors.New("order total must be positive")
    }
    
    if len(order.Items) == 0 {
        return errors.New("order must contain at least one item")
    }
    
    // Validar método de pago
    validMethods := []string{"credit_card", "debit_card", "paypal"}
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

// checkInventoryAvailability verifica disponibilidad de productos
func (s *PaymentService) checkInventoryAvailability(ctx context.Context, items []OrderItem) (bool, error) {
    for _, item := range items {
        stock, err := s.inventory.GetStock(ctx, item.ProductID)
        if err != nil {
            return false, err
        }
        
        if stock < item.Quantity {
            return false, nil
        }
    }
    
    return true, nil
}

// processPayment procesa el pago con el gateway
func (s *PaymentService) processPayment(ctx context.Context, order Order) (*PaymentResult, error) {
    // Calcular fees
    fees := s.calculateProcessingFees(order.Total, order.PaymentMethod)
    totalWithFees := order.Total + fees
    
    // Intentar cobro
    result, err := s.gateway.Charge(ctx, ChargeRequest{
        Amount:        totalWithFees,
        Currency:      "USD",
        CustomerID:    order.UserID,
        OrderID:       order.ID,
        PaymentMethod: order.PaymentMethod,
    })
    
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

// calculateProcessingFees calcula las comisiones según el método de pago
func (s *PaymentService) calculateProcessingFees(amount float64, method string) float64 {
    switch method {
    case "credit_card":
        return amount * 0.029 + 0.30 // 2.9% + $0.30
    case "debit_card":
        return amount * 0.025 + 0.25 // 2.5% + $0.25
    case "paypal":
        return amount * 0.034 + 0.30 // 3.4% + $0.30
    default:
        return 0
    }
}

// updateInventory actualiza el inventario después del pago exitoso
func (s *PaymentService) updateInventory(ctx context.Context, items []OrderItem) error {
    for _, item := range items {
        if err := s.inventory.DeductStock(ctx, item.ProductID, item.Quantity); err != nil {
            return err
        }
    }
    return nil
}

// rollbackPayment revierte un pago en caso de error
func (s *PaymentService) rollbackPayment(ctx context.Context, transactionID string) {
    // Este es un proceso asíncrono crítico
    err := s.gateway.Refund(ctx, transactionID)
    if err != nil {
        // Log critical error - this needs manual intervention
        fmt.Printf("CRITICAL: Failed to rollback payment %s: %v\n", transactionID, err)
    }
}

// sendOrderConfirmation envía confirmación de orden al usuario
func (s *PaymentService) sendOrderConfirmation(ctx context.Context, order Order, payment *PaymentResult) {
    notification := Notification{
        UserID:      order.UserID,
        Type:        "order_confirmation",
        Subject:     fmt.Sprintf("Order %s confirmed", order.ID),
        OrderID:     order.ID,
        PaymentInfo: payment,
    }
    
    if err := s.notifier.Send(ctx, notification); err != nil {
        fmt.Printf("Failed to send order confirmation: %v\n", err)
    }
}

// Interfaces para dependencias externas
type PaymentGateway interface {
    Charge(ctx context.Context, req ChargeRequest) (*PaymentResult, error)
    Refund(ctx context.Context, transactionID string) error
}

type InventoryService interface {
    GetStock(ctx context.Context, productID string) (int, error)
    DeductStock(ctx context.Context, productID string, quantity int) error
}

type NotificationService interface {
    Send(ctx context.Context, notification Notification) error
}

// Tipos de datos
type ChargeRequest struct {
    Amount        float64
    Currency      string
    CustomerID    string
    OrderID       string
    PaymentMethod string
}

type PaymentResult struct {
    TransactionID string
    Status        string
    Amount        float64
    Timestamp     time.Time
}

type Notification struct {
    UserID      string
    Type        string
    Subject     string
    OrderID     string
    PaymentInfo *PaymentResult
}