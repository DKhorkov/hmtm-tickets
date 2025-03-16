package services

import (
	"context"
	"fmt"

	"github.com/DKhorkov/libs/logging"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

func NewToysService(
	toysRepository interfaces.ToysRepository,
	logger logging.Logger,
) *ToysService {
	return &ToysService{
		toysRepository: toysRepository,
		logger:         logger,
	}
}

type ToysService struct {
	toysRepository interfaces.ToysRepository
	logger         logging.Logger
}

func (service *ToysService) GetAllTags(ctx context.Context) ([]entities.Tag, error) {
	tags, err := service.toysRepository.GetAllTags(ctx)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			service.logger,
			"Error occurred while trying to get all Tags",
			err,
		)
	}

	return tags, err
}

func (service *ToysService) GetAllCategories(ctx context.Context) ([]entities.Category, error) {
	categories, err := service.toysRepository.GetAllCategories(ctx)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			service.logger,
			"Error occurred while trying to get all Categories",
			err,
		)
	}

	return categories, err
}

func (service *ToysService) GetMasterByUserID(ctx context.Context, userID uint64) (*entities.Master, error) {
	master, err := service.toysRepository.GetMasterByUserID(ctx, userID)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			service.logger,
			fmt.Sprintf("Error occurred while trying to get Master for User with ID=%d", userID),
			err,
		)
	}

	return master, err
}
