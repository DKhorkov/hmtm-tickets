package repositories

import (
	"context"
	"log/slog"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/tracing"
)

func NewCommonRespondsRepository(
	dbConnector db.Connector,
	logger *slog.Logger,
	traceProvider tracing.TraceProvider,
	spanConfig tracing.SpanConfig,
) *CommonRespondsRepository {
	return &CommonRespondsRepository{
		dbConnector:   dbConnector,
		logger:        logger,
		traceProvider: traceProvider,
		spanConfig:    spanConfig,
	}
}

type CommonRespondsRepository struct {
	dbConnector   db.Connector
	logger        *slog.Logger
	traceProvider tracing.TraceProvider
	spanConfig    tracing.SpanConfig
}

func (repo *CommonRespondsRepository) RespondToTicket(
	ctx context.Context,
	respondData entities.RespondToTicketDTO,
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

	var respondID uint64
	err = connection.QueryRowContext(
		ctx,
		`
			INSERT INTO responds (ticket_id, master_id) 
			VALUES ($1, $2)
			RETURNING responds.id
		`,
		respondData.TicketID,
		respondData.MasterID,
	).Scan(&respondID)

	if err != nil {
		return 0, err
	}

	return respondID, nil
}

func (repo *CommonRespondsRepository) GetRespondByID(ctx context.Context, id uint64) (*entities.Respond, error) {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	connection, err := repo.dbConnector.Connection(ctx)
	if err != nil {
		return nil, err
	}

	defer db.CloseConnectionContext(ctx, connection, repo.logger)

	respond := &entities.Respond{}
	columns := db.GetEntityColumns(respond)
	err = connection.QueryRowContext(
		ctx,
		`
			SELECT * 
			FROM responds AS r
			WHERE r.id = $1
		`,
		id,
	).Scan(columns...)

	if err != nil {
		return nil, err
	}

	return respond, nil
}

func (repo *CommonRespondsRepository) GetTicketResponds(
	ctx context.Context,
	ticketID uint64,
) ([]entities.Respond, error) {
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
			FROM responds AS r
			WHERE r.ticket_id = $1
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

	var responds []entities.Respond
	for rows.Next() {
		respond := entities.Respond{}
		columns := db.GetEntityColumns(&respond) // Only pointer to use rows.Scan() successfully
		err = rows.Scan(columns...)
		if err != nil {
			return nil, err
		}

		responds = append(responds, respond)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return responds, nil
}

func (repo *CommonRespondsRepository) GetMasterResponds(
	ctx context.Context,
	masterID uint64,
) ([]entities.Respond, error) {
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
			FROM responds AS r
			WHERE r.master_id = $1
		`,
		masterID,
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

	var responds []entities.Respond
	for rows.Next() {
		respond := entities.Respond{}
		columns := db.GetEntityColumns(&respond) // Only pointer to use rows.Scan() successfully
		err = rows.Scan(columns...)
		if err != nil {
			return nil, err
		}

		responds = append(responds, respond)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return responds, nil
}
