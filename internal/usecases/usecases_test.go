package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mockservices "github.com/DKhorkov/hmtm-tickets/mocks/services"
	mocklogging "github.com/DKhorkov/libs/logging/mocks"
	mocknats "github.com/DKhorkov/libs/nats/mocks"
	"github.com/DKhorkov/libs/pointers"

	"github.com/DKhorkov/hmtm-tickets/internal/config"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
)

func TestUseCases_CreateTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{
		Subjects: config.NATSSubjects{
			TicketUpdated: "update.ticket",
			TicketDeleted: "delete.ticket",
		},
	}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		ticketData entities.CreateTicketDTO
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expectedID    uint64
		errorExpected bool
	}{
		{
			name: "success",
			ticketData: entities.CreateTicketDTO{
				UserID:      1,
				CategoryID:  1,
				TagIDs:      []uint32{1, 2},
				Name:        "Test Ticket",
				Description: "Test Description",
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return([]entities.Category{{ID: 1}}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{{ID: 1}, {ID: 2}}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					GetUserTickets(gomock.Any(), uint64(1), nil, nil).
					Return([]entities.Ticket{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					CreateTicket(gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					Times(1)
			},
			expectedID:    1,
			errorExpected: false,
		},
		{
			name: "category not found",
			ticketData: entities.CreateTicketDTO{
				UserID:      1,
				CategoryID:  1,
				TagIDs:      []uint32{1},
				Name:        "Test Ticket",
				Description: "Test Description",
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return([]entities.Category{{ID: 2}}, nil).
					Times(1)
			},
			expectedID:    0,
			errorExpected: true,
		},
		{
			name: "tag not found",
			ticketData: entities.CreateTicketDTO{
				UserID:      1,
				CategoryID:  1,
				TagIDs:      []uint32{1, 3},
				Name:        "Test Ticket",
				Description: "Test Description",
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return([]entities.Category{{ID: 1}}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{{ID: 1}, {ID: 2}}, nil).
					Times(1)
			},
			expectedID:    0,
			errorExpected: true,
		},
		{
			name: "ticket already exists",
			ticketData: entities.CreateTicketDTO{
				UserID:      1,
				CategoryID:  1,
				TagIDs:      []uint32{1},
				Name:        "Test Ticket",
				Description: "Test Description",
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return([]entities.Category{{ID: 1}}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{{ID: 1}}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					GetUserTickets(gomock.Any(), uint64(1), nil, nil).
					Return([]entities.Ticket{{
						Name:        "Test Ticket",
						Description: "Test Description",
						CategoryID:  1,
					}}, nil).
					Times(1)
			},
			expectedID:    0,
			errorExpected: true,
		},
		{
			name: "create ticket error",
			ticketData: entities.CreateTicketDTO{
				UserID:      1,
				CategoryID:  1,
				TagIDs:      []uint32{1},
				Name:        "Test Ticket",
				Description: "Test Description",
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return([]entities.Category{{ID: 1}}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{{ID: 1}}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					GetUserTickets(gomock.Any(), uint64(1), nil, nil).
					Return([]entities.Ticket{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					CreateTicket(gomock.Any(), gomock.Any()).
					Return(uint64(0), errors.New("create failed")).
					Times(1)
			},
			expectedID:    0,
			errorExpected: true,
		},
		{
			name: "get all tags error",
			ticketData: entities.CreateTicketDTO{
				UserID:      1,
				CategoryID:  1,
				TagIDs:      []uint32{1},
				Name:        "Test Ticket",
				Description: "Test Description",
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return([]entities.Category{{ID: 1}}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return(nil, errors.New("get all tags error")).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "get all categories error",
			ticketData: entities.CreateTicketDTO{
				UserID:      1,
				CategoryID:  1,
				TagIDs:      []uint32{1},
				Name:        "Test Ticket",
				Description: "Test Description",
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return(nil, errors.New("get all categories error")).
					Times(1)
			},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			id, err := useCases.CreateTicket(context.Background(), tc.ticketData)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, id)
			}
		})
	}
}

func TestUseCases_GetTicketByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		id         uint64
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expectedTicket *entities.Ticket
		errorExpected  bool
	}{
		{
			name: "success",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1}, nil).
					Times(1)
			},
			expectedTicket: &entities.Ticket{ID: 1},
			errorExpected:  false,
		},
		{
			name: "not found",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			expectedTicket: nil,
			errorExpected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			ticket, err := useCases.GetTicketByID(context.Background(), tc.id)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTicket, ticket)
			}
		})
	}
}

