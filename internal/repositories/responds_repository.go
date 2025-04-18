package repositories

import (
	"context"

	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/tracing"

	sq "github.com/Masterminds/squirrel"

	"github.com/DKhorkov/hmtm-tickets/internal/entities"
)

const (
	respondsTableName        = "responds"
	masterIDColumnName       = "master_id"
	respondPriceColumnName   = "price"
	respondCommentColumnName = "comment"
)

type RespondsRepository struct {
	dbConnector   db.Connector
	logger        logging.Logger
	traceProvider tracing.Provider
	spanConfig    tracing.SpanConfig
}

func NewRespondsRepository(
	dbConnector db.Connector,
	logger logging.Logger,
	traceProvider tracing.Provider,
	spanConfig tracing.SpanConfig,
) *RespondsRepository {
	return &RespondsRepository{
		dbConnector:   dbConnector,
		logger:        logger,
		traceProvider: traceProvider,
		spanConfig:    spanConfig,
	}
}

func (repo *RespondsRepository) RespondToTicket(
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

	stmt, params, err := sq.
		Insert(respondsTableName).
		Columns(
			ticketIDColumnName,
			masterIDColumnName,
			respondPriceColumnName,
			respondCommentColumnName,
		).
		Values(
			respondData.TicketID,
			respondData.MasterID,
			respondData.Price,
			respondData.Comment,
		).
		Suffix(returningIDSuffix).
		PlaceholderFormat(sq.Dollar). // pq postgres driver works only with $ placeholders
		ToSql()
	if err != nil {
		return 0, err
	}

	var respondID uint64
	if err = connection.QueryRowContext(ctx, stmt, params...).Scan(&respondID); err != nil {
		return 0, err
	}

	return respondID, nil
}

func (repo *RespondsRepository) GetRespondByID(
	ctx context.Context,
	id uint64,
) (*entities.Respond, error) {
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
		From(respondsTableName).
		Where(sq.Eq{idColumnName: id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	respond := &entities.Respond{}

	columns := db.GetEntityColumns(respond)
	if err = connection.QueryRowContext(ctx, stmt, params...).Scan(columns...); err != nil {
		return nil, err
	}

	return respond, nil
}

func (repo *RespondsRepository) GetTicketResponds(
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

	stmt, params, err := sq.
		Select(selectAllColumns).
		From(respondsTableName).
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

func (repo *RespondsRepository) GetMasterResponds(
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

	stmt, params, err := sq.
		Select(selectAllColumns).
		From(respondsTableName).
		Where(sq.Eq{masterIDColumnName: masterID}).
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

func (repo *RespondsRepository) UpdateRespond(
	ctx context.Context,
	respondData entities.UpdateRespondDTO,
) error {
	ctx, span := repo.traceProvider.Span(ctx, tracing.CallerName(tracing.DefaultSkipLevel))
	defer span.End()

	span.AddEvent(repo.spanConfig.Events.Start.Name, repo.spanConfig.Events.Start.Opts...)
	defer span.AddEvent(repo.spanConfig.Events.End.Name, repo.spanConfig.Events.End.Opts...)

	connection, err := repo.dbConnector.Connection(ctx)
	if err != nil {
		return err
	}

	defer db.CloseConnectionContext(ctx, connection, repo.logger)

	builder := sq.
		Update(respondsTableName).
		Where(sq.Eq{idColumnName: respondData.ID}).
		Set(respondCommentColumnName, respondData.Comment).
		// Update every time, because field is nullable
		PlaceholderFormat(sq.Dollar)
	// pq postgres driver works only with $ placeholders

	// Price is not nullable, so update only when field is not nil:
	if respondData.Price != nil {
		builder = builder.Set(respondPriceColumnName, respondData.Price)
	}

	stmt, params, err := builder.ToSql()
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

func (repo *RespondsRepository) DeleteRespond(ctx context.Context, id uint64) error {
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
		Delete(respondsTableName).
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
