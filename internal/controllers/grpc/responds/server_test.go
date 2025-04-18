package responds

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

func TestServerAPI_UpdateRespond(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.UpdateRespondIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in: &tickets.UpdateRespondIn{
				ID:      1,
				Price:   pointers.New[float32](200),
				Comment: pointers.New("Updated comment"),
			},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					UpdateRespond(gomock.Any(), entities.UpdateRespondDTO{
						ID:      1,
						Price:   pointers.New[float32](200),
						Comment: pointers.New("Updated comment"),
					}).
					Return(nil).
					Times(1)
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "not found error",
			in: &tickets.UpdateRespondIn{
				ID: 1,
			},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					UpdateRespond(gomock.Any(), entities.UpdateRespondDTO{ID: 1}).
					Return(&customerrors.RespondNotFoundError{}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.NotFound, Message: "respond not found"},
			errorExpected: true,
		},
		{
			name: "internal error",
			in: &tickets.UpdateRespondIn{
				ID: 1,
			},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					UpdateRespond(gomock.Any(), entities.UpdateRespondDTO{ID: 1}).
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

			resp, err := api.UpdateRespond(context.Background(), tc.in)
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

func TestServerAPI_DeleteRespond(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.DeleteRespondIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in:   &tickets.DeleteRespondIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					DeleteRespond(gomock.Any(), uint64(1)).
					Return(nil).
					Times(1)
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "not found error",
			in:   &tickets.DeleteRespondIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					DeleteRespond(gomock.Any(), uint64(1)).
					Return(&customerrors.RespondNotFoundError{}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.NotFound, Message: "respond not found"},
			errorExpected: true,
		},
		{
			name: "internal error",
			in:   &tickets.DeleteRespondIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					DeleteRespond(gomock.Any(), uint64(1)).
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

			resp, err := api.DeleteRespond(context.Background(), tc.in)
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

func TestServerAPI_RespondToTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.RespondToTicketIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedOut   *tickets.RespondToTicketOut
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in: &tickets.RespondToTicketIn{
				UserID:   1,
				TicketID: 2,
				Price:    100,
				Comment:  pointers.New("Updated comment")},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					RespondToTicket(gomock.Any(), entities.RawRespondToTicketDTO{
						UserID:   1,
						TicketID: 2,
						Price:    100,
						Comment:  pointers.New("Updated comment")}).
					Return(uint64(1), nil).
					Times(1)
			},
			expectedOut:   &tickets.RespondToTicketOut{RespondID: 1},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "already exists error",
			in: &tickets.RespondToTicketIn{
				UserID:   1,
				TicketID: 2,
			},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					RespondToTicket(gomock.Any(), entities.RawRespondToTicketDTO{
						UserID:   1,
						TicketID: 2,
					}).
					Return(uint64(0), &customerrors.RespondAlreadyExistsError{}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.AlreadyExists, Message: "respond already exists"},
			errorExpected: true,
		},
		{
			name: "internal error",
			in: &tickets.RespondToTicketIn{
				UserID:   1,
				TicketID: 2,
			},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					RespondToTicket(gomock.Any(), entities.RawRespondToTicketDTO{
						UserID:   1,
						TicketID: 2,
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

			resp, err := api.RespondToTicket(context.Background(), tc.in)
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

func TestServerAPI_GetRespond(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.GetRespondIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedOut   *tickets.GetRespondOut
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in:   &tickets.GetRespondIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				respond := &entities.Respond{
					ID:       1,
					TicketID: 2,
					MasterID: 3,
					Price:    100,
					Comment:  pointers.New("Updated comment"), CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				}

				useCases.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(respond, nil).
					Times(1)
			},
			expectedOut: &tickets.GetRespondOut{
				ID:       1,
				TicketID: 2,
				MasterID: 3,
				Price:    100,
				Comment:  pointers.New("Updated comment"), CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
				UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "not found error",
			in:   &tickets.GetRespondIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
					Return(nil, &customerrors.RespondNotFoundError{}).
					Times(1)

				logger.
					EXPECT().
					ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			expectedErr:   &customgrpc.BaseError{Status: codes.NotFound, Message: "respond not found"},
			errorExpected: true,
		},
		{
			name: "internal error",
			in:   &tickets.GetRespondIn{ID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					GetRespondByID(gomock.Any(), uint64(1)).
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

			resp, err := api.GetRespond(context.Background(), tc.in)
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

func TestServerAPI_GetTicketResponds(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.GetTicketRespondsIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedOut   *tickets.GetRespondsOut
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in:   &tickets.GetTicketRespondsIn{TicketID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				responds := []entities.Respond{
					{
						ID:        1,
						TicketID:  2,
						MasterID:  1,
						Price:     100,
						Comment:   pointers.New("Updated comment"),
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}

				useCases.
					EXPECT().
					GetTicketResponds(gomock.Any(), uint64(1)).
					Return(responds, nil).
					Times(1)
			},
			expectedOut: &tickets.GetRespondsOut{
				Responds: []*tickets.GetRespondOut{
					{
						ID:       1,
						TicketID: 2,
						MasterID: 1,
						Price:    100,
						Comment:  pointers.New("Updated comment"), CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "internal error",
			in:   &tickets.GetTicketRespondsIn{TicketID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					GetTicketResponds(gomock.Any(), uint64(1)).
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

			resp, err := api.GetTicketResponds(context.Background(), tc.in)
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

func TestServerAPI_GetUserResponds(t *testing.T) {
	ctrl := gomock.NewController(t)
	useCases := mockusecases.NewMockUseCases(ctrl)
	logger := mocklogging.NewMockLogger(ctrl)
	api := &ServerAPI{
		useCases: useCases,
		logger:   logger,
	}

	testCases := []struct {
		name          string
		in            *tickets.GetUserRespondsIn
		setupMocks    func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger)
		expectedOut   *tickets.GetRespondsOut
		expectedErr   error
		errorExpected bool
	}{
		{
			name: "success",
			in:   &tickets.GetUserRespondsIn{UserID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				responds := []entities.Respond{
					{
						ID:        1,
						TicketID:  2,
						MasterID:  1,
						Price:     100,
						Comment:   pointers.New("Updated comment"),
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}

				useCases.
					EXPECT().
					GetUserResponds(gomock.Any(), uint64(1)).
					Return(responds, nil).
					Times(1)
			},
			expectedOut: &tickets.GetRespondsOut{
				Responds: []*tickets.GetRespondOut{
					{
						ID:       1,
						TicketID: 2,
						MasterID: 1,
						Price:    100,
						Comment:  pointers.New("Updated comment"), CreatedAt: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt: timestamppb.New(time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			expectedErr:   nil,
			errorExpected: false,
		},
		{
			name: "internal error",
			in:   &tickets.GetUserRespondsIn{UserID: 1},
			setupMocks: func(useCases *mockusecases.MockUseCases, logger *mocklogging.MockLogger) {
				useCases.
					EXPECT().
					GetUserResponds(gomock.Any(), uint64(1)).
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

			resp, err := api.GetUserResponds(context.Background(), tc.in)
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