func TestUseCases_GetTickets(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		pagination *entities.Pagination
		filters    *entities.TicketsFilters
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expectedTickets []entities.Ticket
		errorExpected   bool
	}{
		{
			name: "success",
			pagination: &entities.Pagination{
				Limit:  pointers.New[uint64](1),
				Offset: pointers.New[uint64](1),
			},
			filters: &entities.TicketsFilters{
				Search:              pointers.New("ticket"),
				PriceCeil:           pointers.New[float32](1000),
				PriceFloor:          pointers.New[float32](10),
				QuantityFloor:       pointers.New[uint32](1),
				CategoryIDs:         []uint32{1},
				TagIDs:              []uint32{1},
				CreatedAtOrderByAsc: pointers.New(true),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTickets(
						gomock.Any(),
						&entities.Pagination{
							Limit:  pointers.New[uint64](1),
							Offset: pointers.New[uint64](1),
						},
						&entities.TicketsFilters{
							Search:              pointers.New("ticket"),
							PriceCeil:           pointers.New[float32](1000),
							PriceFloor:          pointers.New[float32](10),
							QuantityFloor:       pointers.New[uint32](1),
							CategoryIDs:         []uint32{1},
							TagIDs:              []uint32{1},
							CreatedAtOrderByAsc: pointers.New(true),
						},
					).
					Return([]entities.Ticket{{ID: 1}}, nil).
					Times(1)
			},
			expectedTickets: []entities.Ticket{{ID: 1}},
			errorExpected:   false,
		},
		{
			name: "error",
			pagination: &entities.Pagination{
				Limit:  pointers.New[uint64](1),
				Offset: pointers.New[uint64](1),
			},
			filters: &entities.TicketsFilters{
				Search:              pointers.New("ticket"),
				PriceCeil:           pointers.New[float32](1000),
				PriceFloor:          pointers.New[float32](10),
				QuantityFloor:       pointers.New[uint32](1),
				CategoryIDs:         []uint32{1},
				TagIDs:              []uint32{1},
				CreatedAtOrderByAsc: pointers.New(true),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTickets(
						gomock.Any(),
						&entities.Pagination{
							Limit:  pointers.New[uint64](1),
							Offset: pointers.New[uint64](1),
						},
						&entities.TicketsFilters{
							Search:              pointers.New("ticket"),
							PriceCeil:           pointers.New[float32](1000),
							PriceFloor:          pointers.New[float32](10),
							QuantityFloor:       pointers.New[uint32](1),
							CategoryIDs:         []uint32{1},
							TagIDs:              []uint32{1},
							CreatedAtOrderByAsc: pointers.New(true),
						},
					).
					Return(nil, errors.New("fetch failed")).
					Times(1)
			},
			expectedTickets: nil,
			errorExpected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			tickets, err := useCases.GetTickets(context.Background(), tc.pagination, tc.filters)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTickets, tickets)
			}
		})
	}
}

