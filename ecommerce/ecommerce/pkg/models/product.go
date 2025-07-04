package models

import (
	"time"
)

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	SKU         string    `json:"sku"`
	Category    string    `json:"category"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"image_url"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}