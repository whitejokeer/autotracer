package order

import (
	"context"
	"errors"
	"sync"

	"github.com/example/ecommerce/pkg/models"
)

type Repository struct {
	mu     sync.RWMutex
	orders map[string]*models.Order
}

func NewRepository() *Repository {
	return &Repository{
		orders: make(map[string]*models.Order),
	}
}

func (r *Repository) Save(ctx context.Context, order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; exists {
		return errors.New("order already exists")
	}

	r.orders[order.ID] = order
	return nil
}

func (r *Repository) Update(ctx context.Context, order *models.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; !exists {
		return errors.New("order not found")
	}

	r.orders[order.ID] = order
	return nil
}

func (r *Repository) FindByID(ctx context.Context, orderID string) (*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}

	return order, nil
}

func (r *Repository) FindByUserID(ctx context.Context, userID string) ([]*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userOrders []*models.Order
	for _, order := range r.orders {
		if order.UserID == userID {
			userOrders = append(userOrders, order)
		}
	}

	return userOrders, nil
}