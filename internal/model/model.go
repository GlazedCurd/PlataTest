package model

import "time"

type TaskId = uint64
type Code = string

const (
	STATUS_SUCCESS = "success"
	STATUS_FAILED  = "failed"
)

type Task struct {
	ID             TaskId    `json:"id,omitempty"`
	Price          *float64  `json:"price,omitempty"`
	Code           Code      `json:"code,omitempty"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	TaskdAt        time.Time `json:"updated_at,omitempty"`
	Status         string    `json:"status,omitempty"`
}
