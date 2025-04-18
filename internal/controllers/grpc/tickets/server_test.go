package tickets

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	customgrpc "github.com/DKhorkov/libs/grpc"
	mocklogging "github.com/DKhorkov/libs/logging/mocks"
	"github.com/DKhorkov/libs/pointers"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	mockusecases "github.com/DKhorkov/hmtm-tickets/mocks/usecases"
)

func TestServerAPI_DeleteTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.DeleteTicketIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in:   &tickets.DeleteTicketIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					DeleteTicket(gomock.Any(), uint64(1)).
					Return(nil).
					Times(1)
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "not found error",
			in:   &tickets.DeleteTicketIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					DeleteTicket(gomock.Any(), uint64(1)).
					Return(&customerrors.TicketNotFoundError{Message: "ticket with ID=1 not found"}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.NotFound, Message: "ticket with ID=1 not found"},
			errorExpected: true,
		},
		{
			name: "internal error",
			in:   &tickets.DeleteTicketIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					DeleteTicket(gomock.Any(), uint64(1)).
					Return(errors.New("internal error")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.Internal, Message: "internal error"},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(useCases, logger)
			}

			resp, err := api.DeleteTicket(context.Background(), tc.in)
			if tc.errorExpected {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.IsType(t, &emptypb.Empty{}, resp)
			}
		})
	}
}

func TestServerAPI_UpdateTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.UpdateTicketIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in: &tickets.UpdateTicketIn{
				ID:          1,
				CategoryID:  pointers.New[uint32](2),
				Name:        pointers.New("Updated Ticket"),
				Description: pointers.New("Updated Desc"),
				Price:       pointers.New[float32](50),
				Quantity:    pointers.New[uint32](5),
				TagIDs:      []uint32{1, 2},
				Attachments: []string{"new_attachment.jpg"},
			},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					UpdateTicket(gomock.Any(), entities.RawUpdateTicketDTO{
						ID:          1,
						CategoryID:  pointers.New[uint32](2),
						Name:        pointers.New("Updated Ticket"),
						Description: pointers.New("Updated Desc"),
						Price:       pointers.New[float32](50),
						Quantity:    pointers.New[uint32](5),
						TagIDs:      []uint32{1, 2},
						Attachments: []string{"new_attachment.jpg"},
					}).
					Return(nil).
					Times(1)
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "not found error (ticket)",
			in:   &tickets.UpdateTicketIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					UpdateTicket(gomock.Any(), entities.RawUpdateTicketDTO{ID: 1}).
					Return(&customerrors.TicketNotFoundError{Message: "ticket with ID=1 not found"}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.NotFound, Message: "ticket with ID=1 not found"},
			errorExpected: true,
		},
		{
			name: "not found error (category)",
			in:   &tickets.UpdateTicketIn{ID: 1, CategoryID: pointers.New[uint32](2)},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					UpdateTicket(gomock.Any(), entities.RawUpdateTicketDTO{
						ID:         1,
						CategoryID: pointers.New[uint32](2),
					}).
					Return(&customerrors.CategoryNotFoundError{Message: "2"}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.NotFound, Message: "category with ID=2 not found"},
			errorExpected: true,
		},
		{
			name: "internal error",
			in:   &tickets.UpdateTicketIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					UpdateTicket(gomock.Any(), entities.RawUpdateTicketDTO{ID: 1}).
					Return(errors.New("internal error")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.Internal, Message: "internal error"},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(useCases, logger)
			}

			resp, err := api.UpdateTicket(context.Background(), tc.in)
			if tc.errorExpected {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.IsType(t, &emptypb.Empty{}, resp)
			}
		})
	}
}

func TestServerAPI_CreateTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.CreateTicketIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedOut   *tickets.CreateTicketOut
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in: &tickets.CreateTicketIn{
				UserID:      1,
				CategoryID:  2,
				Name:        "New Ticket",
				Description: "New Desc",
				Price:       pointers.New[float32](50),
				Quantity:    3,
				TagIDs:      []uint32{1, 2},
				Attachments: []string{"attachment.jpg"},
			},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					CreateTicket(gomock.Any(), entities.CreateTicketDTO{
						UserID:      1,
						CategoryID:  2,
						Name:        "New Ticket",
						Description: "New Desc",
						Price:       pointers.New[float32](50),
						Quantity:    3,
						TagIDs:      []uint32{1, 2},
						Attachments: []string{"attachment.jpg"},
					}).
					Return(uint64(1), nil).
					Times(1)
			},
			expectedOut:   &tickets.CreateTicketOut{TicketID: 1},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "already exists error",
			in:   &tickets.CreateTicketIn{UserID: 1, Name: "New Ticket"},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					CreateTicket(gomock.Any(), entities.CreateTicketDTO{
						UserID: 1,
						Name:   "New Ticket",
					}).
					Return(uint64(0), &customerrors.TicketAlreadyExistsError{}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.AlreadyExists, Message: "ticket already exists"},
			errorExpected: true,
		},
		{
			name: "internal error",
			in:   &tickets.CreateTicketIn{UserID: 1, Name: "New Ticket"},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					CreateTicket(gomock.Any(), entities.CreateTicketDTO{
						UserID: 1,
						Name:   "New Ticket",
					}).
					Return(uint64(0), errors.New("internal error")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.Internal, Message: "internal error"},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(useCases, logger)
			}

			resp, err := api.CreateTicket(context.Background(), tc.in)
			if tc.errorExpected {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOut, resp)
			}
		})
	}
}

