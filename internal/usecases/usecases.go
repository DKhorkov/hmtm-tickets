package usecases

import (
	"context"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

func NewCommonUseCases(
	ticketsService interfaces.TicketsService,
	respondsService interfaces.RespondsService,
) *CommonUseCases {
	return &CommonUseCases{
		ticketsService:  ticketsService,
		respondsService: respondsService,
	}
}

type CommonUseCases struct {
	ticketsService  interfaces.TicketsService
	respondsService interfaces.RespondsService
}

func (useCases *CommonUseCases) CreateTicket(ctx context.Context, ticketData entities.CreateTicketDTO) (uint64, error) {
	return useCases.ticketsService.CreateTicket(ctx, ticketData)
}

func (useCases *CommonUseCases) GetTicketByID(ctx context.Context, id uint64) (*entities.Ticket, error) {
	return useCases.ticketsService.GetTicketByID(ctx, id)
}

func (useCases *CommonUseCases) GetAllTickets(ctx context.Context) ([]entities.Ticket, error) {
	return useCases.ticketsService.GetAllTickets(ctx)
}

func (useCases *CommonUseCases) GetUserTickets(ctx context.Context, userID uint64) ([]entities.Ticket, error) {
	return useCases.ticketsService.GetUserTickets(ctx, userID)
}

func (useCases *CommonUseCases) RespondToTicket(
	ctx context.Context,
	respondData entities.RawRespondToTicketDTO,
) (uint64, error) {
	ticket, err := useCases.ticketsService.GetTicketByID(ctx, respondData.TicketID)
	if err != nil {
		return 0, err
	}

	if ticket.UserID == respondData.UserID {
		return 0, &customerrors.RespondToOwnTicketError{}
	}

	return useCases.respondsService.RespondToTicket(ctx, respondData)
}

func (useCases *CommonUseCases) GetRespondByID(ctx context.Context, id uint64) (*entities.Respond, error) {
	return useCases.respondsService.GetRespondByID(ctx, id)
}

func (useCases *CommonUseCases) GetTicketResponds(ctx context.Context, ticketID uint64) ([]entities.Respond, error) {
	return useCases.respondsService.GetTicketResponds(ctx, ticketID)
}

func (useCases *CommonUseCases) GetUserResponds(ctx context.Context, userID uint64) ([]entities.Respond, error) {
	return useCases.respondsService.GetUserResponds(ctx, userID)
}
