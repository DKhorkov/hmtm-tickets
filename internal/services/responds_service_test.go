package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	loggerMock "github.com/DKhorkov/libs/logging/mocks"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/services"
	mockrepositories "github.com/DKhorkov/hmtm-tickets/mocks/repositories"
)

const (
	userID    uint64 = 1
	ticketID  uint64 = 1
	masterID  uint64 = 1
	respondID uint64 = 1
)

var (
	master = &entities.Master{
		ID:     masterID,
		UserID: userID,
	}
	respondToTicketDTO = entities.RespondToTicketDTO{
		TicketID: ticketID,
		MasterID: masterID,
	}
	rawRespondToTicketDTO = entities.RawRespondToTicketDTO{
		TicketID: ticketID,
		UserID:   userID,
	}
	respond = &entities.Respond{
		TicketID: ticketID,
		MasterID: masterID,
		ID:       respondID,
	}
	errMasterNotFound = errors.New("master not found")
)

func TestRespondsService_RespondToTicket(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(
			respondsRepository *mockrepositories.MockRespondsRepository,
			toysRepository *mockrepositories.MockToysRepository,
			logger *loggerMock.MockLogger,
		)
		rawRespondToTicketDTO entities.RawRespondToTicketDTO
		expected              uint64
		errorExpected         bool
		err                   error
	}{
		{
			name: "successfully responded to ticket",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				toysRepository *mockrepositories.MockToysRepository,
				_ *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					RespondToTicket(gomock.Any(), respondToTicketDTO).
					Return(respondID, nil).
					MaxTimes(1)

				respondsRepository.
					EXPECT().
					GetMasterResponds(gomock.Any(), masterID).
					Return(nil, nil).
					MaxTimes(1)

				toysRepository.
					EXPECT().
					GetMasterByUserID(gomock.Any(), userID).
					Return(master, nil).
					MaxTimes(2)
			},
			rawRespondToTicketDTO: rawRespondToTicketDTO,
			expected:              respondID,
			errorExpected:         false,
		},
		{
			name: "failed to respond to ticket respond already exists",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				toysRepository *mockrepositories.MockToysRepository,
				_ *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					RespondToTicket(gomock.Any(), entities.RespondToTicketDTO{TicketID: ticketID, MasterID: 2}).
					Return(uint64(0), &customerrors.RespondAlreadyExistsError{}).
					MaxTimes(1)

				toysRepository.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(2)).
					Return(nil, errMasterNotFound).
					MaxTimes(1)
			},
			rawRespondToTicketDTO: entities.RawRespondToTicketDTO{TicketID: ticketID, UserID: 2},
			errorExpected:         true,
			err:                   errMasterNotFound,
		},
		{
			name: "failed to respond to ticket master not found",
			setupMocks: func(
				_ *mockrepositories.MockRespondsRepository,
				toysRepository *mockrepositories.MockToysRepository,
				_ *loggerMock.MockLogger,
			) {
				toysRepository.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(2)).
					Return(&entities.Master{ID: 2, UserID: 2}, nil).
					MaxTimes(1)

				toysRepository.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(3)).
					Return(nil, errMasterNotFound).
					MaxTimes(1)
			},
			rawRespondToTicketDTO: entities.RawRespondToTicketDTO{TicketID: ticketID, UserID: 3},
			errorExpected:         true,
			err:                   errMasterNotFound,
		},
	}

	mockController := gomock.NewController(t)
	logger := loggerMock.NewMockLogger(mockController)
	respondsRepository := mockrepositories.NewMockRespondsRepository(mockController)
	toysRepository := mockrepositories.NewMockToysRepository(mockController)
	respondsService := services.NewRespondsService(respondsRepository, toysRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(respondsRepository, toysRepository, logger)
			}

			actualRespondID, err := respondsService.RespondToTicket(ctx, tc.rawRespondToTicketDTO)
			if tc.errorExpected {
				require.Error(t, err)
				require.IsType(t, tc.err, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, actualRespondID)
		})
	}
}

func TestRespondsService_GetRespondByID(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(
			respondsRepository *mockrepositories.MockRespondsRepository,
			logger *loggerMock.MockLogger,
		)
		respondID     uint64
		expected      *entities.Respond
		errorExpected bool
		err           error
	}{
		{
			name: "successfully got respond",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				_ *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					GetRespondByID(gomock.Any(), respondID).
					Return(respond, nil).
					MaxTimes(1)
			},
			respondID:     respondID,
			expected:      respond,
			errorExpected: false,
		},
		{
			name: "failed to get respond by ID respond not found",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				logger *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(2)).
					Return(nil, &customerrors.RespondNotFoundError{}).
					MaxTimes(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(1)
			},
			respondID:     uint64(2),
			errorExpected: true,
			err:           &customerrors.RespondNotFoundError{},
		},
	}

	mockController := gomock.NewController(t)
	logger := loggerMock.NewMockLogger(mockController)
	respondsRepository := mockrepositories.NewMockRespondsRepository(mockController)
	respondsService := services.NewRespondsService(respondsRepository, nil, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(respondsRepository, logger)
			}

			actualRespond, err := respondsService.GetRespondByID(ctx, tc.respondID)
			if tc.errorExpected {
				require.Error(t, err)
				require.IsType(t, tc.err, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, actualRespond)
		})
	}
}

func TestRespondsService_GetUserResponds(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(
			respondsRepository *mockrepositories.MockRespondsRepository,
			toysRepository *mockrepositories.MockToysRepository,
			logger *loggerMock.MockLogger,
		)
		userID        uint64
		expected      []entities.Respond
		errorExpected bool
		err           error
	}{
		{
			name: "successfully got user responds",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				toysRepository *mockrepositories.MockToysRepository,
				_ *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					GetMasterResponds(gomock.Any(), masterID).
					Return([]entities.Respond{*respond}, nil).
					MaxTimes(1)

				toysRepository.
					EXPECT().
					GetMasterByUserID(gomock.Any(), userID).
					Return(master, nil).
					MaxTimes(1)
			},
			userID:        userID,
			expected:      []entities.Respond{*respond},
			errorExpected: false,
		},
		{
			name: "failed to get respond by userID master not found",
			setupMocks: func(
				_ *mockrepositories.MockRespondsRepository,
				toysRepository *mockrepositories.MockToysRepository,
				_ *loggerMock.MockLogger,
			) {
				toysRepository.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(2)).
					Return(nil, errMasterNotFound).
					MaxTimes(1)
			},
			userID:        2,
			errorExpected: true,
			err:           errMasterNotFound,
		},
		{
			name: "successfully got user responds with no responds",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				toysRepository *mockrepositories.MockToysRepository,
				_ *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					GetMasterResponds(gomock.Any(), uint64(3)).
					Return([]entities.Respond{}, nil).
					MaxTimes(1)

				toysRepository.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(3)).
					Return(&entities.Master{ID: 3, UserID: 3}, nil).
					MaxTimes(1)
			},
			userID:        3,
			expected:      []entities.Respond{},
			errorExpected: false,
		},
	}

	mockController := gomock.NewController(t)
	logger := loggerMock.NewMockLogger(mockController)
	respondsRepository := mockrepositories.NewMockRespondsRepository(mockController)
	toysRepository := mockrepositories.NewMockToysRepository(mockController)
	respondsService := services.NewRespondsService(respondsRepository, toysRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(respondsRepository, toysRepository, logger)
			}

			actualResponds, err := respondsService.GetUserResponds(ctx, tc.userID)
			if tc.errorExpected {
				require.Error(t, err)
				require.IsType(t, tc.err, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, actualResponds)
		})
	}
}
