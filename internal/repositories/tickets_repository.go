package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/tracing"

	sq "github.com/Masterminds/squirrel"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
)

const (
	selectAllColumns                   = "*"
	selectCount                        = "COUNT(*)"
	ticketsTableName                   = "tickets"
	ticketsAndTagsAssociationTableName = "tickets_tags_associations"
	ticketsAttachmentsTableName        = "tickets_attachments"
	idColumnName                       = "id"
	categoryIDColumnName               = "category_id"
	ticketNameColumnName               = "name"
	ticketDescriptionColumnName        = "description"
	ticketPriceColumnName              = "price"
	ticketQuantityColumnName           = "quantity"
	ticketIDColumnName                 = "ticket_id"
	tagIDColumnName                    = "tag_id"
	userIDColumnName                   = "user_id"
	attachmentLinkColumnName           = "link"
	returningIDSuffix                  = "RETURNING id"
	createdAtColumnName                = "created_at"
	updatedAtColumnName                = "updated_at"
	desc                               = "DESC"
	asc                                = "ASC"
)

type TicketsRepository struct {
	dbConnector   db.Connector
	logger        logging.Logger
	traceProvider tracing.Provider
	spanConfig    tracing.SpanConfig
}

func NewTicketsRepository(
	dbConnector db.Connector,
	logger logging.Logger,
	traceProvider tracing.Provider,
	spanConfig tracing.SpanConfig,
) *TicketsRepository {
	return &TicketsRepository{
		dbConnector:   dbConnector,
		logger:        logger,
		traceProvider: traceProvider,
		spanConfig:    spanConfig,
	}
}

func (repo *TicketsRepository) CreateTicket(
	ctx context.Context,
	ticketData entities.CreateTicketDTO,
) (uint64, error) {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	transaction, err := repo.dbConnector.Transaction(ctx)
	if err != nil {
		return 0, err
	}

	// Rollback transaction according Go best practises https://go.dev/doc/database/execute-transactions.
	defer func() {
		if err = transaction.Rollback(); err != nil {
			logging.LogErrorContext(ctx, repo.logger, "failed to rollback db transaction", err)
		}
	}()

	stmt, params, err := sq.
		Insert(ticketsTableName).
		Columns(
			userIDColumnName,
			categoryIDColumnName,
			ticketNameColumnName,
			ticketDescriptionColumnName,
			ticketPriceColumnName,
			ticketQuantityColumnName,
		).
		Values(
			ticketData.UserID,
			ticketData.CategoryID,
			ticketData.Name,
			ticketData.Description,
			ticketData.Price,
			ticketData.Quantity,
		).
		Suffix(returningIDSuffix).
		PlaceholderFormat(sq.Dollar). // pq postgres driver works only with $ placeholders
		ToSql()
	if err != nil {
		return 0, err
	}

	var ticketID uint64
	if err = transaction.QueryRowContext(ctx, stmt, params...).Scan(&ticketID); err != nil {
		return 0, err
	}

	if err != nil {
		return 0, err
	}

	if len(ticketData.TagIDs) > 0 {
		builder := sq.Insert(ticketsAndTagsAssociationTableName).
			Columns(ticketIDColumnName, tagIDColumnName)
		for _, tagID := range ticketData.TagIDs {
			builder = builder.Values(ticketID, tagID)
		}

		if stmt, params, err = builder.PlaceholderFormat(sq.Dollar).ToSql(); err != nil {
			return 0, err
		}

		if _, err = transaction.ExecContext(ctx, stmt, params...); err != nil {
			return 0, err
		}
	}

	if len(ticketData.Attachments) > 0 {
		builder := sq.Insert(ticketsAttachmentsTableName).
			Columns(ticketIDColumnName, attachmentLinkColumnName)
		for _, attachment := range ticketData.Attachments {
			builder = builder.Values(ticketID, attachment)
		}

		if stmt, params, err = builder.PlaceholderFormat(sq.Dollar).ToSql(); err != nil {
			return 0, err
		}

		if _, err = transaction.ExecContext(ctx, stmt, params...); err != nil {
			return 0, err
		}
	}

	err = transaction.Commit()
	if err != nil {
		return 0, err
	}

	return ticketID, nil
}

