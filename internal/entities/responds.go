package entities

import "time"

type Respond struct {
	ID        uint64    `json:"id"`
	TicketID  uint64    `json:"ticket_id"`
	MasterID  uint64    `json:"master_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RespondToTicketDTO struct {
	MasterID uint64 `json:"master_id"`
	TicketID uint64 `json:"ticket_id"`
}

type RawRespondToTicketDTO struct {
	AccessToken string `json:"access_token"`
	TicketID    uint64 `json:"ticket_id"`
}
