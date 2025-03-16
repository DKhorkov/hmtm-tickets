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
	respondToTicketDTO = entities.RespondToTicketDTO{
		TicketID: ticketID,
		MasterID: masterID,
	}
	respond = &entities.Respond{
		TicketID: ticketID,
		MasterID: masterID,
		ID:       respondID,
	}
)

func TestRespondsService_RespondToTicket(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(
			respondsRepository *mockrepositories.MockRespondsRepository,
			logger *loggerMock.MockLogger,
		)
		respondToTicketDTO entities.RespondToTicketDTO
		expected           uint64
		errorExpected      bool
	}{
		{
			name: "successfully Responded to Ticket",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				_ *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					RespondToTicket(gomock.Any(), respondToTicketDTO).
					Return(respondID, nil).
					Times(1)
			},
			respondToTicketDTO: respondToTicketDTO,
			expected:           respondID,
			errorExpected:      false,
		},
		{
			name: "failed to respond to Ticket",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				_ *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					RespondToTicket(gomock.Any(), respondToTicketDTO).
					Return(uint64(0), errors.New("test")).
					Times(1)
			},
			respondToTicketDTO: respondToTicketDTO,
			errorExpected:      true,
		},
	}

	mockController := gomock.NewController(t)
	logger := loggerMock.NewMockLogger(mockController)
	respondsRepository := mockrepositories.NewMockRespondsRepository(mockController)
	respondsService := services.NewRespondsService(respondsRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(respondsRepository, logger)
			}

			actualRespondID, err := respondsService.RespondToTicket(ctx, tc.respondToTicketDTO)
			if tc.errorExpected {
				require.Error(t, err)
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
					Times(1)
			},
			respondID:     respondID,
			expected:      respond,
			errorExpected: false,
		},
		{
			name: "failed to get Respond by ID",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				logger *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(2)).
					Return(nil, &customerrors.RespondNotFoundError{}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			respondID:     uint64(2),
			errorExpected: true,
		},
	}

	mockController := gomock.NewController(t)
	logger := loggerMock.NewMockLogger(mockController)
	respondsRepository := mockrepositories.NewMockRespondsRepository(mockController)
	respondsService := services.NewRespondsService(respondsRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(respondsRepository, logger)
			}

			actualRespond, err := respondsService.GetRespondByID(ctx, tc.respondID)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, actualRespond)
		})
	}
}

func TestRespondsService_GetMasterResponds(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(
			respondsRepository *mockrepositories.MockRespondsRepository,
			logger *loggerMock.MockLogger,
		)
		masterID      uint64
		expected      []entities.Respond
		errorExpected bool
	}{
		{
			name: "successfully got Master Responds",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				_ *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					GetMasterResponds(gomock.Any(), masterID).
					Return([]entities.Respond{*respond}, nil).
					Times(1)
			},
			masterID:      masterID,
			expected:      []entities.Respond{*respond},
			errorExpected: false,
		},
		{
			name: "failed to get Responds by masterID",
			setupMocks: func(
				respondsRepository *mockrepositories.MockRespondsRepository,
				_ *loggerMock.MockLogger,
			) {
				respondsRepository.
					EXPECT().
					GetMasterResponds(gomock.Any(), masterID).
					Return(nil, errors.New("test")).
					Times(1)
			},
			masterID:      masterID,
			errorExpected: true,
		},
	}

	mockController := gomock.NewController(t)
	logger := loggerMock.NewMockLogger(mockController)
	respondsRepository := mockrepositories.NewMockRespondsRepository(mockController)
	respondsService := services.NewRespondsService(respondsRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(respondsRepository, logger)
			}

			actualResponds, err := respondsService.GetMasterResponds(ctx, tc.masterID)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, actualResponds)
		})
	}
}
