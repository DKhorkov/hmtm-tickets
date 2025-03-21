package config

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/loadenv"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/tracing"
)

func New() Config {
	return Config{
		Environment: loadenv.GetEnv("ENVIRONMENT", "local"),
		Version:     loadenv.GetEnv("VERSION", "latest"),
		HTTP: HTTPConfig{
			Host: loadenv.GetEnv("HOST", "0.0.0.0"),
			Port: loadenv.GetEnvAsInt("PORT", 8050),
		},
		Database: db.Config{
			Host:         loadenv.GetEnv("POSTGRES_HOST", "0.0.0.0"),
			Port:         loadenv.GetEnvAsInt("POSTGRES_PORT", 5432),
			User:         loadenv.GetEnv("POSTGRES_USER", "postgres"),
			Password:     loadenv.GetEnv("POSTGRES_PASSWORD", "postgres"),
			DatabaseName: loadenv.GetEnv("POSTGRES_DB", "postgres"),
			SSLMode:      loadenv.GetEnv("POSTGRES_SSL_MODE", "disable"),
			Driver:       loadenv.GetEnv("POSTGRES_DRIVER", "postgres"),
			Pool: db.PoolConfig{
				MaxIdleConnections: loadenv.GetEnvAsInt("MAX_IDLE_CONNECTIONS", 1),
				MaxOpenConnections: loadenv.GetEnvAsInt("MAX_OPEN_CONNECTIONS", 1),
				MaxConnectionLifetime: time.Second * time.Duration(
					loadenv.GetEnvAsInt("MAX_CONNECTION_LIFETIME", 20),
				),
				MaxConnectionIdleTime: time.Second * time.Duration(
					loadenv.GetEnvAsInt("MAX_CONNECTION_IDLE_TIME", 10),
				),
			},
		},
		Logging: logging.Config{
			Level:       logging.Levels.DEBUG,
			LogFilePath: fmt.Sprintf("logs/%s.log", time.Now().UTC().Format("02-01-2006")),
		},
		Clients: ClientsConfig{
			Toys: ClientConfig{
				Host:         loadenv.GetEnv("TOYS_CLIENT_HOST", "0.0.0.0"),
				Port:         loadenv.GetEnvAsInt("TOYS_CLIENT_PORT", 8060),
				RetriesCount: loadenv.GetEnvAsInt("TOYS_RETRIES_COUNT", 3),
				RetryTimeout: time.Second * time.Duration(
					loadenv.GetEnvAsInt("TOYS_RETRIES_TIMEOUT", 1),
				),
			},
		},
		NATS: NATSConfig{
			ClientURL: fmt.Sprintf(
				"nats://%s:%d",
				loadenv.GetEnv("NATS_HOST", "0.0.0.0"),
				loadenv.GetEnvAsInt("NATS_CLIENT_PORT", 4222),
			),
			Subjects: NATSSubjects{
				UpdateTicket: loadenv.GetEnv("NATS_UPDATE_TICKET_SUBJECT", "update-ticket"),
				DeleteTicket: loadenv.GetEnv("NATS_DELETE_TICKET_SUBJECT", "delete-ticket"),
			},
			Publisher: NATSPublisher{
				Name: loadenv.GetEnv("NATS_PUBLISHER_NAME", "hmtm-tickets-publisher"),
			},
		},
		Tracing: TracingConfig{
			Server: tracing.Config{
				ServiceName:    loadenv.GetEnv("TRACING_SERVICE_NAME", "hmtm-tickets"),
				ServiceVersion: loadenv.GetEnv("VERSION", "latest"),
				JaegerURL: fmt.Sprintf(
					"http://%s:%d/api/traces",
					loadenv.GetEnv("TRACING_JAEGER_HOST", "0.0.0.0"),
					loadenv.GetEnvAsInt("TRACING_API_TRACES_PORT", 14268),
				),
			},
			Spans: SpansConfig{
				Root: tracing.SpanConfig{
					Opts: []trace.SpanStartOption{
						trace.WithAttributes(
							attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
						),
					},
					Events: tracing.SpanEventsConfig{
						Start: tracing.SpanEventConfig{
							Name: "Calling handler",
							Opts: []trace.EventOption{
								trace.WithAttributes(
									attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
								),
							},
						},
						End: tracing.SpanEventConfig{
							Name: "Received response from handler",
							Opts: []trace.EventOption{
								trace.WithAttributes(
									attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
								),
							},
						},
					},
				},
				Repositories: SpanRepositories{
					Responds: tracing.SpanConfig{
						Opts: []trace.SpanStartOption{
							trace.WithAttributes(
								attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
							),
						},
						Events: tracing.SpanEventsConfig{
							Start: tracing.SpanEventConfig{
								Name: "Calling database",
								Opts: []trace.EventOption{
									trace.WithAttributes(
										attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
									),
								},
							},
							End: tracing.SpanEventConfig{
								Name: "Received response from database",
								Opts: []trace.EventOption{
									trace.WithAttributes(
										attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
									),
								},
							},
						},
					},
					Tickets: tracing.SpanConfig{
						Opts: []trace.SpanStartOption{
							trace.WithAttributes(
								attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
							),
						},
						Events: tracing.SpanEventsConfig{
							Start: tracing.SpanEventConfig{
								Name: "Calling database",
								Opts: []trace.EventOption{
									trace.WithAttributes(
										attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
									),
								},
							},
							End: tracing.SpanEventConfig{
								Name: "Received response from database",
								Opts: []trace.EventOption{
									trace.WithAttributes(
										attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
									),
								},
							},
						},
					},
				},
				Clients: SpanClients{
					Toys: tracing.SpanConfig{
						Opts: []trace.SpanStartOption{
							trace.WithAttributes(
								attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
							),
						},
						Events: tracing.SpanEventsConfig{
							Start: tracing.SpanEventConfig{
								Name: "Calling gRPC Toys client",
								Opts: []trace.EventOption{
									trace.WithAttributes(
										attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
									),
								},
							},
							End: tracing.SpanEventConfig{
								Name: "Received response from gRPC Toys client",
								Opts: []trace.EventOption{
									trace.WithAttributes(
										attribute.String("Environment", loadenv.GetEnv("ENVIRONMENT", "local")),
									),
								},
							},
						},
					},
				},
			},
		},
	}
}

type ClientConfig struct {
	Host         string
	Port         int
	RetryTimeout time.Duration
	RetriesCount int
}

type ClientsConfig struct {
	Toys ClientConfig
}

type HTTPConfig struct {
	Host string
	Port int
}

type TracingConfig struct {
	Server tracing.Config
	Spans  SpansConfig
}

type SpansConfig struct {
	Root         tracing.SpanConfig
	Repositories SpanRepositories
	Clients      SpanClients
}

type SpanRepositories struct {
	Responds tracing.SpanConfig
	Tickets  tracing.SpanConfig
}

type SpanClients struct {
	Toys tracing.SpanConfig
}

type NATSConfig struct {
	ClientURL string
	Subjects  NATSSubjects
	Publisher NATSPublisher
}

type NATSSubjects struct {
	UpdateTicket string
	DeleteTicket string
}

type NATSPublisher struct {
	Name string
}

type Config struct {
	HTTP        HTTPConfig
	Database    db.Config
	Logging     logging.Config
	Clients     ClientsConfig
	Tracing     TracingConfig
	Environment string
	Version     string
	NATS        NATSConfig
}
