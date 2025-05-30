package services

import (
	"context"
	"fmt"

	"github.com/DKhorkov/libs/logging"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

type TicketsService struct {
	ticketsRepository interfaces.TicketsRepository
	logger            logging.Logger
}

func NewTicketsService(
	ticketsRepository interfaces.TicketsRepository,
	logger logging.Logger,
) *TicketsService {
	return &TicketsService{
		ticketsRepository: ticketsRepository,
		logger:            logger,
	}
}

func (service *TicketsService) CreateTicket(
	ctx context.Context,
	ticketData entities.CreateTicketDTO,
) (uint64, error) {
	return service.ticketsRepository.CreateTicket(ctx, ticketData)
}

func (service *TicketsService) GetTicketByID(
	ctx context.Context,
	id uint64,
) (*entities.Ticket, error) {
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

func (service *TicketsService) GetTickets(
	ctx context.Context,
	pagination *entities.Pagination,
	filters *entities.TicketsFilters,
) ([]entities.Ticket, error) {
	return service.ticketsRepository.GetTickets(ctx, pagination, filters)
}

func (service *TicketsService) CountTickets(ctx context.Context, filters *entities.TicketsFilters) (uint64, error) {
	return service.ticketsRepository.CountTickets(ctx, filters)
}

func (service *TicketsService) GetUserTickets(
	ctx context.Context,
	userID uint64,
	pagination *entities.Pagination,
	filters *entities.TicketsFilters,
) ([]entities.Ticket, error) {
	return service.ticketsRepository.GetUserTickets(ctx, userID, pagination, filters)
}

func (service *TicketsService) CountUserTickets(
	ctx context.Context,
	userID uint64,
	filters *entities.TicketsFilters,
) (uint64, error) {
	return service.ticketsRepository.CountUserTickets(ctx, userID, filters)
}

func (service *TicketsService) DeleteTicket(ctx context.Context, id uint64) error {
	return service.ticketsRepository.DeleteTicket(ctx, id)
}

func (service *TicketsService) UpdateTicket(
	ctx context.Context,
	ticketData entities.UpdateTicketDTO,
) error {
	return service.ticketsRepository.UpdateTicket(ctx, ticketData)
}