func (repo *TicketsRepository) GetTicketByID(
	ctx context.Context,
	id uint64,
) (*entities.Ticket, error) {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	connection, err := repo.dbConnector.Connection(ctx)
	if err != nil {
		return nil, err
	}

	defer db.CloseConnectionContext(ctx, connection, repo.logger)

	stmt, params, err := sq.
		Select(selectAllColumns).
		From(ticketsTableName).
		Where(sq.Eq{idColumnName: id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	ticket := &entities.Ticket{}
	columns := db.GetEntityColumns(ticket) // Only pointer to use rows.Scan() successfully
	columns = columns[:len(columns)-2]     // Not to paste TagIDs and Attachments fields to Scan function.

	if err = connection.QueryRowContext(ctx, stmt, params...).Scan(columns...); err != nil {
		return nil, err
	}

	tagsIDs, err := repo.getTicketTagsIDs(ctx, ticket.ID, connection)
	if err != nil {
		return nil, err
	}

	ticket.TagIDs = tagsIDs

	attachments, err := repo.getTicketAttachments(ctx, ticket.ID, connection)
	if err != nil {
		return nil, err
	}

	ticket.Attachments = attachments

	return ticket, nil
}

func (repo *TicketsRepository) GetTickets(
	ctx context.Context,
	pagination *entities.Pagination,
	filters *entities.TicketsFilters,
) ([]entities.Ticket, error) {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	connection, err := repo.dbConnector.Connection(ctx)
	if err != nil {
		return nil, err
	}

	defer db.CloseConnectionContext(ctx, connection, repo.logger)

	builder := sq.
		Select(selectAllColumns).
		From(ticketsTableName).
		PlaceholderFormat(sq.Dollar)

	if filters != nil && filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + strings.ToLower(*filters.Search) + "%"
		builder = builder.
			Where(
				sq.Or{
					sq.ILike{
						fmt.Sprintf(
							"%s.%s",
							ticketsTableName,
							ticketNameColumnName,
						): searchTerm,
					},
					sq.ILike{
						fmt.Sprintf(
							"%s.%s",
							ticketsTableName,
							ticketDescriptionColumnName,
						): searchTerm,
					},
				},
			)
	}

	if filters != nil && (filters.PriceFloor != nil || filters.PriceCeil != nil) {
		priceConditions := sq.And{}
		if filters.PriceFloor != nil {
			priceConditions = append(
				priceConditions,
				sq.GtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketPriceColumnName,
					): *filters.PriceFloor,
				},
			)
		}

		if filters.PriceCeil != nil {
			priceConditions = append(
				priceConditions,
				sq.LtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketPriceColumnName,
					): *filters.PriceCeil,
				},
			)
		}

		builder = builder.Where(priceConditions)
	}

	if filters != nil && filters.QuantityFloor != nil {
		builder = builder.
			Where(
				sq.GtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketQuantityColumnName,
					): *filters.QuantityFloor,
				},
			)
	}

	if filters != nil && filters.CategoryIDs != nil {
		builder = builder.
			Where(
				sq.Eq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						categoryIDColumnName,
					): filters.CategoryIDs,
				},
			)
	}

	if filters != nil && len(filters.TagIDs) > 0 {
		for _, tagID := range filters.TagIDs {
			builder = builder.
				Where(
					sq.Expr(
						fmt.Sprintf(
							"EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.%s AND %s.%s = ?)",
							ticketsAndTagsAssociationTableName,
							ticketsAndTagsAssociationTableName,
							ticketIDColumnName,
							ticketsTableName,
							idColumnName,
							ticketsAndTagsAssociationTableName,
							tagIDColumnName,
						),
						tagID,
					),
				)
		}
	}

	createdAtOrder := desc
	if filters != nil && filters.CreatedAtOrderByAsc != nil && *filters.CreatedAtOrderByAsc {
		createdAtOrder = asc
	}

	builder = builder.
		OrderBy(
			fmt.Sprintf(
				"%s.%s %s",
				ticketsTableName,
				createdAtColumnName,
				createdAtOrder,
			),
		)

	if pagination != nil && pagination.Limit != nil {
		builder = builder.Limit(*pagination.Limit)
	}

	if pagination != nil && pagination.Offset != nil {
		builder = builder.Offset(*pagination.Offset)
	}

	stmt, params, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := connection.QueryContext(
		ctx,
		stmt,
		params...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			logging.LogErrorContext(
				ctx,
				repo.logger,
				"error during closing SQL rows",
				err,
			)
		}
	}()

	var tickets []entities.Ticket

	for rows.Next() {
		ticket := entities.Ticket{}
		columns := db.GetEntityColumns(&ticket) // Only pointer to use rows.Scan() successfully
		columns = columns[:len(columns)-2]      // Not to paste TagIDs and Attachments fields to Scan function.

		err = rows.Scan(columns...)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Reading Tags and Attachments for each Ticket in new circle due
	// to next error: https://github.com/lib/pq/issues/635
	// Using ticket index to avoid range iter semantics error, via using copied variable.
	for i, ticket := range tickets {
		tagsIDs, err := repo.getTicketTagsIDs(ctx, ticket.ID, connection)
		if err != nil {
			return nil, err
		}

		tickets[i].TagIDs = tagsIDs

		attachments, err := repo.getTicketAttachments(ctx, ticket.ID, connection)
		if err != nil {
			return nil, err
		}

		tickets[i].Attachments = attachments
	}

	return tickets, nil
}

