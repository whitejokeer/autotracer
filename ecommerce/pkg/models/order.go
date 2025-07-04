package models

import (
	"time"
)

type Order struct {
	ID            string      `json:"id"`
	UserID        string      `json:"user_id"`
	Items         []OrderItem `json:"items"`
	Total         float64     `json:"total"`
	PaymentMethod string      `json:"payment_method"`
	Status        string      `json:"status"`
	ShippingAddr  Address     `json:"shipping_address"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Subtotal    float64 `json:"subtotal"`
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	ZipCode    string `json:"zip_code"`
	Country    string `json:"country"`
}

const (
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusPaid       = "paid"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)