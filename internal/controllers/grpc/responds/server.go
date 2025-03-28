package responds

import (
	"context"
	"errors"
	"fmt"

	"github.com/DKhorkov/libs/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	customgrpc "github.com/DKhorkov/libs/grpc"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

var (
	respondNotFoundError      = &customerrors.RespondNotFoundError{}
	respondAlreadyExistsError = &customerrors.RespondAlreadyExistsError{}
)

// RegisterServer handler (serverAPI) for RespondsServer to gRPC server:.
func RegisterServer(gRPCServer *grpc.Server, useCases interfaces.UseCases, logger logging.Logger) {
	tickets.RegisterRespondsServiceServer(
		gRPCServer,
		&ServerAPI{useCases: useCases, logger: logger},
	)
}

type ServerAPI struct {
	// Helps to test single endpoints, if others is not implemented yet
	tickets.UnimplementedRespondsServiceServer
	useCases interfaces.UseCases
	logger   logging.Logger
}

// UpdateRespond handler updates Respond with provided ID.
func (api *ServerAPI) UpdateRespond(
	ctx context.Context,
	in *tickets.UpdateRespondIn,
) (*emptypb.Empty, error) {
	respondData := entities.UpdateRespondDTO{
		ID:      in.GetID(),
		Price:   in.Price,
		Comment: in.Comment,
	}

	err := api.useCases.UpdateRespond(ctx, respondData)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to update Respond with ID=%d", in.GetID()),
			err,
		)

		switch {
		case errors.As(err, &respondNotFoundError):
			return nil, &customgrpc.BaseError{Status: codes.NotFound, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return &emptypb.Empty{}, nil
}

// DeleteRespond handler deletes Respond with provided ID.
func (api *ServerAPI) DeleteRespond(
	ctx context.Context,
	in *tickets.DeleteRespondIn,
) (*emptypb.Empty, error) {
	err := api.useCases.DeleteRespond(ctx, in.GetID())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to delete Respond with ID=%d", in.GetID()),
			err,
		)

		switch {
		case errors.As(err, &respondNotFoundError):
			return nil, &customgrpc.BaseError{Status: codes.NotFound, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return &emptypb.Empty{}, nil
}

// RespondToTicket handler creates new Respond to Ticket.
func (api *ServerAPI) RespondToTicket(
	ctx context.Context,
	in *tickets.RespondToTicketIn,
) (*tickets.RespondToTicketOut, error) {
	respondData := entities.RawRespondToTicketDTO{
		UserID:   in.GetUserID(),
		TicketID: in.GetTicketID(),
		Price:    in.GetPrice(),
		Comment:  in.Comment,
	}

	respondID, err := api.useCases.RespondToTicket(ctx, respondData)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			"Error occurred while trying to respond to Ticket",
			err,
		)

		switch {
		case errors.As(err, &respondAlreadyExistsError):
			return nil, &customgrpc.BaseError{Status: codes.AlreadyExists, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return &tickets.RespondToTicketOut{RespondID: respondID}, nil
}

// GetRespond handler returns Respond for provided ID.
func (api *ServerAPI) GetRespond(
	ctx context.Context,
	in *tickets.GetRespondIn,
) (*tickets.GetRespondOut, error) {
	respond, err := api.useCases.GetRespondByID(ctx, in.GetID())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to get Respond with ID=%d", in.GetID()),
			err,
		)

		switch {
		case errors.As(err, &respondNotFoundError):
			return nil, &customgrpc.BaseError{Status: codes.NotFound, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return prepareRespondOut(*respond), nil
}

// GetTicketResponds handler returns Responds for provided Ticket ID.
func (api *ServerAPI) GetTicketResponds(
	ctx context.Context,
	in *tickets.GetTicketRespondsIn,
) (*tickets.GetRespondsOut, error) {
	ticketResponds, err := api.useCases.GetTicketResponds(ctx, in.GetTicketID())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf(
				"Error occurred while trying to get Responds for Ticket with ID=%d",
				in.GetTicketID(),
			),
			err,
		)

		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	processedResponds := make([]*tickets.GetRespondOut, len(ticketResponds))
	for i, respond := range ticketResponds {
		processedResponds[i] = prepareRespondOut(respond)
	}

	return &tickets.GetRespondsOut{Responds: processedResponds}, nil
}

// GetUserResponds handler returns Responds for User with provided ID.
func (api *ServerAPI) GetUserResponds(
	ctx context.Context,
	in *tickets.GetUserRespondsIn,
) (*tickets.GetRespondsOut, error) {
	userResponds, err := api.useCases.GetUserResponds(ctx, in.GetUserID())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf(
				"Error occurred while trying to get Responds for User with ID=%d",
				in.GetUserID(),
			),
			err,
		)

		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	processedResponds := make([]*tickets.GetRespondOut, len(userResponds))
	for i, respond := range userResponds {
		processedResponds[i] = prepareRespondOut(respond)
	}

	return &tickets.GetRespondsOut{Responds: processedResponds}, nil
}
