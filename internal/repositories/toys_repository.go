package repositories

import (
	"context"

	"github.com/DKhorkov/hmtm-toys/api/protobuf/generated/go/toys"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

type ToysRepository struct {
	client interfaces.ToysClient
}

func NewToysRepository(client interfaces.ToysClient) *ToysRepository {
	return &ToysRepository{client: client}
}

func (repo *ToysRepository) GetMasterByUserID(
	ctx context.Context,
	userID uint64,
) (*entities.Master, error) {
	response, err := repo.client.GetMasterByUser(
		ctx,
		&toys.GetMasterByUserIn{
			UserID: userID,
		},
	)
	if err != nil {
		return nil, err
	}

	return &entities.Master{
		ID:        response.GetID(),
		UserID:    response.GetUserID(),
		Info:      response.Info,
		CreatedAt: response.GetCreatedAt().AsTime(),
		UpdatedAt: response.GetUpdatedAt().AsTime(),
	}, nil
}

func (repo *ToysRepository) GetAllCategories(ctx context.Context) ([]entities.Category, error) {
	response, err := repo.client.GetCategories(
		ctx,
		&emptypb.Empty{},
	)
	if err != nil {
		return nil, err
	}

	categories := make([]entities.Category, len(response.GetCategories()))
	for index, categoryResponse := range response.GetCategories() {
		categories[index] = *repo.processCategoryResponse(categoryResponse)
	}

	return categories, nil
}

func (repo *ToysRepository) GetAllTags(ctx context.Context) ([]entities.Tag, error) {
	response, err := repo.client.GetTags(
		ctx,
		&emptypb.Empty{},
	)
	if err != nil {
		return nil, err
	}

	tags := make([]entities.Tag, len(response.GetTags()))
	for index, tagResponse := range response.GetTags() {
		tags[index] = *repo.processTagResponse(tagResponse)
	}

	return tags, nil
}

func (repo *ToysRepository) processTagResponse(tagResponse *toys.GetTagOut) *entities.Tag {
	return &entities.Tag{
		ID:   tagResponse.GetID(),
		Name: tagResponse.GetName(),
	}
}

func (repo *ToysRepository) processCategoryResponse(
	categoryResponse *toys.GetCategoryOut,
) *entities.Category {
	return &entities.Category{
		ID:   categoryResponse.GetID(),
		Name: categoryResponse.GetName(),
	}
}