func TestServerAPI_GetTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.GetTicketIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedOut   *tickets.GetTicketOut
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in:   &tickets.GetTicketIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				ticket := &entities.Ticket{
					ID:          1,
					UserID:      2,
					CategoryID:  3,
					Name:        "Test Ticket",
					Description: "Test Desc",
					Price:       pointers.New[float32](50),
					Quantity:    5,
					TagIDs:      []uint32{1, 2},
					Attachments: []entities.Attachment{
						{ID: 1, TicketID: 1, Link: "attachment.jpg", CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)},
					},
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				}

				useCases.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(ticket, nil).
					Times(1)
			},
			expectedOut: &tickets.GetTicketOut{
				ID:          1,
				UserID:      2,
				CategoryID:  3,
				Name:        "Test Ticket",
				Description: "Test Desc",
				Price:       pointers.New[float32](50),
				Quantity:    5,
				TagIDs:      []uint32{1, 2},
				Attachments: []*tickets.Attachment{
					{ID: 1, TicketID: 1, Link: "attachment.jpg", CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)), UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC))},
				},
				CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "not found error",
			in:   &tickets.GetTicketIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(nil, &customerrors.TicketNotFoundError{Message: "ticket with ID=1 not found"}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.NotFound, Message: "ticket with ID=1 not found"},
			errorExpected: true,
		},
		{
			name: "internal error",
			in:   &tickets.GetTicketIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					GetTicketByID(gomock.Any(), uint64(1)).
					Return(nil, errors.New("internal error")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.Internal, Message: "internal error"},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(useCases, logger)
			}

			resp, err := api.GetTicket(context.Background(), tc.in)
			if tc.errorExpected {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOut, resp)
			}
		})
	}
}

func TestServerAPI_GetTickets(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedOut   *tickets.GetTicketsOut
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				ticketsList := []entities.Ticket{
					{
						ID:          1,
						UserID:      2,
						CategoryID:  3,
						Name:        "Ticket 1",
						Description: "Desc 1",
						Price:       pointers.New[float32](50),
						Quantity:    5,
						CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}

				useCases.
					EXPECT().
					GetAllTickets(gomock.Any()).
					Return(ticketsList, nil).
					Times(1)
			},
			expectedOut: &tickets.GetTicketsOut{
				Tickets: []*tickets.GetTicketOut{
					{
						ID:          1,
						UserID:      2,
						CategoryID:  3,
						Name:        "Ticket 1",
						Description: "Desc 1",
						Price:       pointers.New[float32](50),
						Quantity:    5,
						TagIDs:      nil,
						Attachments: []*tickets.Attachment{},
						CreatedAt:   timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt:   timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "internal error",
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					GetAllTickets(gomock.Any()).
					Return(nil, errors.New("internal error")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.Internal, Message: "internal error"},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(useCases, logger)
			}

			resp, err := api.GetTickets(context.Background(), &emptypb.Empty{})
			if tc.errorExpected {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOut, resp)
			}
		})
	}
}

func TestServerAPI_GetUserTickets(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.GetUserTicketsIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedOut   *tickets.GetTicketsOut
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in:   &tickets.GetUserTicketsIn{UserID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				ticketsList := []entities.Ticket{
					{
						ID:          1,
						UserID:      1,
						CategoryID:  2,
						Name:        "User Ticket",
						Description: "User Desc",
						Price:       pointers.New[float32](50),
						Quantity:    3,
						CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}

				useCases.
					EXPECT().
					GetUserTickets(gomock.Any(), uint64(1)).
					Return(ticketsList, nil).
					Times(1)
			},
			expectedOut: &tickets.GetTicketsOut{
				Tickets: []*tickets.GetTicketOut{
					{
						ID:          1,
						UserID:      1,
						CategoryID:  2,
						Name:        "User Ticket",
						Description: "User Desc",
						Price:       pointers.New[float32](50),
						Quantity:    3,
						TagIDs:      nil,
						Attachments: []*tickets.Attachment{},
						CreatedAt:   timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt:   timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "internal error",
			in:   &tickets.GetUserTicketsIn{UserID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					GetUserTickets(gomock.Any(), uint64(1)).
					Return(nil, errors.New("internal error")).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.Internal, Message: "internal error"},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupMocks != nil {
				tc.setupMocks(useCases, logger)
			}

			resp, err := api.GetUserTickets(context.Background(), tc.in)
			if tc.errorExpected {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOut, resp)
			}
		})
	}
}
