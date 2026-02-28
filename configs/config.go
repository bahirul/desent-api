package configs

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Logging  LoggingConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Address           string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type LoggingConfig struct {
	LogsDir            string
	RetentionDays      int
	ConsoleEnabled     bool
	FileEnabled        bool
	HTTPLogFilePrefix  string
	ErrorLogFilePrefix string
}

type DatabaseConfig struct {
	Driver string
	DSN    string
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Address:           Getenv("HTTP_ADDR", ":8080"),
			ReadTimeout:       time.Duration(GetenvInt("HTTP_READ_TIMEOUT_SECONDS", 10)) * time.Second,
			ReadHeaderTimeout: time.Duration(GetenvInt("HTTP_READ_HEADER_TIMEOUT_SECONDS", 5)) * time.Second,
			WriteTimeout:      time.Duration(GetenvInt("HTTP_WRITE_TIMEOUT_SECONDS", 15)) * time.Second,
			IdleTimeout:       time.Duration(GetenvInt("HTTP_IDLE_TIMEOUT_SECONDS", 60)) * time.Second,
		},
		Logging: LoggingConfig{
			LogsDir:            Getenv("LOGS_DIR", "logs"),
			RetentionDays:      GetenvInt("LOG_RETENTION_DAYS", 7),
			ConsoleEnabled:     GetenvBool("HTTP_LOG_CONSOLE_ENABLED", true),
			FileEnabled:        GetenvBool("HTTP_LOG_FILE_ENABLED", true),
			HTTPLogFilePrefix:  "http",
			ErrorLogFilePrefix: "error",
		},
		Database: DatabaseConfig{
			Driver: Getenv("DB_DRIVER", "sqlite"),
			DSN:    Getenv("DB_DSN", "file:data/books.db"),
		},
	}
}

func Getenv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return defaultValue
}

func GetenvInt(key string, defaultValue int) int {
	value := Getenv(key, "")
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func GetenvBool(key string, defaultValue bool) bool {
	value := Getenv(key, "")
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}
