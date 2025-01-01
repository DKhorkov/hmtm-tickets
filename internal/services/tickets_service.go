package services

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
	"github.com/DKhorkov/libs/logging"
)

func NewCommonTicketsService(
	ticketsRepository interfaces.TicketsRepository,
	toysRepository interfaces.ToysRepository,
	logger *slog.Logger,
) *CommonTicketsService {
	return &CommonTicketsService{
		ticketsRepository: ticketsRepository,
		toysRepository:    toysRepository,
		logger:            logger,
	}
}

type CommonTicketsService struct {
	ticketsRepository interfaces.TicketsRepository
	toysRepository    interfaces.ToysRepository
	logger            *slog.Logger
}

func (service *CommonTicketsService) CreateTicket(
	ctx context.Context,
	ticketData entities.CreateTicketDTO,
) (uint64, error) {
	if err := service.processTicketCategory(ctx, ticketData.CategoryID); err != nil {
		return 0, err
	}

	if err := service.processTicketTags(ctx, ticketData.TagIDs); err != nil {
		return 0, err
	}

	if service.checkTicketExistence(ctx, ticketData) {
		return 0, &customerrors.TicketAlreadyExistsError{}
	}

	return service.ticketsRepository.CreateTicket(ctx, ticketData)
}

func (service *CommonTicketsService) processTicketCategory(ctx context.Context, categoryID uint32) error {
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

func (service *CommonTicketsService) processTicketTags(ctx context.Context, tagIDs []uint32) error {
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

func (service *CommonTicketsService) checkTicketExistence(
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

func (service *CommonTicketsService) GetTicketByID(ctx context.Context, id uint64) (*entities.Ticket, error) {
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

func (service *CommonTicketsService) GetAllTickets(ctx context.Context) ([]entities.Ticket, error) {
	return service.ticketsRepository.GetAllTickets(ctx)
}

func (service *CommonTicketsService) GetUserTickets(ctx context.Context, userID uint64) ([]entities.Ticket, error) {
	return service.ticketsRepository.GetUserTickets(ctx, userID)
}
