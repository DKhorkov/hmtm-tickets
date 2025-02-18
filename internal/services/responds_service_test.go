package services_test

import (
	"bytes"
	"context"
	"errors"
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
		name                  string
		rawRespondToTicketDTO entities.RawRespondToTicketDTO
		expected              uint64
		errorExpected         bool
		err                   error
	}{
		{
			name:                  "successfully responded to ticket",
			rawRespondToTicketDTO: rawRespondToTicketDTO,
			expected:              respondID,
			errorExpected:         false,
		},
		{
			name:                  "failed to respond to ticket respond already exists",
			rawRespondToTicketDTO: entities.RawRespondToTicketDTO{TicketID: ticketID, UserID: 2},
			errorExpected:         true,
			err:                   &customerrors.RespondAlreadyExistsError{},
		},
		{
			name:                  "failed to respond to ticket master not found",
			rawRespondToTicketDTO: entities.RawRespondToTicketDTO{TicketID: ticketID, UserID: 3},
			errorExpected:         true,
			err:                   errMasterNotFound,
		},
	}

	mockController := gomock.NewController(t)
	respondsRepository := mockrepositories.NewMockRespondsRepository(mockController)
	respondsRepository.EXPECT().RespondToTicket(gomock.Any(), respondToTicketDTO).Return(respondID, nil).MaxTimes(1)
	respondsRepository.EXPECT().RespondToTicket(
		gomock.Any(),
		entities.RespondToTicketDTO{TicketID: ticketID, MasterID: 2},
	).Return(uint64(0), &customerrors.RespondAlreadyExistsError{}).MaxTimes(1)

	respondsRepository.EXPECT().GetMasterResponds(gomock.Any(), masterID).Return(nil, nil).MaxTimes(1)
	respondsRepository.EXPECT().GetMasterResponds(gomock.Any(), uint64(2)).Return(
		[]entities.Respond{*respond},
		nil,
	).MaxTimes(1)

	toysRepository := mockrepositories.NewMockToysRepository(mockController)
	toysRepository.EXPECT().GetMasterByUserID(gomock.Any(), userID).Return(master, nil).MaxTimes(2)
	toysRepository.EXPECT().GetMasterByUserID(gomock.Any(), uint64(2)).Return(
		&entities.Master{ID: 2, UserID: 2},
		nil,
	).MaxTimes(1)
	toysRepository.EXPECT().GetMasterByUserID(gomock.Any(), uint64(3)).Return(nil, errMasterNotFound).MaxTimes(1)

	logger := slog.New(slog.NewJSONHandler(bytes.NewBuffer(make([]byte, 1000)), nil))
	respondsService := services.NewRespondsService(respondsRepository, toysRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
		name          string
		respondID     uint64
		expected      *entities.Respond
		errorExpected bool
		err           error
	}{
		{
			name:          "successfully got respond",
			respondID:     respondID,
			expected:      respond,
			errorExpected: false,
		},
		{
			name:          "failed to get respond by ID respond not found",
			respondID:     uint64(2),
			errorExpected: true,
			err:           &customerrors.RespondNotFoundError{},
		},
	}

	mockController := gomock.NewController(t)
	respondsRepository := mockrepositories.NewMockRespondsRepository(mockController)
	respondsRepository.EXPECT().GetRespondByID(gomock.Any(), respondID).Return(respond, nil).MaxTimes(1)
	respondsRepository.EXPECT().GetRespondByID(gomock.Any(), uint64(2)).Return(
		nil,
		&customerrors.RespondNotFoundError{},
	).MaxTimes(1)

	logger := slog.New(slog.NewJSONHandler(bytes.NewBuffer(make([]byte, 1000)), nil))
	respondsService := services.NewRespondsService(respondsRepository, nil, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
		name          string
		userID        uint64
		expected      []entities.Respond
		errorExpected bool
		err           error
	}{
		{
			name:          "successfully got user responds",
			userID:        userID,
			expected:      []entities.Respond{*respond},
			errorExpected: false,
		},
		{
			name:          "failed to get respond by userID master not found",
			userID:        2,
			errorExpected: true,
			err:           errMasterNotFound,
		},
		{
			name:          "successfully got user responds with no responds",
			userID:        3,
			expected:      []entities.Respond{},
			errorExpected: false,
		},
	}

	mockController := gomock.NewController(t)
	respondsRepository := mockrepositories.NewMockRespondsRepository(mockController)
	respondsRepository.EXPECT().GetMasterResponds(gomock.Any(), masterID).Return(
		[]entities.Respond{*respond},
		nil,
	).MaxTimes(1)
	respondsRepository.EXPECT().GetMasterResponds(gomock.Any(), uint64(3)).Return(
		[]entities.Respond{},
		nil,
	).MaxTimes(1)

	toysRepository := mockrepositories.NewMockToysRepository(mockController)
	toysRepository.EXPECT().GetMasterByUserID(gomock.Any(), userID).Return(master, nil).MaxTimes(1)
	toysRepository.EXPECT().GetMasterByUserID(gomock.Any(), uint64(2)).Return(nil, errMasterNotFound).MaxTimes(1)
	toysRepository.EXPECT().GetMasterByUserID(gomock.Any(), uint64(3)).Return(
		&entities.Master{ID: 3, UserID: 3},
		nil,
	).MaxTimes(1)

	logger := slog.New(slog.NewJSONHandler(bytes.NewBuffer(make([]byte, 1000)), nil))
	respondsService := services.NewRespondsService(respondsRepository, toysRepository, logger)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
