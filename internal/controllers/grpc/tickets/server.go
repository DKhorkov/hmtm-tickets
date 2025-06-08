package tickets

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
	ticketNotFoundError      = &customerrors.TicketNotFoundError{}
	ticketAlreadyExistsError = &customerrors.TicketAlreadyExistsError{}
	categoryNotFoundError    = &customerrors.CategoryNotFoundError{}
	tagNotFoundError         = &customerrors.TagNotFoundError{}
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

func (api *ServerAPI) CountTickets(ctx context.Context, in *tickets.CountTicketsIn) (*tickets.CountOut, error) {
	var filters *entities.TicketsFilters
	if in.GetFilters() != nil {
		filters = &entities.TicketsFilters{
			Search:              in.Filters.Search,
			PriceCeil:           in.Filters.PriceCeil,
			PriceFloor:          in.Filters.PriceFloor,
			QuantityFloor:       in.Filters.QuantityFloor,
			CategoryIDs:         in.Filters.CategoryIDs,
			TagIDs:              in.Filters.TagIDs,
			CreatedAtOrderByAsc: in.Filters.CreatedAtOrderByAsc,
		}
	}

	count, err := api.useCases.CountTickets(ctx, filters)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			"Error occurred while trying to count Tickets",
			err,
		)

		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	return &tickets.CountOut{Count: count}, nil
}

func (api *ServerAPI) CountUserTickets(ctx context.Context, in *tickets.CountUserTicketsIn) (*tickets.CountOut, error) {
	var filters *entities.TicketsFilters
	if in.GetFilters() != nil {
		filters = &entities.TicketsFilters{
			Search:              in.Filters.Search,
			PriceCeil:           in.Filters.PriceCeil,
			PriceFloor:          in.Filters.PriceFloor,
			QuantityFloor:       in.Filters.QuantityFloor,
			CategoryIDs:         in.Filters.CategoryIDs,
			TagIDs:              in.Filters.TagIDs,
			CreatedAtOrderByAsc: in.Filters.CreatedAtOrderByAsc,
		}
	}

	count, err := api.useCases.CountUserTickets(ctx, in.GetUserID(), filters)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to count Tickets for User with ID=%d", in.GetUserID()),
			err,
		)

		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	return &tickets.CountOut{Count: count}, nil
}

