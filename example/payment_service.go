package payment

import (
    "context"
    "errors"
    "fmt"
    "time"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
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

var tracer = otel.Tracer("payment-service")

// PaymentService maneja el procesamiento de pagos
type PaymentService struct {
    gateway    PaymentGateway
    inventory  InventoryService
    notifier   NotificationService
}

// ProcessOrder procesa una orden completa
func (s *PaymentService) ProcessOrder(ctx context.Context, order Order) error {
    ctx, span := tracer.Start(ctx, "ProcessOrder",
        trace.WithAttributes(
            attribute.String("order.id", order.ID),
            attribute.String("order.user_id", order.UserID),
            attribute.Float64("order.total", order.Total),
            attribute.String("order.payment_method", order.PaymentMethod),
            attribute.Int("order.items_count", len(order.Items)),
        ),
    )
    defer span.End()
    
    // Validar orden
    if err := s.validateOrder(order); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return fmt.Errorf("order validation failed: %w", err)
    }
    
    // Verificar inventario
    available, err := s.checkInventoryAvailability(ctx, order.Items)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return fmt.Errorf("inventory check failed: %w", err)
    }
    
    if !available {
        err := errors.New("insufficient inventory")
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    
    // Procesar pago
    paymentResult, err := s.processPayment(ctx, order)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return fmt.Errorf("payment processing failed: %w", err)
    }
    
    span.AddEvent("payment.processed", trace.WithAttributes(
        attribute.String("payment.transaction_id", paymentResult.TransactionID),
        attribute.Float64("payment.amount", paymentResult.Amount),
        attribute.String("payment.status", paymentResult.Status),
    ))
    
    // Actualizar inventario
    if err := s.updateInventory(ctx, order.Items); err != nil {
        span.RecordError(err)
        span.AddEvent("payment.rollback.initiated", trace.WithAttributes(
            attribute.String("payment.transaction_id", paymentResult.TransactionID),
        ))
        // Rollback payment if inventory update fails
        s.rollbackPayment(ctx, paymentResult.TransactionID)
        span.SetStatus(codes.Error, err.Error())
        return fmt.Errorf("inventory update failed: %w", err)
    }
    
    // Enviar notificaciones
    go s.sendOrderConfirmation(context.Background(), order, paymentResult)
    
    span.AddEvent("order.completed", trace.WithAttributes(
        attribute.String("order.id", order.ID),
        attribute.String("order.status", "completed"),
    ))
    
    return nil
}

// validateOrder valida las reglas de negocio de la orden
func (s *PaymentService) validateOrder(order Order) error {
    _, span := tracer.Start(context.Background(), "validateOrder",
        trace.WithAttributes(
            attribute.String("order.id", order.ID),
            attribute.Float64("order.total", order.Total),
            attribute.String("order.payment_method", order.PaymentMethod),
        ),
    )
    defer span.End()
    if order.Total <= 0 {
        err := errors.New("order total must be positive")
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    
    if len(order.Items) == 0 {
        err := errors.New("order must contain at least one item")
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
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
        err := fmt.Errorf("invalid payment method: %s", order.PaymentMethod)
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    
    return nil
}

// checkInventoryAvailability verifica disponibilidad de productos
func (s *PaymentService) checkInventoryAvailability(ctx context.Context, items []OrderItem) (bool, error) {
    ctx, span := tracer.Start(ctx, "checkInventoryAvailability",
        trace.WithAttributes(
            attribute.Int("inventory.items_count", len(items)),
        ),
    )
    defer span.End()
    for _, item := range items {
        stock, err := s.inventory.GetStock(ctx, item.ProductID)
        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
            return false, err
        }
        
        if stock < item.Quantity {
            span.AddEvent("inventory.insufficient", trace.WithAttributes(
                attribute.String("product.id", item.ProductID),
                attribute.Int("stock.available", stock),
                attribute.Int("stock.required", item.Quantity),
            ))
            return false, nil
        }
    }
    
    return true, nil
}

// processPayment procesa el pago con el gateway
func (s *PaymentService) processPayment(ctx context.Context, order Order) (*PaymentResult, error) {
    ctx, span := tracer.Start(ctx, "processPayment",
        trace.WithAttributes(
            attribute.String("order.id", order.ID),
            attribute.Float64("order.total", order.Total),
            attribute.String("payment.method", order.PaymentMethod),
        ),
    )
    defer span.End()
    // Calcular fees
    fees := s.calculateProcessingFees(order.Total, order.PaymentMethod)
    totalWithFees := order.Total + fees
    
    span.AddEvent("payment.fees_calculated", trace.WithAttributes(
        attribute.Float64("payment.fees", fees),
        attribute.Float64("payment.total_with_fees", totalWithFees),
    ))
    
    // Intentar cobro
    result, err := s.gateway.Charge(ctx, ChargeRequest{
        Amount:        totalWithFees,
        Currency:      "USD",
        CustomerID:    order.UserID,
        OrderID:       order.ID,
        PaymentMethod: order.PaymentMethod,
    })
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    
    return result, nil
}

// calculateProcessingFees calcula las comisiones según el método de pago
func (s *PaymentService) calculateProcessingFees(amount float64, method string) float64 {
    _, span := tracer.Start(context.Background(), "calculateProcessingFees",
        trace.WithAttributes(
            attribute.Float64("payment.amount", amount),
            attribute.String("payment.method", method),
        ),
    )
    defer span.End()
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
    ctx, span := tracer.Start(ctx, "updateInventory",
        trace.WithAttributes(
            attribute.Int("inventory.items_count", len(items)),
        ),
    )
    defer span.End()
    for _, item := range items {
        if err := s.inventory.DeductStock(ctx, item.ProductID, item.Quantity); err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
            return err
        }
    }
    return nil
}

// rollbackPayment revierte un pago en caso de error
func (s *PaymentService) rollbackPayment(ctx context.Context, transactionID string) {
    ctx, span := tracer.Start(ctx, "rollbackPayment",
        trace.WithAttributes(
            attribute.String("payment.transaction_id", transactionID),
        ),
    )
    defer span.End()
    // Este es un proceso asíncrono crítico
    err := s.gateway.Refund(ctx, transactionID)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        span.AddEvent("payment.rollback.failed", trace.WithAttributes(
            attribute.String("payment.transaction_id", transactionID),
        ))
        // Log critical error - this needs manual intervention
        fmt.Printf("CRITICAL: Failed to rollback payment %s: %v\n", transactionID, err)
    } else {
        span.AddEvent("payment.rollback.completed", trace.WithAttributes(
            attribute.String("payment.transaction_id", transactionID),
        ))
    }
}

// sendOrderConfirmation envía confirmación de orden al usuario
func (s *PaymentService) sendOrderConfirmation(ctx context.Context, order Order, payment *PaymentResult) {
    ctx, span := tracer.Start(ctx, "sendOrderConfirmation",
        trace.WithAttributes(
            attribute.String("order.id", order.ID),
            attribute.String("order.user_id", order.UserID),
            attribute.String("payment.transaction_id", payment.TransactionID),
        ),
    )
    defer span.End()
    notification := Notification{
        UserID:      order.UserID,
        Type:        "order_confirmation",
        Subject:     fmt.Sprintf("Order %s confirmed", order.ID),
        OrderID:     order.ID,
        PaymentInfo: payment,
    }
    
    if err := s.notifier.Send(ctx, notification); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        fmt.Printf("Failed to send order confirmation: %v\n", err)
    } else {
        span.AddEvent("notification.sent", trace.WithAttributes(
            attribute.String("notification.type", "order_confirmation"),
            attribute.String("order.id", order.ID),
        ))
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