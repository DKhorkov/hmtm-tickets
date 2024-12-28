package config

import (
	"fmt"
	"time"

	"github.com/DKhorkov/libs/db"
	"github.com/DKhorkov/libs/loadenv"
	"github.com/DKhorkov/libs/logging"
	"github.com/DKhorkov/libs/security"
)

func New() Config {
	return Config{
		HTTP: HTTPConfig{
			Host: loadenv.GetEnv("HOST", "0.0.0.0"),
			Port: loadenv.GetEnvAsInt("PORT", 8050),
		},
		Security: security.Config{
			JWT: security.JWTConfig{
				Algorithm: loadenv.GetEnv("JWT_ALGORITHM", "HS256"),
				SecretKey: loadenv.GetEnv("JWT_SECRET", "defaultSecret"),
			},
		},
		Database: db.Config{
			Host:         loadenv.GetEnv("POSTGRES_HOST", "0.0.0.0"),
			Port:         loadenv.GetEnvAsInt("POSTGRES_PORT", 5432),
			User:         loadenv.GetEnv("POSTGRES_USER", "postgres"),
			Password:     loadenv.GetEnv("POSTGRES_PASSWORD", "postgres"),
			DatabaseName: loadenv.GetEnv("POSTGRES_DB", "postgres"),
			SSLMode:      loadenv.GetEnv("POSTGRES_SSL_MODE", "disable"),
			Driver:       loadenv.GetEnv("POSTGRES_DRIVER", "postgres"),
		},
		Logging: logging.Config{
			Level:       logging.Levels.DEBUG,
			LogFilePath: fmt.Sprintf("logs/%s.log", time.Now().Format("02-01-2006")),
		},
	}
}

type HTTPConfig struct {
	Host string
	Port int
}

type Config struct {
	HTTP     HTTPConfig
	Security security.Config
	Database db.Config
	Logging  logging.Config
}