func TestUseCases_GetUserTickets(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		pagination *entities.Pagination
		filters    *entities.TicketsFilters
		userID     uint64
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expectedTickets []entities.Ticket
		errorExpected   bool
	}{
		{
			name:   "success",
			userID: 1,
			pagination: &entities.Pagination{
				Limit:  pointers.New[uint64](1),
				Offset: pointers.New[uint64](1),
			},
			filters: &entities.TicketsFilters{
				Search:              pointers.New("ticket"),
				PriceCeil:           pointers.New[float32](1000),
				PriceFloor:          pointers.New[float32](10),
				QuantityFloor:       pointers.New[uint32](1),
				CategoryIDs:         []uint32{1},
				TagIDs:              []uint32{1},
				CreatedAtOrderByAsc: pointers.New(true),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetUserTickets(
						gomock.Any(),
						uint64(1),
						&entities.Pagination{
							Limit:  pointers.New[uint64](1),
							Offset: pointers.New[uint64](1),
						},
						&entities.TicketsFilters{
							Search:              pointers.New("ticket"),
							PriceCeil:           pointers.New[float32](1000),
							PriceFloor:          pointers.New[float32](10),
							QuantityFloor:       pointers.New[uint32](1),
							CategoryIDs:         []uint32{1},
							TagIDs:              []uint32{1},
							CreatedAtOrderByAsc: pointers.New(true),
						},
					).
					Return([]entities.Ticket{{ID: 1, UserID: 1}}, nil).
					Times(1)
			},
			expectedTickets: []entities.Ticket{{ID: 1, UserID: 1}},
			errorExpected:   false,
		},
		{
			name:   "error",
			userID: 1,
			pagination: &entities.Pagination{
				Limit:  pointers.New[uint64](1),
				Offset: pointers.New[uint64](1),
			},
			filters: &entities.TicketsFilters{
				Search:              pointers.New("ticket"),
				PriceCeil:           pointers.New[float32](1000),
				PriceFloor:          pointers.New[float32](10),
				QuantityFloor:       pointers.New[uint32](1),
				CategoryIDs:         []uint32{1},
				TagIDs:              []uint32{1},
				CreatedAtOrderByAsc: pointers.New(true),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetUserTickets(
						gomock.Any(),
						uint64(1),
						&entities.Pagination{
							Limit:  pointers.New[uint64](1),
							Offset: pointers.New[uint64](1),
						},
						&entities.TicketsFilters{
							Search:              pointers.New("ticket"),
							PriceCeil:           pointers.New[float32](1000),
							PriceFloor:          pointers.New[float32](10),
							QuantityFloor:       pointers.New[uint32](1),
							CategoryIDs:         []uint32{1},
							TagIDs:              []uint32{1},
							CreatedAtOrderByAsc: pointers.New(true),
						},
					).
					Return(nil, errors.New("fetch failed")).
					Times(1)
			},
			expectedTickets: nil,
			errorExpected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			tickets, err := useCases.GetUserTickets(context.Background(), tc.userID, tc.pagination, tc.filters)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTickets, tickets)
			}
		})
	}
}

func TestUseCases_RespondToTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name        string
		respondData entities.RawRespondToTicketDTO
		setupMocks  func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expectedID    uint64
		errorExpected bool
	}{
		{
			name: "success",
			respondData: entities.RawRespondToTicketDTO{
				TicketID: 1,
				UserID:   2,
				Price:    100,
				Comment:  pointers.New("Test comment"),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1, UserID: 1}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(2)).
					Return(&entities.Master{ID: 2}, nil).
					Times(1)

				respondsService.
					EXPECT().
					GetMasterResponds(gomock.Any(), uint64(2)).
					Return([]entities.Respond{}, nil).
					Times(1)

				respondsService.
					EXPECT().
					RespondToTicket(gomock.Any(), gomock.Any()).
					Return(uint64(1), nil).
					Times(1)
			},
			expectedID:    1,
			errorExpected: false,
		},
		{
			name: "ticket not found",
			respondData: entities.RawRespondToTicketDTO{
				TicketID: 1,
				UserID:   2,
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			expectedID:    0,
			errorExpected: true,
		},
		{
			name: "own ticket",
			respondData: entities.RawRespondToTicketDTO{
				TicketID: 1,
				UserID:   1,
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1, UserID: 1}, nil).
					Times(1)
			},
			expectedID:    0,
			errorExpected: true,
		},
		{
			name: "master not found",
			respondData: entities.RawRespondToTicketDTO{
				TicketID: 1,
				UserID:   2,
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1, UserID: 1}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(2)).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			expectedID:    0,
			errorExpected: true,
		},
		{
			name: "respond already exists",
			respondData: entities.RawRespondToTicketDTO{
				TicketID: 1,
				UserID:   2,
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1, UserID: 1}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(2)).
					Return(&entities.Master{ID: 2}, nil).
					Times(1)

				respondsService.
					EXPECT().
					GetMasterResponds(gomock.Any(), uint64(2)).
					Return([]entities.Respond{{TicketID: 1}}, nil).
					Times(1)
			},
			expectedID:    0,
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			id, err := useCases.RespondToTicket(context.Background(), tc.respondData)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, id)
			}
		})
	}
}

