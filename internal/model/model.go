package model

import "time"

type UpdateId = int
type Code = string

type Update struct {
	ID             UpdateId  `json:"id,omitempty"`
	Price          *float64  `json:"price,omitempty"`
	Code           Code      `json:"code,omitempty"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	Status         string    `json:"status,omitempty"`
}
