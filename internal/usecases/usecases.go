package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/DKhorkov/libs/logging"

	notifications "github.com/DKhorkov/hmtm-notifications/dto"
	customnats "github.com/DKhorkov/libs/nats"

	"github.com/DKhorkov/hmtm-tickets/internal/config"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

func New(
	ticketsService interfaces.TicketsService,
	respondsService interfaces.RespondsService,
	toysService interfaces.ToysService,
	natsPublisher customnats.Publisher,
	natsConfig config.NATSConfig,
	logger logging.Logger,
) *UseCases {
	return &UseCases{
		ticketsService:  ticketsService,
		respondsService: respondsService,
		toysService:     toysService,
		natsPublisher:   natsPublisher,
		natsConfig:      natsConfig,
		logger:          logger,
	}
}

type UseCases struct {
	ticketsService  interfaces.TicketsService
	respondsService interfaces.RespondsService
	toysService     interfaces.ToysService
	natsPublisher   customnats.Publisher
	natsConfig      config.NATSConfig
	logger          logging.Logger
}

func (useCases *UseCases) CreateTicket(
	ctx context.Context,
	ticketData entities.CreateTicketDTO,
) (uint64, error) {
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

func (useCases *UseCases) GetUserTickets(
	ctx context.Context,
	userID uint64,
) ([]entities.Ticket, error) {
	return useCases.ticketsService.GetUserTickets(ctx, userID)
}

func (useCases *UseCases) RespondToTicket(
	ctx context.Context,
	rawRespondData entities.RawRespondToTicketDTO,
) (uint64, error) {
	ticket, err := useCases.GetTicketByID(ctx, rawRespondData.TicketID)
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

func (useCases *UseCases) GetRespondByID(
	ctx context.Context,
	id uint64,
) (*entities.Respond, error) {
	return useCases.respondsService.GetRespondByID(ctx, id)
}

func (useCases *UseCases) GetTicketResponds(
	ctx context.Context,
	ticketID uint64,
) ([]entities.Respond, error) {
	return useCases.respondsService.GetTicketResponds(ctx, ticketID)
}

func (useCases *UseCases) GetUserResponds(
	ctx context.Context,
	userID uint64,
) ([]entities.Respond, error) {
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
	tickets, err := useCases.GetUserTickets(ctx, ticketData.UserID)
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

func (useCases *UseCases) checkRespondExistence(
	ctx context.Context,
	respondData entities.RespondToTicketDTO,
) bool {
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

func (useCases *UseCases) UpdateRespond(
	ctx context.Context,
	respondData entities.UpdateRespondDTO,
) error {
	if _, err := useCases.GetRespondByID(ctx, respondData.ID); err != nil {
		return err
	}

	return useCases.respondsService.UpdateRespond(ctx, respondData)
}

func (useCases *UseCases) DeleteRespond(ctx context.Context, id uint64) error {
	if _, err := useCases.GetRespondByID(ctx, id); err != nil {
		return err
	}

	return useCases.respondsService.DeleteRespond(ctx, id)
}

func (useCases *UseCases) DeleteTicket(ctx context.Context, id uint64) error {
	ticket, err := useCases.GetTicketByID(ctx, id)
	if err != nil {
		return err
	}

	ticketResponds, err := useCases.GetTicketResponds(ctx, ticket.ID)
	if err != nil {
		return err
	}

	if err = useCases.ticketsService.DeleteTicket(ctx, id); err != nil {
		return err
	}

	respondedMastersIDs := make([]uint64, 0, len(ticketResponds))
	for _, respond := range ticketResponds {
		respondedMastersIDs = append(respondedMastersIDs, respond.MasterID)
	}

	deleteTicketDTO := &notifications.DeleteTicketDTO{
		TicketOwnerID:       ticket.UserID,
		Name:                ticket.Name,
		Description:         ticket.Description,
		Price:               ticket.Price,
		Quantity:            ticket.Quantity,
		RespondedMastersIDs: respondedMastersIDs,
	}

	// Not returning error (if exists) from sending message, because it is not main logic and a newsletter:
	content, err := json.Marshal(deleteTicketDTO)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			useCases.logger,
			fmt.Sprintf(
				"Error occurred while trying to encode data for "+
					"delete ticket email for Ticket with Name=%s, Description=%s, Quantity=%d",
				ticket.Name,
				ticket.Description,
				ticket.Quantity,
			),
			err,
		)
	}

	if err = useCases.natsPublisher.Publish(useCases.natsConfig.Subjects.DeleteTicket, content); err != nil {
		logging.LogErrorContext(
			ctx,
			useCases.logger,
			fmt.Sprintf(
				"Error occurred while trying to send delete ticket email for "+
					"Ticket with Name=%s, Description=%s, Quantity=%d",
				ticket.Name,
				ticket.Description,
				ticket.Quantity,
			),
			err,
		)
	}

	return nil
}

func (useCases *UseCases) UpdateTicket(
	ctx context.Context,
	rawTicketData entities.RawUpdateTicketDTO,
) error {
	ticket, err := useCases.GetTicketByID(ctx, rawTicketData.ID)
	if err != nil {
		return err
	}

	if rawTicketData.CategoryID != nil {
		if err = useCases.validateCategory(ctx, *rawTicketData.CategoryID); err != nil {
			return err
		}
	}

	if err = useCases.validateTags(ctx, rawTicketData.TagIDs); err != nil {
		return err
	}

	// Old Ticket Tags IDs set:
	oldTagIDsSet := make(map[uint32]struct{}, len(ticket.TagIDs))
	for _, tagID := range ticket.TagIDs {
		oldTagIDsSet[tagID] = struct{}{}
	}

	// New Ticket Tags IDs set:
	newTagIDsSet := make(map[uint32]struct{}, len(rawTicketData.TagIDs))
	for _, tagID := range rawTicketData.TagIDs {
		newTagIDsSet[tagID] = struct{}{}
	}

	// Add new Tag if it is not already exists:
	tagIDsToAdd := make([]uint32, 0)
	for _, tagID := range rawTicketData.TagIDs {
		if _, ok := oldTagIDsSet[tagID]; !ok {
			tagIDsToAdd = append(tagIDsToAdd, tagID)
		}
	}

	// Delete old Tag if it is not used by Ticket now:
	tagIDsToDelete := make([]uint32, 0)
	for _, tagID := range ticket.TagIDs {
		if _, ok := newTagIDsSet[tagID]; !ok {
			tagIDsToDelete = append(tagIDsToDelete, tagID)
		}
	}

	// Old Ticket Attachments set:
	oldAttachmentsSet := make(map[string]struct{}, len(ticket.Attachments))
	for _, attachment := range ticket.Attachments {
		oldAttachmentsSet[attachment.Link] = struct{}{}
	}

	// New Ticket Attachments set:
	newAttachmentsSet := make(map[string]struct{}, len(rawTicketData.Attachments))
	for _, attachment := range rawTicketData.Attachments {
		newAttachmentsSet[attachment] = struct{}{}
	}

	// Add new Attachments if it is not already exists:
	attachmentsToAdd := make([]string, 0)
	for _, attachment := range rawTicketData.Attachments {
		if _, ok := oldAttachmentsSet[attachment]; !ok {
			attachmentsToAdd = append(attachmentsToAdd, attachment)
		}
	}

	// Delete old Attachments if it is not used by Ticket now:
	attachmentsToDelete := make([]uint64, 0)
	for _, attachment := range ticket.Attachments {
		if _, ok := newAttachmentsSet[attachment.Link]; !ok {
			attachmentsToDelete = append(attachmentsToDelete, attachment.ID)
		}
	}

	ticketData := entities.UpdateTicketDTO{
		ID:                    rawTicketData.ID,
		CategoryID:            rawTicketData.CategoryID,
		Name:                  rawTicketData.Name,
		Description:           rawTicketData.Description,
		Price:                 rawTicketData.Price,
		Quantity:              rawTicketData.Quantity,
		TagIDsToAdd:           tagIDsToAdd,
		TagIDsToDelete:        tagIDsToDelete,
		AttachmentsToAdd:      attachmentsToAdd,
		AttachmentIDsToDelete: attachmentsToDelete,
	}

	if err = useCases.ticketsService.UpdateTicket(ctx, ticketData); err != nil {
		return err
	}

	updateTicketDTO := &notifications.UpdateTicketDTO{
		TicketID: ticket.ID,
	}

	// Not returning error (if exists) from sending message, because it is not main logic and a newsletter:
	content, err := json.Marshal(updateTicketDTO)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			useCases.logger,
			fmt.Sprintf(
				"Error occurred while trying to encode data for update ticket email for Ticket with ID=%d",
				ticket.ID,
			),
			err,
		)
	}

	if err = useCases.natsPublisher.Publish(useCases.natsConfig.Subjects.UpdateTicket, content); err != nil {
		logging.LogErrorContext(
			ctx,
			useCases.logger,
			fmt.Sprintf(
				"Error occurred while trying send update ticket message for Ticket with ID=%d",
				ticket.ID,
			),
			err,
		)
	}

	return nil
}
