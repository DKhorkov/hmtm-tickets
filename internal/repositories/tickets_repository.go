package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/tracing"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
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

	var ticketID uint64
	err = transaction.QueryRow(
		`
			INSERT INTO tickets (user_id, category_id, name, description, price, quantity) 
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING tickets.id
		`,
		ticketData.UserID,
		ticketData.CategoryID,
		ticketData.Name,
		ticketData.Description,
		ticketData.Price,
		ticketData.Quantity,
	).Scan(&ticketID)

	if err != nil {
		return 0, err
	}

	if len(ticketData.TagIDs) > 0 {
		// Bulk insert of Ticket's Tags.
		ticketTagsInsertPlaceholders := make([]string, 0, len(ticketData.TagIDs))
		ticketTagsInsertValues := make([]interface{}, 0, len(ticketData.TagIDs))
		for index, tagID := range ticketData.TagIDs {
			ticketTagsInsertPlaceholder := fmt.Sprintf("($%d,$%d)",
				index*2+1, // (*2) - where 2 is number of inserted params.
				index*2+2,
			)

			ticketTagsInsertPlaceholders = append(ticketTagsInsertPlaceholders, ticketTagsInsertPlaceholder)
			ticketTagsInsertValues = append(ticketTagsInsertValues, ticketID, tagID)
		}

		_, err = transaction.Exec(
			`
				INSERT INTO tickets_tags_associations (ticket_id, tag_id)
				VALUES 
			`+strings.Join(ticketTagsInsertPlaceholders, ","),
			ticketTagsInsertValues...,
		)

		if err != nil {
			return 0, err
		}
	}

	if len(ticketData.Attachments) > 0 {
		// Bulk insert of Ticket's Attachments.
		ticketAttachmentsInsertPlaceholders := make([]string, 0, len(ticketData.Attachments))
		ticketAttachmentsInsertValues := make([]interface{}, 0, len(ticketData.Attachments))
		for index, attachment := range ticketData.Attachments {
			ticketAttachmentsInsertPlaceholder := fmt.Sprintf("($%d,$%d)",
				index*2+1, // (*2) - where 2 is number of inserted params.
				index*2+2,
			)

			ticketAttachmentsInsertPlaceholders = append(
				ticketAttachmentsInsertPlaceholders,
				ticketAttachmentsInsertPlaceholder,
			)

			ticketAttachmentsInsertValues = append(ticketAttachmentsInsertValues, ticketID, attachment)
		}

		_, err = transaction.Exec(
			`
				INSERT INTO tickets_attachments (ticket_id, link)
				VALUES 
			`+strings.Join(ticketAttachmentsInsertPlaceholders, ","),
			ticketAttachmentsInsertValues...,
		)

		if err != nil {
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

	ticket := &entities.Ticket{}
	columns := db.GetEntityColumns(ticket)
	columns = columns[:len(columns)-2] // Not to paste TagIDs and Attachments fields to Scan function.
	err = connection.QueryRowContext(
		ctx,
		`
			SELECT * 
			FROM tickets AS t
			WHERE t.id = $1
		`,
		id,
	).Scan(columns...)

	if err != nil {
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

	rows, err := connection.QueryContext(
		ctx,
		`
			SELECT tta.tag_id
			FROM tickets_tags_associations AS tta
			WHERE tta.ticket_id = $1
		`,
		ticketID,
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

	rows, err := connection.QueryContext(
		ctx,
		`
			SELECT *
			FROM tickets_attachments AS ta
			WHERE ta.ticket_id = $1
		`,
		ticketID,
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

	rows, err := connection.QueryContext(
		ctx,
		`
			SELECT * 
			FROM tickets
		`,
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

	rows, err := connection.QueryContext(
		ctx,
		`
			SELECT * 
			FROM tickets AS t
			WHERE t.user_id = $1
		`,
		userID,
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
