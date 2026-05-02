package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	MinIO    MinIOConfig
	SMTP     SMTPConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port string
	URL  string
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	PoolMin  int
	PoolMax  int
}

func (c DatabaseConfig) DSN() string {
	if c.URL != "" {
		return c.URL
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}

type RedisConfig struct {
	URL      string
	Host     string
	Port     string
	Password string
	DB       int
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type JWTConfig struct {
	AccessSecret     string
	RefreshSecret    string
	AccessExpiresIn  time.Duration
	RefreshExpiresIn time.Duration
}

type MinIOConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	UseSSL     bool
	PublicHost string
}

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	_ = viper.ReadInConfig() // Ignore error - env vars will be used instead

	setDefaults()

	cfg := &Config{
		App: AppConfig{
			Name: viper.GetString("APP_NAME"),
			Env:  viper.GetString("APP_ENV"),
			Port: viper.GetString("APP_PORT"),
			URL:  viper.GetString("APP_URL"),
		},
		Database: DatabaseConfig{
			URL:      viper.GetString("DATABASE_URL"),
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
			PoolMin:  viper.GetInt("DB_POOL_MIN"),
			PoolMax:  viper.GetInt("DB_POOL_MAX"),
		},
		Redis: RedisConfig{
			URL:      viper.GetString("REDIS_URL"),
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			AccessSecret:     viper.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret:    viper.GetString("JWT_REFRESH_SECRET"),
			AccessExpiresIn:  viper.GetDuration("JWT_ACCESS_EXPIRES_IN"),
			RefreshExpiresIn: viper.GetDuration("JWT_REFRESH_EXPIRES_IN"),
		},
		MinIO: MinIOConfig{
			Endpoint:   viper.GetString("MINIO_ENDPOINT"),
			AccessKey:  viper.GetString("MINIO_ACCESS_KEY"),
			SecretKey:  viper.GetString("MINIO_SECRET_KEY"),
			Bucket:     viper.GetString("MINIO_BUCKET"),
			UseSSL:     viper.GetBool("MINIO_USE_SSL"),
			PublicHost: viper.GetString("MINIO_PUBLIC_HOST"),
		},
		SMTP: SMTPConfig{
			Host: viper.GetString("SMTP_HOST"),
			Port: viper.GetInt("SMTP_PORT"),
			User: viper.GetString("SMTP_USER"),
			Pass: viper.GetString("SMTP_PASS"),
			From: viper.GetString("SMTP_FROM"),
		},
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setDefaults() {
	viper.SetDefault("APP_NAME", "trello-agent")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_URL", "http://localhost:8080")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_POOL_MIN", 2)
	viper.SetDefault("DB_POOL_MAX", 10)

	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("JWT_ACCESS_EXPIRES_IN", "15m")
	viper.SetDefault("JWT_REFRESH_EXPIRES_IN", "168h")

	viper.SetDefault("MINIO_USE_SSL", false)

	viper.SetDefault("SMTP_PORT", 1025)
}

func validate(cfg *Config) error {
	if cfg.Database.URL == "" {
		if cfg.Database.User == "" {
			return fmt.Errorf("DATABASE_URL or DB_USER is required")
		}
		if cfg.Database.Password == "" {
			return fmt.Errorf("DATABASE_URL or DB_PASSWORD is required")
		}
		if cfg.Database.Name == "" {
			return fmt.Errorf("DATABASE_URL or DB_NAME is required")
		}
	}
	if cfg.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	if cfg.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	return nil
}
