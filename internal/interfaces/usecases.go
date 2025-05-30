package interfaces

import (
	"context"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
)

//go:generate mockgen -source=usecases.go -destination=../../mocks/usecases/usecases.go -package=mockusecases
type UseCases interface {
	// Tickets cases:
	CreateTicket(
		ctx context.Context,
		ticketData entities.CreateTicketDTO,
	) (ticketID uint64, err error)
	GetTicketByID(ctx context.Context, id uint64) (*entities.Ticket, error)
	GetTickets(
		ctx context.Context,
		pagination *entities.Pagination,
		filters *entities.TicketsFilters,
	) ([]entities.Ticket, error)
	CountTickets(ctx context.Context, filters *entities.TicketsFilters) (uint64, error)
	GetUserTickets(
		ctx context.Context,
		userID uint64,
		pagination *entities.Pagination,
		filters *entities.TicketsFilters,
	) ([]entities.Ticket, error)
	CountUserTickets(ctx context.Context, userID uint64, filters *entities.TicketsFilters) (uint64, error)
	DeleteTicket(ctx context.Context, id uint64) error
	UpdateTicket(ctx context.Context, rawTicketData entities.RawUpdateTicketDTO) error

	// Responds cases:
	RespondToTicket(
		ctx context.Context,
		rawRespondData entities.RawRespondToTicketDTO,
	) (respondID uint64, err error)
	GetRespondByID(ctx context.Context, id uint64) (*entities.Respond, error)
	GetTicketResponds(ctx context.Context, ticketID uint64) ([]entities.Respond, error)
	GetUserResponds(ctx context.Context, userID uint64) ([]entities.Respond, error)
	UpdateRespond(ctx context.Context, respondData entities.UpdateRespondDTO) error
	DeleteRespond(ctx context.Context, id uint64) error
}
