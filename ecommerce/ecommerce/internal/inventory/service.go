package inventory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/example/ecommerce/pkg/models"
)

type Service struct {
	mu       sync.RWMutex
	products map[string]*models.Product
	stock    map[string]int
	reserved map[string]int
}

func NewService() *Service {
	svc := &Service{
		products: make(map[string]*models.Product),
		stock:    make(map[string]int),
		reserved: make(map[string]int),
	}
	
	svc.seedProducts()
	return svc
}

func (s *Service) CheckAvailability(ctx context.Context, productID string, quantity int) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currentStock, exists := s.stock[productID]
	if !exists {
		return false, fmt.Errorf("product %s not found", productID)
	}

	reserved := s.reserved[productID]
	available := currentStock - reserved

	return available >= quantity, nil
}

func (s *Service) ReserveStock(ctx context.Context, productID string, quantity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentStock, exists := s.stock[productID]
	if !exists {
		return fmt.Errorf("product %s not found", productID)
	}

	reserved := s.reserved[productID]
	available := currentStock - reserved

	if available < quantity {
		return errors.New("insufficient stock")
	}

	s.reserved[productID] += quantity

	go s.releaseReservationAfterTimeout(productID, quantity)

	return nil
}

func (s *Service) DeductStock(ctx context.Context, productID string, quantity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentStock, exists := s.stock[productID]
	if !exists {
		return fmt.Errorf("product %s not found", productID)
	}

	if currentStock < quantity {
		return errors.New("insufficient stock")
	}

	s.stock[productID] -= quantity
	
	if s.reserved[productID] >= quantity {
		s.reserved[productID] -= quantity
	}

	go s.checkLowStockAlert(productID, s.stock[productID])

	return nil
}

func (s *Service) GetProduct(ctx context.Context, productID string) (*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, exists := s.products[productID]
	if !exists {
		return nil, fmt.Errorf("product %s not found", productID)
	}

	return product, nil
}

func (s *Service) UpdateStock(ctx context.Context, productID string, newStock int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[productID]; !exists {
		return fmt.Errorf("product %s not found", productID)
	}

	if newStock < 0 {
		return errors.New("stock cannot be negative")
	}

	oldStock := s.stock[productID]
	s.stock[productID] = newStock

	if oldStock == 0 && newStock > 0 {
		go s.notifyBackInStock(productID)
	}

	return nil
}

func (s *Service) releaseReservationAfterTimeout(productID string, quantity int) {
	time.Sleep(15 * time.Minute)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.reserved[productID] >= quantity {
		s.reserved[productID] -= quantity
	}
}

func (s *Service) checkLowStockAlert(productID string, currentStock int) {
	if currentStock < 10 {
		fmt.Printf("LOW STOCK ALERT: Product %s has only %d items left\n", productID, currentStock)
	}
}

func (s *Service) notifyBackInStock(productID string) {
	fmt.Printf("Product %s is back in stock!\n", productID)
}

func (s *Service) seedProducts() {
	products := []*models.Product{
		{
			ID:          "PROD-001",
			Name:        "Laptop Pro 15",
			Description: "High-performance laptop",
			Price:       1299.99,
			SKU:         "LPT-PRO-15",
			Category:    "Electronics",
			Stock:       50,
			Active:      true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "PROD-002",
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse",
			Price:       49.99,
			SKU:         "MS-WL-001",
			Category:    "Accessories",
			Stock:       150,
			Active:      true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "PROD-003",
			Name:        "USB-C Hub",
			Description: "7-in-1 USB-C Hub",
			Price:       79.99,
			SKU:         "HUB-C-7IN1",
			Category:    "Accessories",
			Stock:       75,
			Active:      true,
			CreatedAt:   time.Now(),
		},
	}

	for _, product := range products {
		s.products[product.ID] = product
		s.stock[product.ID] = product.Stock
		s.reserved[product.ID] = 0
	}
}