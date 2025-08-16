package model

type Update struct {
	ID             int    `json:"id,omitempty"`
	Price          string `json:"price,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	RequestCreated int    `json:"request_created,omitempty"`
	RequestUpdated int    `json:"request_updated,omitempty"`
}
