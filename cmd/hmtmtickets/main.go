package main

import (
	"github.com/DKhorkov/hmtm-tickets/internal/app"
	toysgrpcclient "github.com/DKhorkov/hmtm-tickets/internal/clients/toys/grpc"
	"github.com/DKhorkov/hmtm-tickets/internal/config"
	grpccontroller "github.com/DKhorkov/hmtm-tickets/internal/controllers/grpc"
	"github.com/DKhorkov/hmtm-tickets/internal/repositories"
	"github.com/DKhorkov/hmtm-tickets/internal/services"
	"github.com/DKhorkov/hmtm-tickets/internal/usecases"
	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/logging"
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
	)

	if err != nil {
		panic(err)
	}

	defer func() {
		if err = dbConnector.Close(); err != nil {
			logging.LogError(logger, "Failed to close db connections pool", err)
		}
	}()

	toysClient, err := toysgrpcclient.New(
		settings.Clients.Toys.Host,
		settings.Clients.Toys.Port,
		settings.Clients.Toys.RetriesCount,
		settings.Clients.Toys.RetryTimeout,
		logger,
	)

	if err != nil {
		panic(err)
	}

	ticketsRepository := repositories.NewCommonTicketsRepository(dbConnector, logger)
	toysRepository := repositories.NewGrpcToysRepository(toysClient)
	ticketsService := services.NewCommonTicketsService(
		ticketsRepository,
		toysRepository,
		logger,
	)

	respondsRepository := repositories.NewCommonRespondsRepository(dbConnector, logger)
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
	)

	application := app.New(controller)
	application.Run()
}
