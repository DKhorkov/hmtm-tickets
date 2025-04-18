package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mocklogger "github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/pointers"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/services"
	mockrepositories "github.com/DKhorkov/hmtm-tickets/mocks/repositories"
)

const (
	categoryID uint32 = 1
	tagID      uint32 = 1
)

var (
	ticket = &entities.Ticket{
		ID:          ticketID,
		CategoryID:  categoryID,
		UserID:      userID,
		Name:        "test ticket",
		Description: "test description",
		Quantity:    1,
		Price:       pointers.New[float32](1),
		TagIDs:      []uint32{tagID},
	}
	createTicketDTO = entities.CreateTicketDTO{
		CategoryID:  categoryID,
		UserID:      userID,
		Name:        "test ticket",
		Description: "test description",
		Quantity:    1,
		Price:       pointers.New[float32](1),
		TagIDs:      []uint32{tagID},
	}
)

func TestTicketsService_CreateTicket(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(
			ticketsRepository *mockrepositories.MockTicketsRepository,
			logger *mocklogger.MockLogger,
		)
		createTicketDTO entities.CreateTicketDTO
		expected        uint64
		errorExpected   bool
	}{
		{
			name: "successfully created ticket",
			setupMocks: func(
				ticketsRepository *mockrepositories.MockTicketsRepository,
				_ *mocklogger.MockLogger,
			) {
				ticketsRepository.
					EXPECT().
					CreateTicket(gomock.Any(), createTicketDTO).
					Return(ticketID, nil).
					Times(1)
			},
			createTicketDTO: createTicketDTO,
			expected:        ticketID,
			errorExpected:   false,
		},
		{
			name: "fail to create ticket due to already exists",
			setupMocks: func(
				ticketsRepository *mockrepositories.MockTicketsRepository,
				_ *mocklogger.MockLogger,
			) {
				ticketsRepository.
					EXPECT().
					CreateTicket(gomock.Any(), createTicketDTO).
					Return(uint64(0), errors.New("test")).
					Times(1)
			},
			createTicketDTO: createTicketDTO,
			errorExpected:   true,
		},
	}

	ctrl := gomock.NewController(t)
	ticketsRepository := mockrepositories.NewMockTicketsRepository(ctrl)
	logger := mocklogger.NewMockLogger(ctrl)
	ticketsService := services.NewTicketsService(ticketsRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsRepository, logger)
			}

			actualTicketID, err := ticketsService.CreateTicket(ctx, tc.createTicketDTO)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, actualTicketID)
		})
	}
}

func TestTicketsService_GetTicketByID(t *testing.T) {
	testCases := []struct {
		name       string
		setupMocks func(
			ticketsRepository *mockrepositories.MockTicketsRepository,
			logger *mocklogger.MockLogger,
		)
		ticketID      uint64
		expected      *entities.Ticket
		errorExpected bool
		err           error
	}{
		{
			name: "successfully got ticket",
			setupMocks: func(
				ticketsRepository *mockrepositories.MockTicketsRepository,
				_ *mocklogger.MockLogger,
			) {
				ticketsRepository.
					EXPECT().
					GetTicketByID(gomock.Any(), ticketID).
					Return(ticket, nil).
					Times(1)
			},
			ticketID:      ticketID,
			expected:      ticket,
			errorExpected: false,
		},
		{
			name: "failed to get ticket by id ticket not found",
			setupMocks: func(
				ticketsRepository *mockrepositories.MockTicketsRepository,
				logger *mocklogger.MockLogger,
			) {
				ticketsRepository.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(2)).
					Return(nil, &customerrors.TicketNotFoundError{}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			ticketID:      uint64(2),
			errorExpected: true,
			err:           &customerrors.TicketNotFoundError{},
		},
	}

	ctrl := gomock.NewController(t)
	logger := mocklogger.NewMockLogger(ctrl)
	ticketsRepository := mockrepositories.NewMockTicketsRepository(ctrl)
	ticketsService := services.NewTicketsService(ticketsRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsRepository, logger)
			}

			actualTicket, err := ticketsService.GetTicketByID(ctx, tc.ticketID)
			if tc.errorExpected {
				require.Error(t, err)
				require.IsType(t, tc.err, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, actualTicket)
		})
	}
}

func TestTicketsService_GetAllTickets(t *testing.T) {
	testCases := []struct {
		name            string
		setupMocks      func(ticketsRepository *mockrepositories.MockTicketsRepository)
		expectedTickets []entities.Ticket
		errorExpected   bool
	}{
		{
			name: "success",
			setupMocks: func(ticketsRepository *mockrepositories.MockTicketsRepository) {
				ticketsRepository.
					EXPECT().
					GetAllTickets(gomock.Any()).
					Return([]entities.Ticket{{ID: 1}}, nil).
					Times(1)
			},
			expectedTickets: []entities.Ticket{{ID: 1}},
			errorExpected:   false,
		},
		{
			name: "repository error",
			setupMocks: func(ticketsRepository *mockrepositories.MockTicketsRepository) {
				ticketsRepository.
					EXPECT().
					GetAllTickets(gomock.Any()).
					Return(nil, errors.New("fetch failed")).
					Times(1)
			},
			expectedTickets: nil,
			errorExpected:   true,
		},
	}

	ctrl := gomock.NewController(t)
	logger := mocklogger.NewMockLogger(ctrl)
	ticketsRepository := mockrepositories.NewMockTicketsRepository(ctrl)
	ticketsService := services.NewTicketsService(ticketsRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsRepository)
			}

			tickets, err := ticketsService.GetAllTickets(ctx)
			if tc.errorExpected {
				require.Error(t, err)
				require.Equal(t, tc.expectedTickets, tickets)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTickets, tickets)
			}
		})
	}
}

