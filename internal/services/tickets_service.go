package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DKhorkov/libs/logging"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

func NewTicketsService(
	ticketsRepository interfaces.TicketsRepository,
	toysRepository interfaces.ToysRepository,
	logger logging.Logger,
) *TicketsService {
	return &TicketsService{
		ticketsRepository: ticketsRepository,
		toysRepository:    toysRepository,
		logger:            logger,
	}
}

type TicketsService struct {
	ticketsRepository interfaces.TicketsRepository
	toysRepository    interfaces.ToysRepository
	logger            logging.Logger
}

func (service *TicketsService) CreateTicket(
	ctx context.Context,
	ticketData entities.CreateTicketDTO,
) (uint64, error) {
	if err := service.validateCategory(ctx, ticketData.CategoryID); err != nil {
		return 0, err
	}

	if err := service.validateTags(ctx, ticketData.TagIDs); err != nil {
		return 0, err
	}

	if service.checkTicketExistence(ctx, ticketData) {
		return 0, &customerrors.TicketAlreadyExistsError{}
	}

	return service.ticketsRepository.CreateTicket(ctx, ticketData)
}

func (service *TicketsService) validateCategory(ctx context.Context, categoryID uint32) error {
	categories, err := service.toysRepository.GetAllCategories(ctx)
	if err != nil {
		return err
	}

	for _, category := range categories {
		if category.ID == categoryID {
			return nil
		}
	}

	return &customerrors.CategoryNotFoundError{Message: strconv.FormatUint(uint64(categoryID), 10)}
}

func (service *TicketsService) validateTags(ctx context.Context, tagIDs []uint32) error {
	tags, err := service.toysRepository.GetAllTags(ctx)
	if err != nil {
		return err
	}

	tagsMap := make(map[uint32]struct{}, len(tags))
	for _, tag := range tags {
		tagsMap[tag.ID] = struct{}{}
	}

	for _, tagID := range tagIDs {
		if _, ok := tagsMap[tagID]; !ok {
			return &customerrors.TagNotFoundError{Message: strconv.FormatUint(uint64(tagID), 10)}
		}
	}

	return nil
}

func (service *TicketsService) checkTicketExistence(
	ctx context.Context,
	ticketData entities.CreateTicketDTO,
) bool {
	tickets, err := service.ticketsRepository.GetUserTickets(ctx, ticketData.UserID)
	if err == nil {
		for _, ticket := range tickets {
			if ticket.Name == ticketData.Name &&
				ticket.CategoryID == ticketData.CategoryID &&
				ticket.Description == ticketData.Description {
				return true
			}
		}
	}

	return false
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
