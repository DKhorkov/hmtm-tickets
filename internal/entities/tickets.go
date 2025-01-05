package entities

import "time"

type Ticket struct {
	ID          uint64       `json:"id"`
	UserID      uint64       `json:"user_id"`
	CategoryID  uint32       `json:"category_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       float32      `json:"price"`
	Quantity    uint32       `json:"quantity"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	TagIDs      []uint32     `json:"tag_ids,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	ID        uint64    `json:"id"`
	TicketID  uint64    `json:"ticket_id"`
	Link      string    `json:"link"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTicketDTO struct {
	UserID      uint64   `json:"user_id"`
	CategoryID  uint32   `json:"category_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float32  `json:"price"`
	Quantity    uint32   `json:"quantity"`
	TagIDs      []uint32 `json:"tag_ids,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}
