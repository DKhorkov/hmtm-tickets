package interfaces

import (
	"context"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
)

//go:generate mockgen -source=repositories.go -destination=../../mocks/repositories/tickets_repository.go -exclude_interfaces=RespondsRepository,ToysRepository -package=mockrepositories
type TicketsRepository interface {
	CreateTicket(ctx context.Context, ticketData entities.CreateTicketDTO) (ticketID uint64, err error)
	GetTicketByID(ctx context.Context, id uint64) (*entities.Ticket, error)
	GetAllTickets(ctx context.Context) ([]entities.Ticket, error)
	GetUserTickets(ctx context.Context, userID uint64) ([]entities.Ticket, error)
}

//go:generate mockgen -source=repositories.go  -destination=../../mocks/repositories/responds_repository.go -exclude_interfaces=TicketsRepository,ToysRepository -package=mockrepositories
type RespondsRepository interface {
	RespondToTicket(ctx context.Context, respondData entities.RespondToTicketDTO) (respondID uint64, err error)
	GetRespondByID(ctx context.Context, id uint64) (*entities.Respond, error)
	GetTicketResponds(ctx context.Context, ticketID uint64) ([]entities.Respond, error)
	GetMasterResponds(ctx context.Context, masterID uint64) ([]entities.Respond, error)
}

//go:generate mockgen -source=repositories.go  -destination=../../mocks/repositories/toys_repository.go -exclude_interfaces=RespondsRepository,TicketsRepository -package=mockrepositories
type ToysRepository interface {
	GetAllTags(ctx context.Context) ([]entities.Tag, error)
	GetAllCategories(ctx context.Context) ([]entities.Category, error)
	GetMasterByUserID(ctx context.Context, userID uint64) (*entities.Master, error)
}
