package tickets

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	customgrpc "github.com/DKhorkov/libs/grpc"
	"github.com/DKhorkov/libs/logging"

	"github.com/DKhorkov/hmtm-tickets/api/protobuf/generated/go/tickets"
	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	customerrors "github.com/DKhorkov/hmtm-tickets/internal/errors"
	"github.com/DKhorkov/hmtm-tickets/internal/interfaces"
)

// RegisterServer handler (serverAPI) for TicketsServer to gRPC server:.
func RegisterServer(gRPCServer *grpc.Server, useCases interfaces.UseCases, logger logging.Logger) {
	tickets.RegisterTicketsServiceServer(gRPCServer, &ServerAPI{useCases: useCases, logger: logger})
}

type ServerAPI struct {
	// Helps to test single endpoints, if others is not implemented yet
	tickets.UnimplementedTicketsServiceServer
	useCases interfaces.UseCases
	logger   logging.Logger
}

// CreateTicket handler creates new Ticket.
func (api *ServerAPI) CreateTicket(ctx context.Context, in *tickets.CreateTicketIn) (*tickets.CreateTicketOut, error) {
	ticketData := entities.CreateTicketDTO{
		UserID:      in.GetUserID(),
		CategoryID:  in.GetCategoryID(),
		Name:        in.GetName(),
		Description: in.GetDescription(),
		Price:       in.GetPrice(),
		Quantity:    in.GetQuantity(),
		TagIDs:      in.GetTagIDs(),
		Attachments: in.GetAttachments(),
	}

	ticketID, err := api.useCases.CreateTicket(ctx, ticketData)
	if err != nil {
		logging.LogErrorContext(ctx, api.logger, "Error occurred while trying to create new Ticket", err)

		switch {
		case errors.As(err, &customerrors.TicketAlreadyExistsError{}):
			return nil, &customgrpc.BaseError{Status: codes.AlreadyExists, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return &tickets.CreateTicketOut{TicketID: ticketID}, nil
}

// GetTicket handler returns Ticket for provided ID.
func (api *ServerAPI) GetTicket(ctx context.Context, in *tickets.GetTicketIn) (*tickets.GetTicketOut, error) {
	ticket, err := api.useCases.GetTicketByID(ctx, in.GetID())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to get Ticket with ID=%d", in.GetID()),
			err,
		)

		switch {
		case errors.As(err, &customerrors.TicketNotFoundError{}):
			return nil, &customgrpc.BaseError{Status: codes.NotFound, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return prepareTicketOut(*ticket), nil
}

// GetTickets handler returns all Tickets.
func (api *ServerAPI) GetTickets(ctx context.Context, _ *emptypb.Empty) (*tickets.GetTicketsOut, error) {
	allTickets, err := api.useCases.GetAllTickets(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, api.logger, "Error occurred while trying to get all Tickets", err)
		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	processedTickets := make([]*tickets.GetTicketOut, len(allTickets))
	for i, ticket := range allTickets {
		processedTickets[i] = prepareTicketOut(ticket)
	}

	return &tickets.GetTicketsOut{Tickets: processedTickets}, nil
}

// GetUserTickets handler returns Tickets for User with provided ID.
func (api *ServerAPI) GetUserTickets(
	ctx context.Context,
	in *tickets.GetUserTicketsIn,
) (*tickets.GetTicketsOut, error) {
	userTickets, err := api.useCases.GetUserTickets(ctx, in.GetUserID())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to get Tickets for User with ID=%d", in.GetUserID()),
			err,
		)

		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	processedTickets := make([]*tickets.GetTicketOut, len(userTickets))
	for i, ticket := range userTickets {
		processedTickets[i] = prepareTicketOut(ticket)
	}

	return &tickets.GetTicketsOut{Tickets: processedTickets}, nil
}
