package user

import (
	"context"
	"errors"
	"sync"

	"github.com/example/ecommerce/pkg/models"
)

type Repository struct {
	mu    sync.RWMutex
	users map[string]*models.User
	byEmail map[string]*models.User
	byUsername map[string]*models.User
}

func NewRepository() *Repository {
	return &Repository{
		users:      make(map[string]*models.User),
		byEmail:    make(map[string]*models.User),
		byUsername: make(map[string]*models.User),
	}
}

func (r *Repository) Save(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	r.users[user.ID] = user
	r.byEmail[user.Email] = user
	r.byUsername[user.Username] = user

	return nil
}

func (r *Repository) Update(ctx context.Context, user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	r.users[user.ID] = user
	r.byEmail[user.Email] = user
	r.byUsername[user.Username] = user

	return nil
}

func (r *Repository) FindByID(ctx context.Context, userID string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.byEmail[email]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.byUsername[username]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}