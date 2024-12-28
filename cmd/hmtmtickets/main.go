package main

import (
	"github.com/DKhorkov/hmtm-tickets/internal/app"
	"github.com/DKhorkov/hmtm-tickets/internal/config"
	grpccontroller "github.com/DKhorkov/hmtm-tickets/internal/controllers/grpc"
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

	controller := grpccontroller.New(
		settings.HTTP.Host,
		settings.HTTP.Port,
		nil,
		// useCases,
		logger,
	)

	application := app.New(controller)
	application.Run()
}