func (repo *TicketsRepository) CountTickets(ctx context.Context, filters *entities.TicketsFilters) (uint64, error) {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	connection, err := repo.dbConnector.Connection(ctx)
	if err != nil {
		return 0, err
	}

	defer db.CloseConnectionContext(ctx, connection, repo.logger)

	builder := sq.
		Select(selectCount).
		From(ticketsTableName).
		PlaceholderFormat(sq.Dollar)

	if filters != nil && filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + strings.ToLower(*filters.Search) + "%"
		builder = builder.
			Where(
				sq.Or{
					sq.ILike{
						fmt.Sprintf(
							"%s.%s",
							ticketsTableName,
							ticketNameColumnName,
						): searchTerm,
					},
					sq.ILike{
						fmt.Sprintf(
							"%s.%s",
							ticketsTableName,
							ticketDescriptionColumnName,
						): searchTerm,
					},
				},
			)
	}

	if filters != nil && (filters.PriceFloor != nil || filters.PriceCeil != nil) {
		priceConditions := sq.And{}
		if filters.PriceFloor != nil {
			priceConditions = append(
				priceConditions,
				sq.GtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketPriceColumnName,
					): *filters.PriceFloor,
				},
			)
		}

		if filters.PriceCeil != nil {
			priceConditions = append(
				priceConditions,
				sq.LtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketPriceColumnName,
					): *filters.PriceCeil,
				},
			)
		}

		builder = builder.Where(priceConditions)
	}

	if filters != nil && filters.QuantityFloor != nil {
		builder = builder.
			Where(
				sq.GtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketQuantityColumnName,
					): *filters.QuantityFloor,
				},
			)
	}

	if filters != nil && filters.CategoryIDs != nil {
		builder = builder.
			Where(
				sq.Eq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						categoryIDColumnName,
					): filters.CategoryIDs,
				},
			)
	}

	if filters != nil && len(filters.TagIDs) > 0 {
		for _, tagID := range filters.TagIDs {
			builder = builder.
				Where(
					sq.Expr(
						fmt.Sprintf(
							"EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.%s AND %s.%s = ?)",
							ticketsAndTagsAssociationTableName,
							ticketsAndTagsAssociationTableName,
							ticketIDColumnName,
							ticketsTableName,
							idColumnName,
							ticketsAndTagsAssociationTableName,
							tagIDColumnName,
						),
						tagID,
					),
				)
		}
	}

	// Для запросов COUNT сортировка не нужна, поэтому параметр CreatedAtOrderByAsc не используется
	stmt, params, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var count uint64
	if err = connection.QueryRowContext(ctx, stmt, params...).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *TicketsRepository) GetUserTickets(
	ctx context.Context,
	userID uint64,
	pagination *entities.Pagination,
	filters *entities.TicketsFilters,
) ([]entities.Ticket, error) {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	connection, err := repo.dbConnector.Connection(ctx)
	if err != nil {
		return nil, err
	}

	defer db.CloseConnectionContext(ctx, connection, repo.logger)

	builder := sq.
		Select(selectAllColumns).
		From(ticketsTableName).
		Where(sq.Eq{userIDColumnName: userID}).
		PlaceholderFormat(sq.Dollar)

	if filters != nil && filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + strings.ToLower(*filters.Search) + "%"
		builder = builder.
			Where(
				sq.Or{
					sq.ILike{
						fmt.Sprintf(
							"%s.%s",
							ticketsTableName,
							ticketNameColumnName,
						): searchTerm,
					},
					sq.ILike{
						fmt.Sprintf(
							"%s.%s",
							ticketsTableName,
							ticketDescriptionColumnName,
						): searchTerm,
					},
				},
			)
	}

	if filters != nil && (filters.PriceFloor != nil || filters.PriceCeil != nil) {
		priceConditions := sq.And{}
		if filters.PriceFloor != nil {
			priceConditions = append(
				priceConditions,
				sq.GtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketPriceColumnName,
					): *filters.PriceFloor,
				},
			)
		}

		if filters.PriceCeil != nil {
			priceConditions = append(
				priceConditions,
				sq.LtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketPriceColumnName,
					): *filters.PriceCeil,
				},
			)
		}

		builder = builder.Where(priceConditions)
	}

	if filters != nil && filters.QuantityFloor != nil {
		builder = builder.
			Where(
				sq.GtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketQuantityColumnName,
					): *filters.QuantityFloor,
				},
			)
	}

	if filters != nil && filters.CategoryIDs != nil {
		builder = builder.
			Where(
				sq.Eq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						categoryIDColumnName,
					): filters.CategoryIDs,
				},
			)
	}

	if filters != nil && len(filters.TagIDs) > 0 {
		for _, tagID := range filters.TagIDs {
			builder = builder.
				Where(
					sq.Expr(
						fmt.Sprintf(
							"EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.%s AND %s.%s = ?)",
							ticketsAndTagsAssociationTableName,
							ticketsAndTagsAssociationTableName,
							ticketIDColumnName,
							ticketsTableName,
							idColumnName,
							ticketsAndTagsAssociationTableName,
							tagIDColumnName,
						),
						tagID,
					),
				)
		}
	}

	createdAtOrder := desc
	if filters != nil && filters.CreatedAtOrderByAsc != nil && *filters.CreatedAtOrderByAsc {
		createdAtOrder = asc
	}

	builder = builder.
		OrderBy(
			fmt.Sprintf(
				"%s.%s %s",
				ticketsTableName,
				createdAtColumnName,
				createdAtOrder,
			),
		)

	if pagination != nil && pagination.Limit != nil {
		builder = builder.Limit(*pagination.Limit)
	}

	if pagination != nil && pagination.Offset != nil {
		builder = builder.Offset(*pagination.Offset)
	}

	stmt, params, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := connection.QueryContext(
		ctx,
		stmt,
		params...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			logging.LogErrorContext(
				ctx,
				repo.logger,
				"error during closing SQL rows",
				err,
			)
		}
	}()

	var tickets []entities.Ticket

	for rows.Next() {
		ticket := entities.Ticket{}
		columns := db.GetEntityColumns(&ticket) // Only pointer to use rows.Scan() successfully
		columns = columns[:len(columns)-2]      // Not to paste TagIDs and Attachments fields to Scan function.

		err = rows.Scan(columns...)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Reading Tags and Attachments for each Ticket in new circle due
	// to next error: https://github.com/lib/pq/issues/635
	// Using ticket index to avoid range iter semantics error, via using copied variable.
	for i, ticket := range tickets {
		tagsIDs, err := repo.getTicketTagsIDs(ctx, ticket.ID, connection)
		if err != nil {
			return nil, err
		}

		tickets[i].TagIDs = tagsIDs

		attachments, err := repo.getTicketAttachments(ctx, ticket.ID, connection)
		if err != nil {
			return nil, err
		}

		tickets[i].Attachments = attachments
	}

	return tickets, nil
}