func TestTicketsService_GetUserTickets(t *testing.T) {
	testCases := []struct {
		name            string
		userID          uint64
		setupMocks      func(ticketsRepository *mockrepositories.MockTicketsRepository)
		expectedTickets []entities.Ticket
		errorExpected   bool
	}{
		{
			name:   "success",
			userID: 1,
			setupMocks: func(ticketsRepository *mockrepositories.MockTicketsRepository) {
				ticketsRepository.
					EXPECT().
					GetUserTickets(gomock.Any(), uint64(1)).
					Return([]entities.Ticket{{ID: 1, UserID: 1}}, nil).
					Times(1)
			},
			expectedTickets: []entities.Ticket{{ID: 1, UserID: 1}},
			errorExpected:   false,
		},
		{
			name:   "repository error",
			userID: 2,
			setupMocks: func(ticketsRepository *mockrepositories.MockTicketsRepository) {
				ticketsRepository.
					EXPECT().
					GetUserTickets(gomock.Any(), uint64(2)).
					Return(nil, errors.New("fetch failed")).
					Times(1)
			},
			expectedTickets: nil,
			errorExpected:   true,
		},
	}

	ctrl := gomock.NewController(t)
	logger := mocklogger.NewMockLogger(ctrl)
	ticketsRepository := mockrepositories.NewMockTicketsRepository(ctrl)
	ticketsService := services.NewTicketsService(ticketsRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsRepository)
			}

			tickets, err := ticketsService.GetUserTickets(ctx, tc.userID)
			if tc.errorExpected {
				require.Error(t, err)
				require.Equal(t, tc.expectedTickets, tickets)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTickets, tickets)
			}
		})
	}
}

func TestTicketsService_DeleteTicket(t *testing.T) {
	testCases := []struct {
		name          string
		id            uint64
		setupMocks    func(ticketsRepository *mockrepositories.MockTicketsRepository)
		errorExpected bool
	}{
		{
			name: "success",
			id:   1,
			setupMocks: func(ticketsRepository *mockrepositories.MockTicketsRepository) {
				ticketsRepository.
					EXPECT().
					DeleteTicket(gomock.Any(), uint64(1)).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},
		{
			name: "repository error",
			id:   1,
			setupMocks: func(ticketsRepository *mockrepositories.MockTicketsRepository) {
				ticketsRepository.
					EXPECT().
					DeleteTicket(gomock.Any(), uint64(1)).
					Return(errors.New("delete failed")).
					Times(1)
			},
			errorExpected: true,
		},
	}

	ctrl := gomock.NewController(t)
	logger := mocklogger.NewMockLogger(ctrl)
	ticketsRepository := mockrepositories.NewMockTicketsRepository(ctrl)
	ticketsService := services.NewTicketsService(ticketsRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsRepository)
			}
			err := ticketsService.DeleteTicket(ctx, tc.id)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTicketsService_UpdateTicket(t *testing.T) {
	testCases := []struct {
		name          string
		ticketData    entities.UpdateTicketDTO
		setupMocks    func(ticketsRepository *mockrepositories.MockTicketsRepository)
		errorExpected bool
	}{
		{
			name: "success",
			ticketData: entities.UpdateTicketDTO{
				ID:          1,
				Name:        pointers.New("Updated Ticket"),
				Description: pointers.New("Updated Desc"),
			},
			setupMocks: func(ticketsRepository *mockrepositories.MockTicketsRepository) {
				ticketsRepository.
					EXPECT().
					UpdateTicket(gomock.Any(), entities.UpdateTicketDTO{
						ID:          1,
						Name:        pointers.New("Updated Ticket"),
						Description: pointers.New("Updated Desc"),
					}).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},
		{
			name: "repository error",
			ticketData: entities.UpdateTicketDTO{
				ID: 1,
			},
			setupMocks: func(ticketsRepository *mockrepositories.MockTicketsRepository) {
				ticketsRepository.
					EXPECT().
					UpdateTicket(gomock.Any(), entities.UpdateTicketDTO{ID: 1}).
					Return(errors.New("update failed")).
					Times(1)
			},
			errorExpected: true,
		},
	}

	ctrl := gomock.NewController(t)
	logger := mocklogger.NewMockLogger(ctrl)
	ticketsRepository := mockrepositories.NewMockTicketsRepository(ctrl)
	ticketsService := services.NewTicketsService(ticketsRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsRepository)
			}

			err := ticketsService.UpdateTicket(ctx, tc.ticketData)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
