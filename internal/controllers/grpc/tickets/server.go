package tickets

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

// RegisterServer handler (serverAPI) for TicketsServer to gRPC server:.
func RegisterServer(gRPCServer *grpc.Server, useCases interfaces.UseCases, logger *slog.Logger) {
	tickets.RegisterTicketsServiceServer(gRPCServer, &ServerAPI{useCases: useCases, logger: logger})
}

type ServerAPI struct {
	// Helps to test single endpoints, if others is not implemented yet
	tickets.UnimplementedTicketsServiceServer
	useCases interfaces.UseCases
	logger   *slog.Logger
}

// CreateTicket handler creates new Ticket.
func (api *ServerAPI) CreateTicket(ctx context.Context, in *tickets.CreateTicketIn) (*tickets.CreateTicketOut, error) {
	ticketData := entities.RawCreateTicketDTO{
		AccessToken: in.GetAccessToken(),
		CategoryID:  in.GetCategoryID(),
		Name:        in.GetName(),
		Description: in.GetDescription(),
		Price:       in.GetPrice(),
		Quantity:    in.GetQuantity(),
		TagsIDs:     in.GetTagIDs(),
	}

	ticketID, err := api.useCases.CreateTicket(ctx, ticketData)
	if err != nil {
		logging.LogErrorContext(ctx, api.logger, "Error occurred while trying to create new Ticket", err)

		switch {
		case errors.As(err, &security.InvalidJWTError{}):
			return nil, &customgrpc.BaseError{Status: codes.Unauthenticated, Message: err.Error()}
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

	return &tickets.GetTicketOut{
		ID:          ticket.ID,
		UserID:      ticket.UserID,
		Name:        ticket.Name,
		Description: ticket.Description,
		Price:       ticket.Price,
		Quantity:    ticket.Quantity,
		CategoryID:  ticket.CategoryID,
		TagIDs:      ticket.TagIDs,
		CreatedAt:   timestamppb.New(ticket.CreatedAt),
		UpdatedAt:   timestamppb.New(ticket.UpdatedAt),
	}, nil
}

// GetTickets handler returns all Tickets.
func (api *ServerAPI) GetTickets(ctx context.Context, in *tickets.GetTicketsIn) (*tickets.GetTicketsOut, error) {
	allTickets, err := api.useCases.GetAllTickets(ctx)
	if err != nil {
		logging.LogErrorContext(ctx, api.logger, "Error occurred while trying to get all Tickets", err)
		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	processedTickets := make([]*tickets.GetTicketOut, len(allTickets))
	for i, ticket := range allTickets {
		processedTickets[i] = &tickets.GetTicketOut{
			ID:          ticket.ID,
			UserID:      ticket.UserID,
			Name:        ticket.Name,
			Description: ticket.Description,
			Price:       ticket.Price,
			Quantity:    ticket.Quantity,
			CategoryID:  ticket.CategoryID,
			TagIDs:      ticket.TagIDs,
			CreatedAt:   timestamppb.New(ticket.CreatedAt),
			UpdatedAt:   timestamppb.New(ticket.UpdatedAt),
		}
	}

	return &tickets.GetTicketsOut{Tickets: processedTickets}, nil
}

// GetMyTickets handler returns Tickets for current User.
func (api *ServerAPI) GetMyTickets(ctx context.Context, in *tickets.GetMyTicketsIn) (*tickets.GetTicketsOut, error) {
	myTickets, err := api.useCases.GetMyTickets(ctx, in.GetAccessToken())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to get Tickets for user with AccessToken=%s", in.GetAccessToken()),
			err,
		)

		switch {
		case errors.As(err, &security.InvalidJWTError{}):
			return nil, &customgrpc.BaseError{Status: codes.Unauthenticated, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	processedTickets := make([]*tickets.GetTicketOut, len(myTickets))
	for i, ticket := range myTickets {
		processedTickets[i] = &tickets.GetTicketOut{
			ID:          ticket.ID,
			UserID:      ticket.UserID,
			Name:        ticket.Name,
			Description: ticket.Description,
			Price:       ticket.Price,
			Quantity:    ticket.Quantity,
			CategoryID:  ticket.CategoryID,
			TagIDs:      ticket.TagIDs,
			CreatedAt:   timestamppb.New(ticket.CreatedAt),
			UpdatedAt:   timestamppb.New(ticket.UpdatedAt),
		}
	}

	return &tickets.GetTicketsOut{Tickets: processedTickets}, nil
}
