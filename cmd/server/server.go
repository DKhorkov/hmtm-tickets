package main

import (
	"context"

	"github.com/nats-io/nats.go"

	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/logging"
	customnats "github.com/DKhorkov/libs/nats"
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
	logger := logging.New(
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

	natsPublisher, err := customnats.NewPublisher(
		settings.NATS.ClientURL,
		nats.Name(settings.NATS.Publisher.Name),
	)

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

	toysRepository := repositories.NewToysRepository(toysClient)
	toysService := services.NewToysService(
		toysRepository,
		logger,
	)

	ticketsRepository := repositories.NewTicketsRepository(
		dbConnector,
		logger,
		traceProvider,
		settings.Tracing.Spans.Repositories.Tickets,
	)

	ticketsService := services.NewTicketsService(
		ticketsRepository,
		logger,
	)

	respondsRepository := repositories.NewRespondsRepository(
		dbConnector,
		logger,
		traceProvider,
		settings.Tracing.Spans.Repositories.Responds,
	)

	respondsService := services.NewRespondsService(
		respondsRepository,
		logger,
	)

	useCases := usecases.New(
		ticketsService,
		respondsService,
		toysService,
		natsPublisher,
		settings.NATS,
		logger,
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
