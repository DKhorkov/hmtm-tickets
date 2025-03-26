package services

import (
	"context"
	"fmt"

	"github.com/DKhorkov/libs/logging"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

func NewRespondsService(
	respondsRepository interfaces.RespondsRepository,
	logger logging.Logger,
) *RespondsService {
	return &RespondsService{
		respondsRepository: respondsRepository,
		logger:             logger,
	}
}

type RespondsService struct {
	respondsRepository interfaces.RespondsRepository
	logger             logging.Logger
}

func (service *RespondsService) RespondToTicket(
	ctx context.Context,
	respondData entities.RespondToTicketDTO,
) (uint64, error) {
	return service.respondsRepository.RespondToTicket(ctx, respondData)
}

func (service *RespondsService) GetRespondByID(
	ctx context.Context,
	id uint64,
) (*entities.Respond, error) {
	respond, err := service.respondsRepository.GetRespondByID(ctx, id)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			service.logger,
			fmt.Sprintf("Error occurred while trying to get Respond with ID=%d", id),
			err,
		)

		return nil, &customerrors.RespondNotFoundError{}
	}

	return respond, nil
}

func (service *RespondsService) GetTicketResponds(
	ctx context.Context,
	ticketID uint64,
) ([]entities.Respond, error) {
	return service.respondsRepository.GetTicketResponds(ctx, ticketID)
}

func (service *RespondsService) GetMasterResponds(
	ctx context.Context,
	masterID uint64,
) ([]entities.Respond, error) {
	return service.respondsRepository.GetMasterResponds(ctx, masterID)
}

func (service *RespondsService) UpdateRespond(
	ctx context.Context,
	respondData entities.UpdateRespondDTO,
) error {
	return service.respondsRepository.UpdateRespond(ctx, respondData)
}

func (service *RespondsService) DeleteRespond(ctx context.Context, id uint64) error {
	return service.respondsRepository.DeleteRespond(ctx, id)
}
