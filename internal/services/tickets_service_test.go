package services_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/mock/gomock"

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
		Price:       1,
		TagIDs:      []uint32{tagID},
	}
	createTicketDTO = entities.CreateTicketDTO{
		CategoryID:  categoryID,
		UserID:      userID,
		Name:        "test ticket",
		Description: "test description",
		Quantity:    1,
		Price:       1,
		TagIDs:      []uint32{tagID},
	}
)

func TestCommonTicketsService_CreateTicket(t *testing.T) {
	testCases := []struct {
		name            string
		createTicketDTO entities.CreateTicketDTO
		expected        uint64
		errorExpected   bool
		err             error
	}{
		{
			name:            "successfully created ticket",
			createTicketDTO: createTicketDTO,
			expected:        ticketID,
			errorExpected:   false,
		},
		{
			name: "fail to create ticket due to already exists",
			createTicketDTO: entities.CreateTicketDTO{
				CategoryID:  categoryID,
				UserID:      2,
				Name:        "test ticket fail",
				Description: "test description fail",
			},
			errorExpected: true,
			err:           &customerrors.TicketAlreadyExistsError{},
		},
		{
			name:            "fail to create ticket due to tag not found",
			createTicketDTO: entities.CreateTicketDTO{CategoryID: categoryID, TagIDs: []uint32{2}},
			errorExpected:   true,
			err:             &customerrors.TagNotFoundError{},
		},
		{
			name:            "fail to create ticket due to category not found",
			createTicketDTO: entities.CreateTicketDTO{CategoryID: 2},
			errorExpected:   true,
			err:             &customerrors.CategoryNotFoundError{},
		},
	}

	mockController := gomock.NewController(t)
	ticketsRepository := mockrepositories.NewMockTicketsRepository(mockController)
	ticketsRepository.EXPECT().GetUserTickets(gomock.Any(), userID).Return(nil, nil).MaxTimes(1)
	ticketsRepository.EXPECT().GetUserTickets(gomock.Any(), uint64(2)).Return(
		[]entities.Ticket{
			{
				ID:          2,
				CategoryID:  categoryID,
				UserID:      2,
				Name:        "test ticket fail",
				Description: "test description fail",
			},
		},
		nil,
	).MaxTimes(1)
	ticketsRepository.EXPECT().CreateTicket(gomock.Any(), createTicketDTO).Return(ticketID, nil).MaxTimes(1)

	toysRepository := mockrepositories.NewMockToysRepository(mockController)
	toysRepository.EXPECT().GetAllTags(gomock.Any()).Return([]entities.Tag{{ID: tagID}}, nil).MaxTimes(2)
	toysRepository.EXPECT().GetAllTags(gomock.Any()).Return(nil, nil).MaxTimes(1)
	toysRepository.EXPECT().GetAllCategories(gomock.Any()).Return(
		[]entities.Category{{ID: categoryID}},
		nil,
	).MaxTimes(3)
	toysRepository.EXPECT().GetAllCategories(gomock.Any()).Return(nil, nil).MaxTimes(1)

	logger := slog.New(slog.NewJSONHandler(bytes.NewBuffer(make([]byte, 1000)), nil))
	ticketsService := services.NewCommonTicketsService(ticketsRepository, toysRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualTicketID, err := ticketsService.CreateTicket(ctx, tc.createTicketDTO)
			if tc.errorExpected {
				require.Error(t, err)
				require.IsType(t, tc.err, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, actualTicketID)
		})
	}
}

func TestCommonTicketsService_GetTicketByID(t *testing.T) {
	testCases := []struct {
		name          string
		ticketID      uint64
		expected      *entities.Ticket
		errorExpected bool
		err           error
	}{
		{
			name:          "successfully got ticket",
			ticketID:      ticketID,
			expected:      ticket,
			errorExpected: false,
		},
		{
			name:          "failed to get ticket by id ticket not found",
			ticketID:      uint64(2),
			errorExpected: true,
			err:           &customerrors.TicketNotFoundError{},
		},
	}

	mockController := gomock.NewController(t)
	ticketsRepository := mockrepositories.NewMockTicketsRepository(mockController)
	ticketsRepository.EXPECT().GetTicketByID(gomock.Any(), ticketID).Return(ticket, nil).MaxTimes(1)
	ticketsRepository.EXPECT().GetTicketByID(gomock.Any(), uint64(2)).Return(
		nil,
		&customerrors.TicketNotFoundError{},
	).MaxTimes(1)

	logger := slog.New(slog.NewJSONHandler(bytes.NewBuffer(make([]byte, 1000)), nil))
	ticketsService := services.NewCommonTicketsService(ticketsRepository, nil, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