func TestUseCases_GetRespondByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		id         uint64
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expectedRespond *entities.Respond
		errorExpected   bool
	}{
		{
			name: "success",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(&entities.Respond{ID: 1}, nil).
					Times(1)
			},
			expectedRespond: &entities.Respond{ID: 1},
			errorExpected:   false,
		},
		{
			name: "not found",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			expectedRespond: nil,
			errorExpected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			respond, err := useCases.GetRespondByID(context.Background(), tc.id)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedRespond, respond)
			}
		})
	}
}

func TestUseCases_GetTicketResponds(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		ticketID   uint64
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expectedResponds []entities.Respond
		errorExpected    bool
	}{
		{
			name:     "success",
			ticketID: 1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetTicketResponds(gomock.Any(), uint64(1)).
					Return([]entities.Respond{{ID: 1}}, nil).
					Times(1)
			},
			expectedResponds: []entities.Respond{{ID: 1}},
			errorExpected:    false,
		},
		{
			name:     "error",
			ticketID: 1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetTicketResponds(gomock.Any(), uint64(1)).
					Return(nil, errors.New("fetch failed")).
					Times(1)
			},
			expectedResponds: nil,
			errorExpected:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			responds, err := useCases.GetTicketResponds(context.Background(), tc.ticketID)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponds, responds)
			}
		})
	}
}

func TestUseCases_GetUserResponds(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		userID     uint64
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expectedResponds []entities.Respond
		errorExpected    bool
	}{
		{
			name:   "success",
			userID: 1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(1)).
					Return(&entities.Master{ID: 1}, nil).
					Times(1)

				respondsService.
					EXPECT().
					GetMasterResponds(gomock.Any(), uint64(1)).
					Return([]entities.Respond{{ID: 1}}, nil).
					Times(1)
			},
			expectedResponds: []entities.Respond{{ID: 1}},
			errorExpected:    false,
		},
		{
			name:   "master not found",
			userID: 1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			expectedResponds: nil,
			errorExpected:    true,
		},
		{
			name:   "fetch responds error",
			userID: 1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				toysService.
					EXPECT().
					GetMasterByUserID(gomock.Any(), uint64(1)).
					Return(&entities.Master{ID: 1}, nil).
					Times(1)

				respondsService.
					EXPECT().
					GetMasterResponds(gomock.Any(), uint64(1)).
					Return(nil, errors.New("fetch failed")).
					Times(1)
			},
			expectedResponds: nil,
			errorExpected:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			responds, err := useCases.GetUserResponds(context.Background(), tc.userID)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponds, responds)
			}
		})
	}
}

