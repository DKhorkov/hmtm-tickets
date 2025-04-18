package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DKhorkov/hmtm-toys/api/protobuf/generated/go/toys"
	"github.com/DKhorkov/libs/pointers"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	mockclients "github.com/DKhorkov/hmtm-tickets/mocks/clients"
)

func TestToysRepository_GetAllCategories(t *testing.T) {
	ctrl := gomock.NewController(t)
	toysClient := mockclients.NewMockToysClient(ctrl)
	repo := NewToysRepository(toysClient)

	testCases := []struct {
		name               string
		setupMocks         func(toysClient *mockclients.MockToysClient)
		expectedCategories []entities.Category
		errorExpected      bool
	}{
		{
			name: "success",
			setupMocks: func(toysClient *mockclients.MockToysClient) {
				toysClient.
					EXPECT().
					GetCategories(
						gomock.Any(),
						&emptypb.Empty{},
					).
					Return(&toys.GetCategoriesOut{
						Categories: []*toys.GetCategoryOut{
							{ID: 1, Name: "Category1"},
						},
					}, nil).
					Times(1)
			},
			expectedCategories: []entities.Category{
				{ID: 1, Name: "Category1"},
			},
			errorExpected: false,
		},
		{
			name: "error",
			setupMocks: func(toysClient *mockclients.MockToysClient) {
				toysClient.
					EXPECT().
					GetCategories(
						gomock.Any(),
						&emptypb.Empty{},
					).
					Return(nil, errors.New("fetch failed")).
					Times(1)
			},
			expectedCategories: nil,
			errorExpected:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(toysClient)
			}

			categories, err := repo.GetAllCategories(context.Background())
			if tc.errorExpected {
				require.Error(t, err)
				require.Nil(t, categories)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedCategories, categories)
			}
		})
	}
}

func TestToysRepository_GetAllTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	toysClient := mockclients.NewMockToysClient(ctrl)
	repo := NewToysRepository(toysClient)

	testCases := []struct {
		name          string
		setupMocks    func(toysClient *mockclients.MockToysClient)
		expectedTags  []entities.Tag
		errorExpected bool
	}{
		{
			name: "success",
			setupMocks: func(toysClient *mockclients.MockToysClient) {
				toysClient.
					EXPECT().
					GetTags(
						gomock.Any(),
						&emptypb.Empty{},
					).
					Return(&toys.GetTagsOut{
						Tags: []*toys.GetTagOut{
							{ID: 1, Name: "Tag1"},
						},
					}, nil).
					Times(1)
			},
			expectedTags: []entities.Tag{
				{ID: 1, Name: "Tag1"},
			},
			errorExpected: false,
		},
		{
			name: "error",
			setupMocks: func(toysClient *mockclients.MockToysClient) {
				toysClient.
					EXPECT().
					GetTags(
						gomock.Any(),
						&emptypb.Empty{},
					).
					Return(nil, errors.New("fetch failed")).
					Times(1)
			},
			expectedTags:  nil,
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(toysClient)
			}

			tags, err := repo.GetAllTags(context.Background())
			if tc.errorExpected {
				require.Error(t, err)
				require.Nil(t, tags)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTags, tags)
			}
		})
	}
}

func TestToysRepository_GetMasterByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	toysClient := mockclients.NewMockToysClient(ctrl)
	repo := NewToysRepository(toysClient)

	now := time.Now().UTC().Truncate(time.Second)

	testCases := []struct {
		name           string
		userID         uint64
		setupMocks     func(toysClient *mockclients.MockToysClient)
		expectedMaster *entities.Master
		errorExpected  bool
	}{
		{
			name:   "success",
			userID: 1,
			setupMocks: func(toysClient *mockclients.MockToysClient) {
				toysClient.
					EXPECT().
					GetMasterByUser(
						gomock.Any(),
						&toys.GetMasterByUserIn{UserID: 1},
					).
					Return(&toys.GetMasterOut{
						ID:        1,
						UserID:    1,
						Info:      pointers.New("Master Info"),
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					}, nil).
					Times(1)
			},
			expectedMaster: &entities.Master{
				ID:        1,
				UserID:    1,
				Info:      pointers.New("Master Info"),
				CreatedAt: now,
				UpdatedAt: now,
			},
			errorExpected: false,
		},
		{
			name:   "error",
			userID: 1,
			setupMocks: func(toysClient *mockclients.MockToysClient) {
				toysClient.
					EXPECT().
					GetMasterByUser(
						gomock.Any(),
						&toys.GetMasterByUserIn{UserID: 1},
					).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			expectedMaster: nil,
			errorExpected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(toysClient)
			}

			master, err := repo.GetMasterByUserID(context.Background(), tc.userID)
			if tc.errorExpected {
				require.Error(t, err)
				require.Nil(t, master)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedMaster, master)
			}
		})
	}
}
