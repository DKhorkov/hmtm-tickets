package usecases

import (
	"context"
	"strconv"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

func New(
	ticketsService interfaces.TicketsService,
	respondsService interfaces.RespondsService,
	toysService interfaces.ToysService,
) *UseCases {
	return &UseCases{
		ticketsService:  ticketsService,
		respondsService: respondsService,
		toysService:     toysService,
	}
}

type UseCases struct {
	ticketsService  interfaces.TicketsService
	respondsService interfaces.RespondsService
	toysService     interfaces.ToysService
}

func (useCases *UseCases) CreateTicket(ctx context.Context, ticketData entities.CreateTicketDTO) (uint64, error) {
	if err := useCases.validateCategory(ctx, ticketData.CategoryID); err != nil {
		return 0, err
	}

	if err := useCases.validateTags(ctx, ticketData.TagIDs); err != nil {
		return 0, err
	}

	if useCases.checkTicketExistence(ctx, ticketData) {
		return 0, &customerrors.TicketAlreadyExistsError{}
	}

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
	rawRespondData entities.RawRespondToTicketDTO,
) (uint64, error) {
	ticket, err := useCases.ticketsService.GetTicketByID(ctx, rawRespondData.TicketID)
	if err != nil {
		return 0, err
	}

	if ticket.UserID == rawRespondData.UserID {
		return 0, &customerrors.RespondToOwnTicketError{}
	}

	master, err := useCases.toysService.GetMasterByUserID(ctx, rawRespondData.UserID)
	if err != nil {
		return 0, err
	}

	respondData := entities.RespondToTicketDTO{
		MasterID: master.ID,
		TicketID: rawRespondData.TicketID,
		Price:    rawRespondData.Price,
		Comment:  rawRespondData.Comment,
	}

	if useCases.checkRespondExistence(ctx, respondData) {
		return 0, &customerrors.RespondAlreadyExistsError{}
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
	master, err := useCases.toysService.GetMasterByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return useCases.respondsService.GetMasterResponds(ctx, master.ID)
}

func (useCases *UseCases) validateCategory(ctx context.Context, categoryID uint32) error {
	categories, err := useCases.toysService.GetAllCategories(ctx)
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

func (useCases *UseCases) validateTags(ctx context.Context, tagIDs []uint32) error {
	tags, err := useCases.toysService.GetAllTags(ctx)
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

func (useCases *UseCases) checkTicketExistence(
	ctx context.Context,
	ticketData entities.CreateTicketDTO,
) bool {
	tickets, err := useCases.ticketsService.GetUserTickets(ctx, ticketData.UserID)
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

func (useCases *UseCases) checkRespondExistence(ctx context.Context, respondData entities.RespondToTicketDTO) bool {
	responds, err := useCases.respondsService.GetMasterResponds(ctx, respondData.MasterID)
	if err == nil {
		for _, respond := range responds {
			if respond.TicketID == respondData.TicketID {
				return true
			}
		}
	}

	return false
}

func (useCases *UseCases) UpdateRespond(ctx context.Context, respondData entities.UpdateRespondDTO) error {
	if _, err := useCases.respondsService.GetRespondByID(ctx, respondData.ID); err != nil {
		return err
	}

	return useCases.respondsService.UpdateRespond(ctx, respondData)
}

func (useCases *UseCases) DeleteRespond(ctx context.Context, id uint64) error {
	if _, err := useCases.respondsService.GetRespondByID(ctx, id); err != nil {
		return err
	}

	return useCases.respondsService.DeleteRespond(ctx, id)
}
