package entities

import "time"

type Respond struct {
	ID        uint64    `json:"id"`
	TicketID  uint64    `json:"ticketId"`
	MasterID  uint64    `json:"masterId"`
	Price     float32   `json:"price"`
	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type RespondToTicketDTO struct {
	TicketID uint64  `json:"ticketId"`
	MasterID uint64  `json:"masterId"`
	Price    float32 `json:"price"`
	Comment  *string `json:"comment,omitempty"`
}

type RawRespondToTicketDTO struct {
	TicketID uint64  `json:"ticketId"`
	UserID   uint64  `json:"userId"`
	Price    float32 `json:"price"`
	Comment  *string `json:"comment,omitempty"`
}

type UpdateRespondDTO struct {
	ID      uint64   `json:"id"`
	Price   *float32 `json:"price,omitempty"`
	Comment *string  `json:"comment,omitempty"`
}
