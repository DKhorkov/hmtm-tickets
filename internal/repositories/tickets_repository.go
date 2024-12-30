package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/logging"
)

func NewCommonTicketsRepository(
	dbConnector db.Connector,
	logger *slog.Logger,
) *CommonTicketsRepository {
	return &CommonTicketsRepository{
		dbConnector: dbConnector,
		logger:      logger,
	}
}

type CommonTicketsRepository struct {
	dbConnector db.Connector
	logger      *slog.Logger
}

func (repo *CommonTicketsRepository) CreateTicket(
	ctx context.Context,
	ticketData entities.CreateTicketDTO,
) (uint64, error) {
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

	// Bulk insert of Ticket's Tags.
	ticketTagsInsertPlaceholders := make([]string, 0, len(ticketData.TagsIDs))
	ticketTagsInsertValues := make([]interface{}, 0, len(ticketData.TagsIDs))
	for index, tagID := range ticketData.TagsIDs {
		ticketTagsInsertPlaceholder := fmt.Sprintf("($%d,$%d)",
			index*2+1, // (*2) - where 2 is number of inserted params.
			index*2+2,
		)

		ticketTagsInsertPlaceholders = append(ticketTagsInsertPlaceholders, ticketTagsInsertPlaceholder)
		ticketTagsInsertValues = append(ticketTagsInsertValues, ticketID, tagID)
	}

	_, err = transaction.Exec(
		`
			INSERT INTO ticket_tags_associations (ticket_id, tag_id)
			VALUES 
		`+strings.Join(ticketTagsInsertPlaceholders, ","),
		ticketTagsInsertValues...,
	)

	if err != nil {
		return 0, err
	}

	err = transaction.Commit()
	if err != nil {
		return 0, err
	}

	return ticketID, nil
}

func (repo *CommonTicketsRepository) GetTicketByID(ctx context.Context, id uint64) (*entities.Ticket, error) {
	connection, err := repo.dbConnector.Connection(ctx)
	if err != nil {
		return nil, err
	}

	defer db.CloseConnectionContext(ctx, connection, repo.logger)

	ticket := &entities.Ticket{}
	columns := db.GetEntityColumns(ticket)
	columns = columns[:len(columns)-1] // not to paste TagIDs field ([]uint32) to Scan function.
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

	if err = repo.processTicketTags(ctx, ticket, connection); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (repo *CommonTicketsRepository) processTicketTags(
	ctx context.Context,
	ticket *entities.Ticket,
	connection *sql.Conn,
) error {
	rows, err := connection.QueryContext(
		ctx,
		`
			SELECT tta.tag_id
			FROM ticket_tags_associations AS tta
			WHERE tta.ticket_id = $1
		`,
		ticket.ID,
	)

	if err != nil {
		return err
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
			return err
		}

		tagIDs = append(tagIDs, tagID)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	ticket.TagIDs = tagIDs
	return nil
}

func (repo *CommonTicketsRepository) GetAllTickets(ctx context.Context) ([]entities.Ticket, error) {
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
		columns = columns[:len(columns)-1]      // not to paste TagIDs field ([]uint32) to Scan function.
		err = rows.Scan(columns...)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Reading Tags for each Ticket in new circle due to next error: https://github.com/lib/pq/issues/635
	// Using ticketIndex to avoid range iter semantics error, via using copied variable.
	for ticketIndex := range tickets {
		if err = repo.processTicketTags(ctx, &tickets[ticketIndex], connection); err != nil {
			return nil, err
		}
	}

	return tickets, nil
}

func (repo *CommonTicketsRepository) GetUserTickets(ctx context.Context, userID uint64) ([]entities.Ticket, error) {
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
		columns = columns[:len(columns)-1]      // not to paste TagIDs field ([]uint32) to Scan function.
		err = rows.Scan(columns...)
		if err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Reading Tags for each Ticket in new circle due to next error: https://github.com/lib/pq/issues/635
	// Using ticketIndex to avoid range iter semantics error, via using copied variable.
	for ticketIndex := range tickets {
		if err = repo.processTicketTags(ctx, &tickets[ticketIndex], connection); err != nil {
			return nil, err
		}
	}

	return tickets, nil
}
