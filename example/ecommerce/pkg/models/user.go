package models

import (
	"time"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	PhoneNumber  string    `json:"phone_number"`
	Address      Address   `json:"address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  time.Time `json:"last_login_at"`
}

type UserRole struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	GrantedAt time.Time `json:"granted_at"`
	GrantedBy string    `json:"granted_by"`
}

const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"
	RoleSeller   = "seller"
)