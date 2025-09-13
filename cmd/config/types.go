package config

var Config = &config{}

type config struct {
	R2       R2
	Postgres Postgres
}

type R2 struct {
	AccessKey       string `env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`

	AccountID string `env:"ACCOUNT_ID"`
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"postgres"`
	DBName   string `env:"POSTGRES_DB" envDefault:"task"`
	SSLMode  string `env:"POSTGRES_SSL_MODE" envDefault:"disable"`
}
