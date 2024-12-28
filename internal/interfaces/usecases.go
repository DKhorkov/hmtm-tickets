package interfaces

import (
	"context"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
)

type UseCases interface {
	CreateTicket(ctx context.Context, ticketData entities.RawCreateTicketDTO) (ticketID uint64, err error)
	GetTicketByID(ctx context.Context, id uint64) (*entities.Ticket, error)
	GetAllTickets(ctx context.Context) ([]entities.Ticket, error)
	GetMyTickets(ctx context.Context, accessToken string) ([]entities.Ticket, error)
	RespondToTicket(ctx context.Context, respondData entities.RawRespondToTicketDTO) (respondID uint64, err error)
	GetRespondByID(ctx context.Context, id uint64) (*entities.Respond, error)
	GetTicketResponds(ctx context.Context, ticketID uint64) ([]entities.Respond, error)
	GetMyResponds(ctx context.Context, accessToken string) ([]entities.Respond, error)
}
