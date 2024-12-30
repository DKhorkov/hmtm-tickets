package entities

import "time"

type Ticket struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	CategoryID  uint32    `json:"category_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float32   `json:"price"`
	Quantity    uint32    `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TagIDs      []uint32  `json:"tag_ids"`
}

type CreateTicketDTO struct {
	UserID      uint64   `json:"user_id"`
	CategoryID  uint32   `json:"category_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float32  `json:"price"`
	Quantity    uint32   `json:"quantity"`
	TagsIDs     []uint32 `json:"tag_ids"`
}