func TestUseCases_UpdateRespond(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name        string
		respondData entities.UpdateRespondDTO
		setupMocks  func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		errorExpected bool
	}{
		{
			name: "success",
			respondData: entities.UpdateRespondDTO{
				ID:      1,
				Price:   pointers.New[float32](200),
				Comment: pointers.New("Test comment"),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(&entities.Respond{ID: 1}, nil).
					Times(1)

				respondsService.
					EXPECT().
					UpdateRespond(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},
		{
			name: "respond not found",
			respondData: entities.UpdateRespondDTO{
				ID: 1,
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "update error",
			respondData: entities.UpdateRespondDTO{
				ID: 1,
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(&entities.Respond{ID: 1}, nil).
					Times(1)

				respondsService.
					EXPECT().
					UpdateRespond(gomock.Any(), gomock.Any()).
					Return(errors.New("update failed")).
					Times(1)
			},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			err := useCases.UpdateRespond(context.Background(), tc.respondData)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCases_DeleteRespond(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		id         uint64
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		errorExpected bool
	}{
		{
			name: "success",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(&entities.Respond{ID: 1}, nil).
					Times(1)

				respondsService.
					EXPECT().
					DeleteRespond(gomock.Any(), uint64(1)).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},
		{
			name: "respond not found",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "delete error",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				respondsService.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(&entities.Respond{ID: 1}, nil).
					Times(1)

				respondsService.
					EXPECT().
					DeleteRespond(gomock.Any(), uint64(1)).
					Return(errors.New("delete failed")).
					Times(1)
			},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			err := useCases.DeleteRespond(context.Background(), tc.id)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCases_DeleteTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{
		Subjects: config.NATSSubjects{
			TicketDeleted: "delete.ticket",
		},
	}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		id         uint64
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		errorExpected bool
	}{
		{
			name: "success",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticket := entities.Ticket{ID: 1, UserID: 1, Name: "Test", Description: "Desc", Quantity: 1}
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&ticket, nil).
					Times(1)

				respondsService.
					EXPECT().
					GetTicketResponds(gomock.Any(), uint64(1)).
					Return([]entities.Respond{{MasterID: 2}}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					DeleteTicket(gomock.Any(), uint64(1)).
					Return(nil).
					Times(1)

				natsPublisher.
					EXPECT().
					Publish("delete.ticket", gomock.Any()).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},
		{
			name: "ticket not found",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "responds fetch error",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1}, nil).
					Times(1)

				respondsService.
					EXPECT().
					GetTicketResponds(gomock.Any(), uint64(1)).
					Return(nil, errors.New("fetch failed")).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "delete error",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1}, nil).
					Times(1)

				respondsService.
					EXPECT().
					GetTicketResponds(gomock.Any(), uint64(1)).
					Return([]entities.Respond{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					DeleteTicket(gomock.Any(), uint64(1)).
					Return(errors.New("delete failed")).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "nats publish error",
			id:   1,
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticket := entities.Ticket{ID: 1, Name: "Test", Description: "Desc", Quantity: 1}
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&ticket, nil).
					Times(1)

				respondsService.
					EXPECT().
					GetTicketResponds(gomock.Any(), uint64(1)).
					Return([]entities.Respond{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					DeleteTicket(gomock.Any(), uint64(1)).
					Return(nil).
					Times(1)

				natsPublisher.
					EXPECT().
					Publish("delete.ticket", gomock.Any()).
					Return(errors.New("publish failed")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			errorExpected: false, // NATS error не возвращается
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}
			err := useCases.DeleteTicket(context.Background(), tc.id)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCases_UpdateTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{
		Subjects: config.NATSSubjects{
			TicketUpdated: "update.ticket",
		},
	}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	testCases := []struct {
		name       string
		ticketData entities.RawUpdateTicketDTO
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		errorExpected bool
	}{
		{
			name: "success",
			ticketData: entities.RawUpdateTicketDTO{
				ID:          1,
				CategoryID:  pointers.New[uint32](1),
				TagIDs:      []uint32{1},
				Name:        pointers.New("Updated Ticket"),
				Description: pointers.New("Updated Desc"),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticket := entities.Ticket{ID: 1, TagIDs: []uint32{2}}
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&ticket, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return([]entities.Category{{ID: 1}}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{{ID: 1}}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					UpdateTicket(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)

				natsPublisher.
					EXPECT().
					Publish("update.ticket", gomock.Any()).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},
		{
			name: "ticket not found",
			ticketData: entities.RawUpdateTicketDTO{
				ID: 1,
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("not found")).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "category not found",
			ticketData: entities.RawUpdateTicketDTO{
				ID:          1,
				CategoryID:  pointers.New[uint32](1),
				TagIDs:      []uint32{1},
				Name:        pointers.New("Updated Ticket"),
				Description: pointers.New("Updated Desc"),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllCategories(gomock.Any()).
					Return([]entities.Category{{ID: 2}}, nil).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "tag not found",
			ticketData: entities.RawUpdateTicketDTO{
				ID:     1,
				TagIDs: []uint32{1},
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{{ID: 2}}, nil).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "update error",
			ticketData: entities.RawUpdateTicketDTO{
				ID: 1,
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&entities.Ticket{ID: 1}, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					UpdateTicket(gomock.Any(), gomock.Any()).
					Return(errors.New("update failed")).
					Times(1)
			},
			errorExpected: true,
		},
		{
			name: "nats publish error",
			ticketData: entities.RawUpdateTicketDTO{
				ID: 1,
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticket := entities.Ticket{ID: 1}
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&ticket, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					UpdateTicket(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)

				natsPublisher.
					EXPECT().
					Publish("update.ticket", gomock.Any()).
					Return(errors.New("publish failed")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			errorExpected: false, // NATS error не возвращается
		},
		// Новый случай: добавление нового вложения
		{
			name: "add new attachment",
			ticketData: entities.RawUpdateTicketDTO{
				ID:          1,
				Attachments: []string{"new_attachment.jpg"},
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticket := entities.Ticket{
					ID:          1,
					Attachments: []entities.Attachment{{ID: 1, Link: "old_attachment.jpg"}},
				}
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&ticket, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					UpdateTicket(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, updateData entities.UpdateTicketDTO) {
						require.Equal(t, []string{"new_attachment.jpg"}, updateData.AttachmentsToAdd)
						require.Equal(t, []uint64{1}, updateData.AttachmentIDsToDelete)
					}).
					Return(nil).
					Times(1)

				natsPublisher.
					EXPECT().
					Publish("update.ticket", gomock.Any()).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},

		// Новый случай: удаление старого вложения
		{
			name: "delete old attachment",
			ticketData: entities.RawUpdateTicketDTO{
				ID:          1,
				Attachments: []string{},
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticket := entities.Ticket{
					ID:          1,
					Attachments: []entities.Attachment{{ID: 1, Link: "old_attachment.jpg"}},
				}
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&ticket, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					UpdateTicket(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, updateData entities.UpdateTicketDTO) {
						require.Equal(t, []string{}, updateData.AttachmentsToAdd)
						require.Equal(t, []uint64{1}, updateData.AttachmentIDsToDelete)
					}).
					Return(nil).
					Times(1)

				natsPublisher.
					EXPECT().
					Publish("update.ticket", gomock.Any()).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},

		// Новый случай: сохранение существующего вложения без изменений
		{
			name: "keep existing attachment",
			ticketData: entities.RawUpdateTicketDTO{
				ID:          1,
				Attachments: []string{"old_attachment.jpg"},
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticket := entities.Ticket{
					ID:          1,
					Attachments: []entities.Attachment{{ID: 1, Link: "old_attachment.jpg"}},
				}
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&ticket, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					UpdateTicket(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, updateData entities.UpdateTicketDTO) {
						require.Equal(t, []string{}, updateData.AttachmentsToAdd)
						require.Equal(t, []uint64{}, updateData.AttachmentIDsToDelete)
					}).
					Return(nil).
					Times(1)

				natsPublisher.
					EXPECT().
					Publish("update.ticket", gomock.Any()).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},

		// Новый случай: добавление и удаление вложений одновременно
		{
			name: "add and delete attachments",
			ticketData: entities.RawUpdateTicketDTO{
				ID:          1,
				Attachments: []string{"old_attachment.jpg", "new_attachment.jpg"},
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticket := entities.Ticket{
					ID:          1,
					Attachments: []entities.Attachment{{ID: 1, Link: "old_attachment.jpg"}, {ID: 2, Link: "to_delete.jpg"}},
				}
				ticketsService.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(&ticket, nil).
					Times(1)

				toysService.
					EXPECT().
					GetAllTags(gomock.Any()).
					Return([]entities.Tag{}, nil).
					Times(1)

				ticketsService.
					EXPECT().
					UpdateTicket(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, updateData entities.UpdateTicketDTO) {
						require.Equal(t, []string{"new_attachment.jpg"}, updateData.AttachmentsToAdd)
						require.Equal(t, []uint64{2}, updateData.AttachmentIDsToDelete)
					}).
					Return(nil).
					Times(1)

				natsPublisher.
					EXPECT().
					Publish("update.ticket", gomock.Any()).
					Return(nil).
					Times(1)
			},
			errorExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}

			err := useCases.UpdateTicket(context.Background(), tc.ticketData)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUseCases_CountTickets(t *testing.T) {
	testCases := []struct {
		name       string
		filters    *entities.TicketsFilters
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expected      uint64
		errorExpected bool
	}{
		{
			name: "success",
			filters: &entities.TicketsFilters{
				Search:              pointers.New("toy2"),
				PriceCeil:           pointers.New[float32](1000),
				PriceFloor:          pointers.New[float32](10),
				QuantityFloor:       pointers.New[uint32](1),
				CategoryIDs:         []uint32{1},
				TagIDs:              []uint32{1},
				CreatedAtOrderByAsc: pointers.New(true),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					CountTickets(
						gomock.Any(),
						&entities.TicketsFilters{
							Search:              pointers.New("toy2"),
							PriceCeil:           pointers.New[float32](1000),
							PriceFloor:          pointers.New[float32](10),
							QuantityFloor:       pointers.New[uint32](1),
							CategoryIDs:         []uint32{1},
							TagIDs:              []uint32{1},
							CreatedAtOrderByAsc: pointers.New(true),
						},
					).
					Return(uint64(1), nil).
					Times(1)
			},
			expected: 1,
		},
	}

	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{
		Subjects: config.NATSSubjects{
			TicketUpdated: "update.ticket",
		},
	}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}

			actual, err := useCases.CountTickets(context.Background(), tc.filters)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, actual)
		})
	}
}
func TestUseCases_CountUserTickets(t *testing.T) {
	testCases := []struct {
		name       string
		userID     uint64
		filters    *entities.TicketsFilters
		setupMocks func(
			ticketsService *mockservices.MockTicketsService,
			respondsService *mockservices.MockRespondsService,
			toysService *mockservices.MockToysService,
			natsPublisher *mocknats.MockPublisher,
			logger *mocklogging.MockLogger,
		)
		expected      uint64
		errorExpected bool
	}{
		{
			name:   "success",
			userID: 1,
			filters: &entities.TicketsFilters{
				Search:              pointers.New("toy2"),
				PriceCeil:           pointers.New[float32](1000),
				PriceFloor:          pointers.New[float32](10),
				QuantityFloor:       pointers.New[uint32](1),
				CategoryIDs:         []uint32{1},
				TagIDs:              []uint32{1},
				CreatedAtOrderByAsc: pointers.New(true),
			},
			setupMocks: func(
				ticketsService *mockservices.MockTicketsService,
				respondsService *mockservices.MockRespondsService,
				toysService *mockservices.MockToysService,
				natsPublisher *mocknats.MockPublisher,
				logger *mocklogging.MockLogger,
			) {
				ticketsService.
					EXPECT().
					CountUserTickets(
						gomock.Any(),
						uint64(1),
						&entities.TicketsFilters{
							Search:              pointers.New("toy2"),
							PriceCeil:           pointers.New[float32](1000),
							PriceFloor:          pointers.New[float32](10),
							QuantityFloor:       pointers.New[uint32](1),
							CategoryIDs:         []uint32{1},
							TagIDs:              []uint32{1},
							CreatedAtOrderByAsc: pointers.New(true),
						},
					).
					Return(uint64(1), nil).
					Times(1)
			},
			expected: 1,
		},
	}

	ctrl := gomock.NewController(t)
	ticketsService := mockservices.NewMockTicketsService(ctrl)
	respondsService := mockservices.NewMockRespondsService(ctrl)
	toysService := mockservices.NewMockToysService(ctrl)
	natsPublisher := mocknats.NewMockPublisher(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	natsConfig := config.NATSConfig{
		Subjects: config.NATSSubjects{
			TicketUpdated: "update.ticket",
		},
	}

	useCases := New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		natsConfig,
		logger,
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(ticketsService, respondsService, toysService, natsPublisher, logger)
			}

			actual, err := useCases.CountUserTickets(context.Background(), tc.userID, tc.filters)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, actual)
		})
	}
}