// DeleteTicket handler deletes Ticket with provided ID.
func (api *ServerAPI) DeleteTicket(
	ctx context.Context,
	in *tickets.DeleteTicketIn,
) (*emptypb.Empty, error) {
	if err := api.useCases.DeleteTicket(ctx, in.GetID()); err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to delete Ticket with ID=%d", in.GetID()),
			err,
		)

		switch {
		case errors.As(err, &ticketNotFoundError):
			return nil, &customgrpc.BaseError{Status: codes.NotFound, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return &emptypb.Empty{}, nil
}

// UpdateTicket handler updates Ticket with provided ID.
func (api *ServerAPI) UpdateTicket(
	ctx context.Context,
	in *tickets.UpdateTicketIn,
) (*emptypb.Empty, error) {
	ticketData := entities.RawUpdateTicketDTO{
		ID:          in.GetID(),
		TagIDs:      in.GetTagIDs(),
		Attachments: in.GetAttachments(),
	}

	if in != nil {
		ticketData.CategoryID = in.CategoryID
		ticketData.Name = in.Name
		ticketData.Description = in.Description
		ticketData.Price = in.Price
		ticketData.Quantity = in.Quantity
	}

	if err := api.useCases.UpdateTicket(ctx, ticketData); err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to update Ticket with ID=%d", in.GetID()),
			err,
		)

		switch {
		case errors.As(err, &ticketNotFoundError),
			errors.As(err, &categoryNotFoundError),
			errors.As(err, &tagNotFoundError):
			return nil, &customgrpc.BaseError{Status: codes.NotFound, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return &emptypb.Empty{}, nil
}

// CreateTicket handler creates new Ticket.
func (api *ServerAPI) CreateTicket(
	ctx context.Context,
	in *tickets.CreateTicketIn,
) (*tickets.CreateTicketOut, error) {
	ticketData := entities.CreateTicketDTO{
		UserID:      in.GetUserID(),
		CategoryID:  in.GetCategoryID(),
		Name:        in.GetName(),
		Description: in.GetDescription(),
		Quantity:    in.GetQuantity(),
		TagIDs:      in.GetTagIDs(),
		Attachments: in.GetAttachments(),
	}

	if in != nil {
		ticketData.Price = in.Price
	}

	ticketID, err := api.useCases.CreateTicket(ctx, ticketData)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			"Error occurred while trying to create new Ticket",
			err,
		)

		switch {
		case errors.As(err, &ticketAlreadyExistsError),
			errors.As(err, &categoryNotFoundError),
			errors.As(err, &tagNotFoundError):
			return nil, &customgrpc.BaseError{Status: codes.AlreadyExists, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return &tickets.CreateTicketOut{TicketID: ticketID}, nil
}

// GetTicket handler returns Ticket for provided ID.
func (api *ServerAPI) GetTicket(
	ctx context.Context,
	in *tickets.GetTicketIn,
) (*tickets.GetTicketOut, error) {
	ticket, err := api.useCases.GetTicketByID(ctx, in.GetID())
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf("Error occurred while trying to get Ticket with ID=%d", in.GetID()),
			err,
		)

		switch {
		case errors.As(err, &ticketNotFoundError):
			return nil, &customgrpc.BaseError{Status: codes.NotFound, Message: err.Error()}
		default:
			return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
		}
	}

	return mapTicketToOut(*ticket), nil
}

// GetTickets handler returns all Tickets.
func (api *ServerAPI) GetTickets(
	ctx context.Context,
	in *tickets.GetTicketsIn,
) (*tickets.GetTicketsOut, error) {
	var filters *entities.TicketsFilters
	if in.GetFilters() != nil {
		filters = &entities.TicketsFilters{
			Search:              in.Filters.Search,
			PriceCeil:           in.Filters.PriceCeil,
			PriceFloor:          in.Filters.PriceFloor,
			QuantityFloor:       in.Filters.QuantityFloor,
			CategoryIDs:         in.Filters.CategoryIDs,
			TagIDs:              in.Filters.TagIDs,
			CreatedAtOrderByAsc: in.Filters.CreatedAtOrderByAsc,
		}
	}

	var pagination *entities.Pagination
	if in.GetPagination() != nil {
		pagination = &entities.Pagination{
			Limit:  in.Pagination.Limit,
			Offset: in.Pagination.Offset,
		}
	}

	allTickets, err := api.useCases.GetTickets(ctx, pagination, filters)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			"Error occurred while trying to get all Tickets",
			err,
		)

		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	processedTickets := make([]*tickets.GetTicketOut, len(allTickets))
	for i, ticket := range allTickets {
		processedTickets[i] = mapTicketToOut(ticket)
	}

	return &tickets.GetTicketsOut{Tickets: processedTickets}, nil
}

// GetUserTickets handler returns Tickets for User with provided ID.
func (api *ServerAPI) GetUserTickets(
	ctx context.Context,
	in *tickets.GetUserTicketsIn,
) (*tickets.GetTicketsOut, error) {
	var filters *entities.TicketsFilters
	if in.GetFilters() != nil {
		filters = &entities.TicketsFilters{
			Search:              in.Filters.Search,
			PriceCeil:           in.Filters.PriceCeil,
			PriceFloor:          in.Filters.PriceFloor,
			QuantityFloor:       in.Filters.QuantityFloor,
			CategoryIDs:         in.Filters.CategoryIDs,
			TagIDs:              in.Filters.TagIDs,
			CreatedAtOrderByAsc: in.Filters.CreatedAtOrderByAsc,
		}
	}

	var pagination *entities.Pagination
	if in.GetPagination() != nil {
		pagination = &entities.Pagination{
			Limit:  in.Pagination.Limit,
			Offset: in.Pagination.Offset,
		}
	}

	userTickets, err := api.useCases.GetUserTickets(ctx, in.GetUserID(), pagination, filters)
	if err != nil {
		logging.LogErrorContext(
			ctx,
			api.logger,
			fmt.Sprintf(
				"Error occurred while trying to get Tickets for User with ID=%d",
				in.GetUserID(),
			),
			err,
		)

		return nil, &customgrpc.BaseError{Status: codes.Internal, Message: err.Error()}
	}

	processedTickets := make([]*tickets.GetTicketOut, len(userTickets))
	for i, ticket := range userTickets {
		processedTickets[i] = mapTicketToOut(ticket)
	}

	return &tickets.GetTicketsOut{Tickets: processedTickets}, nil
}
