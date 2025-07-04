package notification

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Service struct {
	emailClient EmailClient
	smsClient   SMSClient
	queue       chan *Message
}

type Message struct {
	ID        string
	UserID    string
	Type      string
	Channel   string
	Subject   string
	Content   string
	Data      map[string]interface{}
	Timestamp time.Time
}

func NewService() *Service {
	svc := &Service{
		emailClient: NewEmailClient(),
		smsClient:   NewSMSClient(),
		queue:       make(chan *Message, 1000),
	}

	go svc.processQueue()

	return svc
}

func (s *Service) Send(ctx context.Context, msg *Message) error {
	if err := s.validateMessage(msg); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	msg.ID = s.generateMessageID()
	msg.Timestamp = time.Now()

	if msg.Channel == "" {
		msg.Channel = "email"
	}

	select {
	case s.queue <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errors.New("notification queue is full")
	}
}

func (s *Service) SendOrderConfirmation(ctx context.Context, userID, orderID string, amount float64) error {
	msg := &Message{
		UserID:  userID,
		Type:    "order_confirmation",
		Subject: fmt.Sprintf("Order #%s Confirmed", orderID),
		Content: fmt.Sprintf("Your order has been confirmed. Total amount: $%.2f", amount),
		Data: map[string]interface{}{
			"order_id": orderID,
			"amount":   amount,
		},
	}

	return s.Send(ctx, msg)
}

func (s *Service) SendPaymentReceipt(ctx context.Context, userID, transactionID string, amount float64) error {
	msg := &Message{
		UserID:  userID,
		Type:    "payment_receipt",
		Subject: "Payment Receipt",
		Content: fmt.Sprintf("Payment of $%.2f has been processed successfully.", amount),
		Data: map[string]interface{}{
			"transaction_id": transactionID,
			"amount":         amount,
		},
	}

	return s.Send(ctx, msg)
}

func (s *Service) processQueue() {
	for msg := range s.queue {
		s.deliverMessage(msg)
	}
}

func (s *Service) deliverMessage(msg *Message) {
	switch msg.Channel {
	case "email":
		if err := s.emailClient.Send(msg); err != nil {
			s.handleDeliveryFailure(msg, err)
		}
	case "sms":
		if err := s.smsClient.Send(msg); err != nil {
			s.handleDeliveryFailure(msg, err)
		}
	default:
		fmt.Printf("Unknown channel: %s\n", msg.Channel)
	}
}

func (s *Service) handleDeliveryFailure(msg *Message, err error) {
	fmt.Printf("Failed to deliver %s notification to user %s: %v\n", 
		msg.Type, msg.UserID, err)
	
	go s.retryDelivery(msg, 1)
}

func (s *Service) retryDelivery(msg *Message, attempt int) {
	if attempt > 3 {
		fmt.Printf("Max retries exceeded for message %s\n", msg.ID)
		return
	}

	backoff := time.Duration(attempt*attempt) * time.Second
	time.Sleep(backoff)

	s.deliverMessage(msg)
}

func (s *Service) validateMessage(msg *Message) error {
	if msg.UserID == "" {
		return errors.New("user ID is required")
	}

	if msg.Type == "" {
		return errors.New("message type is required")
	}

	if msg.Subject == "" && msg.Content == "" {
		return errors.New("message must have subject or content")
	}

	return nil
}

func (s *Service) generateMessageID() string {
	return fmt.Sprintf("MSG-%d-%04d", time.Now().Unix(), time.Now().Nanosecond()%10000)
}

type EmailClient interface {
	Send(msg *Message) error
}

type SMSClient interface {
	Send(msg *Message) error
}

type emailClient struct{}

func NewEmailClient() EmailClient {
	return &emailClient{}
}

func (e *emailClient) Send(msg *Message) error {
	fmt.Printf("📧 Sending email to user %s: %s\n", msg.UserID, msg.Subject)
	time.Sleep(100 * time.Millisecond)
	return nil
}

type smsClient struct{}

func NewSMSClient() SMSClient {
	return &smsClient{}
}

func (s *smsClient) Send(msg *Message) error {
	fmt.Printf("📱 Sending SMS to user %s: %s\n", msg.UserID, msg.Content)
	time.Sleep(200 * time.Millisecond)
	return nil
}