func (repo *TicketsRepository) CountUserTickets(
	ctx context.Context,
	userID uint64,
	filters *entities.TicketsFilters,
) (uint64, error) {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	connection, err := repo.dbConnector.Connection(ctx)
	if err != nil {
		return 0, err
	}

	defer db.CloseConnectionContext(ctx, connection, repo.logger)

	builder := sq.
		Select(selectCount).
		From(ticketsTableName).
		Where(sq.Eq{userIDColumnName: userID}).
		PlaceholderFormat(sq.Dollar)

	if filters != nil && filters.Search != nil && *filters.Search != "" {
		searchTerm := "%" + strings.ToLower(*filters.Search) + "%"
		builder = builder.
			Where(
				sq.Or{
					sq.ILike{
						fmt.Sprintf(
							"%s.%s",
							ticketsTableName,
							ticketNameColumnName,
						): searchTerm,
					},
					sq.ILike{
						fmt.Sprintf(
							"%s.%s",
							ticketsTableName,
							ticketDescriptionColumnName,
						): searchTerm,
					},
				},
			)
	}

	if filters != nil && (filters.PriceFloor != nil || filters.PriceCeil != nil) {
		priceConditions := sq.And{}
		if filters.PriceFloor != nil {
			priceConditions = append(
				priceConditions,
				sq.GtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketPriceColumnName,
					): *filters.PriceFloor,
				},
			)
		}

		if filters.PriceCeil != nil {
			priceConditions = append(
				priceConditions,
				sq.LtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketPriceColumnName,
					): *filters.PriceCeil,
				},
			)
		}

		builder = builder.Where(priceConditions)
	}

	if filters != nil && filters.QuantityFloor != nil {
		builder = builder.
			Where(
				sq.GtOrEq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						ticketQuantityColumnName,
					): *filters.QuantityFloor,
				},
			)
	}

	if filters != nil && filters.CategoryIDs != nil {
		builder = builder.
			Where(
				sq.Eq{
					fmt.Sprintf(
						"%s.%s",
						ticketsTableName,
						categoryIDColumnName,
					): filters.CategoryIDs,
				},
			)
	}

	if filters != nil && len(filters.TagIDs) > 0 {
		for _, tagID := range filters.TagIDs {
			builder = builder.
				Where(
					sq.Expr(
						fmt.Sprintf(
							"EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.%s AND %s.%s = ?)",
							ticketsAndTagsAssociationTableName,
							ticketsAndTagsAssociationTableName,
							ticketIDColumnName,
							ticketsTableName,
							idColumnName,
							ticketsAndTagsAssociationTableName,
							tagIDColumnName,
						),
						tagID,
					),
				)
		}
	}

	// Для запросов COUNT сортировка не нужна, поэтому параметр CreatedAtOrderByAsc не используется
	stmt, params, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var count uint64
	if err = connection.QueryRowContext(ctx, stmt, params...).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *TicketsRepository) DeleteTicket(ctx context.Context, id uint64) error {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	connection, err := repo.dbConnector.Connection(ctx)
	if err != nil {
		return err
	}

	defer db.CloseConnectionContext(ctx, connection, repo.logger)

	stmt, params, err := sq.
		Delete(ticketsTableName).
		Where(sq.Eq{idColumnName: id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	_, err = connection.ExecContext(
		ctx,
		stmt,
		params...,
	)

	return err
}

func (repo *TicketsRepository) UpdateTicket(
	ctx context.Context,
	ticketData entities.UpdateTicketDTO,
) error {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	transaction, err := repo.dbConnector.Transaction(ctx)
	if err != nil {
		return err
	}

	// Rollback transaction according Go best practises https://go.dev/doc/database/execute-transactions.
	defer func() {
		if err = transaction.Rollback(); err != nil {
			logging.LogErrorContext(ctx, repo.logger, "failed to rollback db transaction", err)
		}
	}()

	builder := sq.
		Update(ticketsTableName).
		Where(sq.Eq{idColumnName: ticketData.ID}).
		Set(ticketPriceColumnName, ticketData.Price).
		// Update every time, because field is nullable
		PlaceholderFormat(sq.Dollar)
	// pq postgres driver works only with $ placeholders

	if ticketData.CategoryID != nil {
		builder = builder.Set(categoryIDColumnName, ticketData.CategoryID)
	}

	if ticketData.Name != nil {
		builder = builder.Set(ticketNameColumnName, ticketData.Name)
	}

	if ticketData.Description != nil {
		builder = builder.Set(ticketDescriptionColumnName, ticketData.Description)
	}

	if ticketData.Quantity != nil {
		builder = builder.Set(ticketQuantityColumnName, ticketData.Quantity)
	}

	stmt, params, err := builder.ToSql()
	if err != nil {
		return err
	}

	if _, err = transaction.ExecContext(ctx, stmt, params...); err != nil {
		return err
	}

	if len(ticketData.TagIDsToAdd) > 0 {
		builder := sq.Insert(ticketsAndTagsAssociationTableName).
			Columns(ticketIDColumnName, tagIDColumnName)
		for _, tagID := range ticketData.TagIDsToAdd {
			builder = builder.Values(ticketData.ID, tagID)
		}

		if stmt, params, err = builder.PlaceholderFormat(sq.Dollar).ToSql(); err != nil {
			return err
		}

		if _, err = transaction.ExecContext(ctx, stmt, params...); err != nil {
			return err
		}
	}

	if len(ticketData.TagIDsToDelete) > 0 {
		stmt, params, err = sq.
			Delete(ticketsAndTagsAssociationTableName).
			Where(
				sq.And{
					sq.Eq{ticketIDColumnName: ticketData.ID},
					sq.Eq{tagIDColumnName: ticketData.TagIDsToDelete},
				},
			).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return err
		}

		if _, err = transaction.ExecContext(ctx, stmt, params...); err != nil {
			return err
		}
	}

	if len(ticketData.AttachmentsToAdd) > 0 {
		builder := sq.Insert(ticketsAttachmentsTableName).
			Columns(ticketIDColumnName, attachmentLinkColumnName)
		for _, attachment := range ticketData.AttachmentsToAdd {
			builder = builder.Values(ticketData.ID, attachment)
		}

		if stmt, params, err = builder.PlaceholderFormat(sq.Dollar).ToSql(); err != nil {
			return err
		}

		if _, err = transaction.ExecContext(ctx, stmt, params...); err != nil {
			return err
		}
	}

	if len(ticketData.AttachmentIDsToDelete) > 0 {
		stmt, params, err = sq.
			Delete(ticketsAttachmentsTableName).
			Where(sq.Eq{idColumnName: ticketData.AttachmentIDsToDelete}).
			PlaceholderFormat(sq.Dollar).
			ToSql()
		if err != nil {
			return err
		}

		if _, err = transaction.ExecContext(ctx, stmt, params...); err != nil {
			return err
		}
	}

	return transaction.Commit()
}

func (repo *TicketsRepository) getTicketTagsIDs(
	ctx context.Context,
	ticketID uint64,
	connection *sql.Conn,
) ([]uint32, error) {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	stmt, params, err := sq.
		Select(tagIDColumnName).
		From(ticketsAndTagsAssociationTableName).
		Where(sq.Eq{ticketIDColumnName: ticketID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := connection.QueryContext(
		ctx,
		stmt,
		params...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			logging.LogErrorContext(
				ctx,
				repo.logger,
				"error during closing SQL rows",
				err,
			)
		}
	}()

	var tagIDs []uint32

	for rows.Next() {
		var tagID uint32

		err = rows.Scan(&tagID)
		if err != nil {
			return nil, err
		}

		tagIDs = append(tagIDs, tagID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tagIDs, err
}

func (repo *TicketsRepository) getTicketAttachments(
	ctx context.Context,
	ticketID uint64,
	connection *sql.Conn,
) ([]entities.Attachment, error) {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	stmt, params, err := sq.
		Select(selectAllColumns).
		From(ticketsAttachmentsTableName).
		Where(sq.Eq{ticketIDColumnName: ticketID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := connection.QueryContext(
		ctx,
		stmt,
		params...,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = rows.Close(); err != nil {
			logging.LogErrorContext(
				ctx,
				repo.logger,
				"error during closing SQL rows",
				err,
			)
		}
	}()

	var attachments []entities.Attachment

	for rows.Next() {
		var attachment entities.Attachment
		columns := db.GetEntityColumns(&attachment) // Only pointer to use rows.Scan() successfully

		err = rows.Scan(columns...)
		if err != nil {
			return nil, err
		}

		attachments = append(attachments, attachment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}
