package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DKhorkov/libs/logging"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

func NewRespondsService(
	respondsRepository interfaces.RespondsRepository,
	toysRepository interfaces.ToysRepository,
	logger *slog.Logger,
) *RespondsService {
	return &RespondsService{
		respondsRepository: respondsRepository,
		toysRepository:     toysRepository,
		logger:             logger,
	}
}

type RespondsService struct {
	respondsRepository interfaces.RespondsRepository
	toysRepository     interfaces.ToysRepository
	logger             *slog.Logger
}

func (service *RespondsService) RespondToTicket(
	ctx context.Context,
	respondData entities.RawRespondToTicketDTO,
) (uint64, error) {
	master, err := service.toysRepository.GetMasterByUserID(ctx, respondData.UserID)
	if err != nil {
		return 0, err
	}

	processedRespondData := entities.RespondToTicketDTO{
		MasterID: master.ID,
		TicketID: respondData.TicketID,
	}

	if service.checkRespondExistence(ctx, processedRespondData) {
		return 0, &customerrors.RespondAlreadyExistsError{}
	}

	return service.respondsRepository.RespondToTicket(ctx, processedRespondData)
}

func (service *RespondsService) checkRespondExistence(
	ctx context.Context,
	respondData entities.RespondToTicketDTO,
) bool {
	responds, err := service.respondsRepository.GetMasterResponds(ctx, respondData.MasterID)
	if err == nil {
		for _, respond := range responds {
			if respond.TicketID == respondData.TicketID {
				return true
			}
		}
	}

	return false
}

func (service *RespondsService) GetRespondByID(ctx context.Context, id uint64) (*entities.Respond, error) {
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

func (service *RespondsService) GetUserResponds(
	ctx context.Context,
	userID uint64,
) ([]entities.Respond, error) {
	master, err := service.toysRepository.GetMasterByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return service.respondsRepository.GetMasterResponds(ctx, master.ID)
}
