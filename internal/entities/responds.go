package entities

import "time"

type Respond struct {
	ID        uint64    `json:"id"`
	TicketID  uint64    `json:"ticket_id"`
	MasterID  uint64    `json:"master_id"`
	Price     float32   `json:"price"`
	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RespondToTicketDTO struct {
	TicketID uint64  `json:"ticket_id"`
	MasterID uint64  `json:"master_id"`
	Price    float32 `json:"price"`
	Comment  *string `json:"comment,omitempty"`
}

type RawRespondToTicketDTO struct {
	TicketID uint64  `json:"ticket_id"`
	UserID   uint64  `json:"user_id"`
	Price    float32 `json:"price"`
	Comment  *string `json:"comment,omitempty"`
}

type UpdateRespondDTO struct {
	ID      uint64   `json:"id"`
	Price   *float32 `json:"price,omitempty"`
	Comment *string  `json:"comment,omitempty"`
}
