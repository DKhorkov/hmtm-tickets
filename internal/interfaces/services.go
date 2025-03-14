package interfaces

import (
	"context"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
)

type TicketsService interface {
	TicketsRepository
}

type RespondsService interface {
	RespondToTicket(ctx context.Context, rawRespondData entities.RawRespondToTicketDTO) (respondID uint64, err error)
	GetRespondByID(ctx context.Context, id uint64) (*entities.Respond, error)
	GetTicketResponds(ctx context.Context, ticketID uint64) ([]entities.Respond, error)
	GetUserResponds(ctx context.Context, userID uint64) ([]entities.Respond, error)
}
