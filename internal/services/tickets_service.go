package services

import (
	"context"
	"fmt"

	"github.com/DKhorkov/libs/logging"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

func NewTicketsService(
	ticketsRepository interfaces.TicketsRepository,
	logger logging.Logger,
) *TicketsService {
	return &TicketsService{
		ticketsRepository: ticketsRepository,
		logger:            logger,
	}
}

type TicketsService struct {
	ticketsRepository interfaces.TicketsRepository
	logger            logging.Logger
}

func (service *TicketsService) CreateTicket(
	ctx context.Context,
	ticketData entities.CreateTicketDTO,
) (uint64, error) {
	return service.ticketsRepository.CreateTicket(ctx, ticketData)
}

func (service *TicketsService) GetTicketByID(ctx context.Context, id uint64) (*entities.Ticket, error) {
	ticket, err := service.ticketsRepository.GetTicketByID(ctx, id)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			service.logger,
			fmt.Sprintf("Error occurred while trying to get Ticket with ID=%d", id),
			err,
		)

		return nil, &customerrors.TicketNotFoundError{}
	}

	return ticket, nil
}

func (service *TicketsService) GetAllTickets(ctx context.Context) ([]entities.Ticket, error) {
	return service.ticketsRepository.GetAllTickets(ctx)
}

func (service *TicketsService) GetUserTickets(ctx context.Context, userID uint64) ([]entities.Ticket, error) {
	return service.ticketsRepository.GetUserTickets(ctx, userID)
}
