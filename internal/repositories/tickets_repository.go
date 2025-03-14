package repositories

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/tracing"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
)

const (
	selectAllColumns                   = "*"
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
)

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

type TicketsRepository struct {
	dbConnector   db.Connector
	logger        logging.Logger
	traceProvider tracing.Provider
	spanConfig    tracing.SpanConfig
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
		builder := sq.Insert(ticketsAndTagsAssociationTableName).Columns(ticketIDColumnName, tagIDColumnName)
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
		builder := sq.Insert(ticketsAttachmentsTableName).Columns(ticketIDColumnName, attachmentLinkColumnName)
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

func (repo *TicketsRepository) GetTicketByID(ctx context.Context, id uint64) (*entities.Ticket, error) {
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

func (repo *TicketsRepository) GetAllTickets(ctx context.Context) ([]entities.Ticket, error) {
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

func (repo *TicketsRepository) GetUserTickets(ctx context.Context, userID uint64) ([]entities.Ticket, error) {
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
		Where(sq.Eq{userIDColumnName: userID}).
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
