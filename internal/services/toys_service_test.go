package services

import (
	"context"
	"errors"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
	"time"

	mockrepositories "github.com/DKhorkov/hmtm-tickets/mocks/repositories"
	mocklogging "github.com/DKhorkov/libs/logging/mocks"
)

func TestToysService_GetAllCategories(t *testing.T) {
	ctrl := gomock.NewController(t)
	toysRepository := mockrepositories.NewMockToysRepository(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	service := NewToysService(toysRepository, logger)

	testCases := []struct {
		name               string
		setupMocks         func(toysRepository *mockrepositories.MockToysRepository, logger *mocklogging.MockLogger)
		expectedCategories []entities.Category
		errorExpected      bool
	}{
		{
			name: "success",
			setupMocks: func(toysRepository *mockrepositories.MockToysRepository, logger *mocklogging.MockLogger) {
				toysRepository.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return([]entities.Category{
						{ID: 1, Name: "Category 1"},
					}, nil).
					Times(1)
			},
			expectedCategories: []entities.Category{
				{ID: 1, Name: "Category 1"},
			},
			errorExpected: false,
		},
		{
			name: "error",
			setupMocks: func(toysRepository *mockrepositories.MockToysRepository, logger *mocklogging.MockLogger) {
				toysRepository.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return(nil, errors.New("fetch failed")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedCategories: nil,
			errorExpected:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(toysRepository, logger)
			}

			categories, err := service.GetAllCategories(context.Background())
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

func TestToysService_GetAllTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	toysRepository := mockrepositories.NewMockToysRepository(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	service := NewToysService(toysRepository, logger)

	testCases := []struct {
		name          string
		setupMocks    func(toysRepository *mockrepositories.MockToysRepository, logger *mocklogging.MockLogger)
		expectedTags  []entities.Tag
		errorExpected bool
	}{
		{
			name: "success",
			setupMocks: func(toysRepository *mockrepositories.MockToysRepository, logger *mocklogging.MockLogger) {
				toysRepository.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{
						{ID: 1, Name: "tag1"},
					}, nil).
					Times(1)
			},
			expectedTags: []entities.Tag{
				{ID: 1, Name: "tag1"},
			},
			errorExpected: false,
		},
		{
			name: "error",
			setupMocks: func(toysRepository *mockrepositories.MockToysRepository, logger *mocklogging.MockLogger) {
				toysRepository.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return(nil, errors.New("fetch failed")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedTags:  nil,
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(toysRepository, logger)
			}

			tags, err := service.GetAllTags(context.Background())
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

func TestToysService_GetMasterByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	toysRepository := mockrepositories.NewMockToysRepository(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	service := NewToysService(toysRepository, logger)

	now := time.Now()
	info := "Master Info"
	testCases := []struct {
		name           string
		userID         uint64
		setupMocks     func(toysRepository *mockrepositories.MockToysRepository, logger *mocklogging.MockLogger)
		expectedMaster *entities.Master
		errorExpected  bool
	}{
		{
			name:   "success",
			userID: 1,
			setupMocks: func(toysRepository *mockrepositories.MockToysRepository, logger *mocklogging.MockLogger) {
				toysRepository.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(1)).
					Return(&entities.Master{
						ID:        1,
						UserID:    1,
						Info:      &info,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil).
					Times(1)
			},
			expectedMaster: &entities.Master{
				ID:        1,
				UserID:    1,
				Info:      &info,
				CreatedAt: now,
				UpdatedAt: now,
			},
			errorExpected: false,
		},
		{
			name:   "error",
			userID: 1,
			setupMocks: func(toysRepository *mockrepositories.MockToysRepository, logger *mocklogging.MockLogger) {
				toysRepository.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("not found")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedMaster: nil,
			errorExpected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(toysRepository, logger)
			}

			master, err := service.GetMasterByUserID(context.Background(), tc.userID)
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
