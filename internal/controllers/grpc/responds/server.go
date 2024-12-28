package responds

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
	customgrpc "github.com/DKhorkov/libs/grpc"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/security"
)

// RegisterServer handler (serverAPI) for RespondsServer to gRPC server:.
func RegisterServer(gRPCServer *grpc.Server, useCases interfaces.UseCases, logger *slog.Logger) {
	tickets.RegisterRespondsServiceServer(gRPCServer, &ServerAPI{useCases: useCases, logger: logger})
}

type ServerAPI struct {
	// Helps to test single endpoints, if others is not implemented yet
	tickets.UnimplementedRespondsServiceServer
	useCases interfaces.UseCases
	logger   *slog.Logger
}

// RespondToTicket handler creates new Respond to Ticket.
func (api *ServerAPI) RespondToTicket(
	ctx context.Context,
	in *tickets.RespondToTicketIn,
) (*tickets.RespondToTicketOut, error) {
	respondData := entities.RawRespondToTicketDTO{
		AccessToken: in.GetAccessToken(),
		TicketID:    in.GetTicketID(),
	}

	respondID, err := api.useCases.RespondToTicket(ctx, respondData)
	if err != nil {
		logging.LogErrorContext(ctx, api.logger, "Error occurred while trying to respond to Ticket", err)

		switch {
		case errors.As(err, &security.InvalidJWTError{}):
			return nil, &customgrpc.BaseError{Status: codes.Unauthenticated, Message: err.Error()}
		case errors.As(err, &customerrors.RespondAlreadyExistsError{}):
			return nil, &customgrpc.BaseError{Status: codes.AlreadyExists, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return &tickets.RespondToTicketOut{RespondID: respondID}, nil
}

// GetRespond handler returns Respond for provided ID.
func (api *ServerAPI) GetRespond(ctx context.Context, in *tickets.GetRespondIn) (*tickets.GetRespondOut, error) {
	respond, err := api.useCases.GetRespondByID(ctx, in.GetID())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to get Respond with ID=%d", in.GetID()),
			err,
		)

		switch {
		case errors.As(err, &customerrors.RespondNotFoundError{}):
			return nil, &customgrpc.BaseError{Status: codes.NotFound, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return &tickets.GetRespondOut{
		ID:        respond.ID,
		TicketID:  respond.TicketID,
		MasterID:  respond.MasterID,
		CreatedAt: timestamppb.New(respond.CreatedAt),
		UpdatedAt: timestamppb.New(respond.UpdatedAt),
	}, nil
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
			fmt.Sprintf("Error occurred while trying to get Responds for Ticket with ID=%d", in.GetTicketID()),
			err,
		)

		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	processedResponds := make([]*tickets.GetRespondOut, len(ticketResponds))
	for i, respond := range ticketResponds {
		processedResponds[i] = &tickets.GetRespondOut{
			ID:        respond.ID,
			TicketID:  respond.TicketID,
			MasterID:  respond.MasterID,
			CreatedAt: timestamppb.New(respond.CreatedAt),
			UpdatedAt: timestamppb.New(respond.UpdatedAt),
		}
	}

	return &tickets.GetRespondsOut{Responds: processedResponds}, nil
}

// GetMyResponds handler returns Responds for current User.
func (api *ServerAPI) GetMyResponds(
	ctx context.Context,
	in *tickets.GetMyRespondsIn,
) (*tickets.GetRespondsOut, error) {
	myResponds, err := api.useCases.GetMyResponds(ctx, in.GetAccessToken())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to get Responds for user with AccessToken=%s", in.GetAccessToken()),
			err,
		)

		switch {
		case errors.As(err, &security.InvalidJWTError{}):
			return nil, &customgrpc.BaseError{Status: codes.Unauthenticated, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	processedResponds := make([]*tickets.GetRespondOut, len(myResponds))
	for i, respond := range myResponds {
		processedResponds[i] = &tickets.GetRespondOut{
			ID:        respond.ID,
			TicketID:  respond.TicketID,
			MasterID:  respond.MasterID,
			CreatedAt: timestamppb.New(respond.CreatedAt),
			UpdatedAt: timestamppb.New(respond.UpdatedAt),
		}
	}

	return &tickets.GetRespondsOut{Responds: processedResponds}, nil
}
