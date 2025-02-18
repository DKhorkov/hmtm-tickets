package usecases

import (
	"context"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

func New(
	ticketsService interfaces.TicketsService,
	respondsService interfaces.RespondsService,
) *UseCases {
	return &UseCases{
		ticketsService:  ticketsService,
		respondsService: respondsService,
	}
}

type UseCases struct {
	ticketsService  interfaces.TicketsService
	respondsService interfaces.RespondsService
}

func (useCases *UseCases) CreateTicket(ctx context.Context, ticketData entities.CreateTicketDTO) (uint64, error) {
	return useCases.ticketsService.CreateTicket(ctx, ticketData)
}

func (useCases *UseCases) GetTicketByID(ctx context.Context, id uint64) (*entities.Ticket, error) {
	return useCases.ticketsService.GetTicketByID(ctx, id)
}

func (useCases *UseCases) GetAllTickets(ctx context.Context) ([]entities.Ticket, error) {
	return useCases.ticketsService.GetAllTickets(ctx)
}

func (useCases *UseCases) GetUserTickets(ctx context.Context, userID uint64) ([]entities.Ticket, error) {
	return useCases.ticketsService.GetUserTickets(ctx, userID)
}

func (useCases *UseCases) RespondToTicket(
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

func (useCases *UseCases) GetRespondByID(ctx context.Context, id uint64) (*entities.Respond, error) {
	return useCases.respondsService.GetRespondByID(ctx, id)
}

func (useCases *UseCases) GetTicketResponds(ctx context.Context, ticketID uint64) ([]entities.Respond, error) {
	return useCases.respondsService.GetTicketResponds(ctx, ticketID)
}

func (useCases *UseCases) GetUserResponds(ctx context.Context, userID uint64) ([]entities.Respond, error) {
	return useCases.respondsService.GetUserResponds(ctx, userID)
}
