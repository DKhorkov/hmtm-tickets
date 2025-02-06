package main

import (
	"context"

	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/tracing"

	"github.com/DKhorkov/hmtm-tickets/internal/app"
	toysgrpcclient "github.com/DKhorkov/hmtm-tickets/internal/clients/toys/grpc"
	"github.com/DKhorkov/hmtm-tickets/internal/config"
	grpccontroller "github.com/DKhorkov/hmtm-tickets/internal/controllers/grpc"
	"github.com/DKhorkov/hmtm-tickets/internal/repositories"
	"github.com/DKhorkov/hmtm-tickets/internal/services"
	"github.com/DKhorkov/hmtm-tickets/internal/usecases"
)

func main() {
	settings := config.New()
	logger := logging.GetInstance(
		settings.Logging.Level,
		settings.Logging.LogFilePath,
	)

	dbConnector, err := db.New(
		db.BuildDsn(settings.Database),
		settings.Database.Driver,
		logger,
		db.WithMaxOpenConnections(settings.Database.Pool.MaxOpenConnections),
		db.WithMaxIdleConnections(settings.Database.Pool.MaxIdleConnections),
		db.WithMaxConnectionLifetime(settings.Database.Pool.MaxConnectionLifetime),
		db.WithMaxConnectionIdleTime(settings.Database.Pool.MaxConnectionIdleTime),
	)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err = dbConnector.Close(); err != nil {
			logging.LogError(logger, "Failed to close db connections pool", err)
		}
	}()

	traceProvider, err := tracing.New(settings.Tracing.Server)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = traceProvider.Shutdown(context.Background()); err != nil {
			logging.LogError(logger, "Error shutting down tracer", err)
		}
	}()

	toysClient, err := toysgrpcclient.New(
		settings.Clients.Toys.Host,
		settings.Clients.Toys.Port,
		settings.Clients.Toys.RetriesCount,
		settings.Clients.Toys.RetryTimeout,
		logger,
		traceProvider,
		settings.Tracing.Spans.Clients.Toys,
	)

	if err != nil {
		panic(err)
	}

	ticketsRepository := repositories.NewCommonTicketsRepository(
		dbConnector,
		logger,
		traceProvider,
		settings.Tracing.Spans.Repositories.Tickets,
	)

	toysRepository := repositories.NewGrpcToysRepository(toysClient)
	ticketsService := services.NewCommonTicketsService(
		ticketsRepository,
		toysRepository,
		logger,
	)

	respondsRepository := repositories.NewCommonRespondsRepository(
		dbConnector,
		logger,
		traceProvider,
		settings.Tracing.Spans.Repositories.Responds,
	)

	respondsService := services.NewCommonRespondsService(
		respondsRepository,
		toysRepository,
		logger,
	)

	useCases := usecases.NewCommonUseCases(
		ticketsService,
		respondsService,
	)

	controller := grpccontroller.New(
		settings.HTTP.Host,
		settings.HTTP.Port,
		useCases,
		logger,
		traceProvider,
		settings.Tracing.Spans.Root,
	)

	application := app.New(controller)
	application.Run()
}
