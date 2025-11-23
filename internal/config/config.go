package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	SuperTokens SuperTokensConfig
	Redis       RedisConfig
	Asynq       AsynqConfig
	Email       EmailConfig
	Invitation  InvitationConfig
	Log         LogConfig
	TenantInit  TenantInitConfig
}

type AppConfig struct {
	Env  string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type SuperTokensConfig struct {
	ConnectionURI string
	APIKey        string
	APIDomain     string
	WebsiteDomain string
	APIBasePath   string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type AsynqConfig struct {
	Concurrency int
	Queues      map[string]int
}

type EmailConfig struct {
	Provider     string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
}

type InvitationConfig struct {
	ExpiryHours int
	BaseURL     string
}

type LogConfig struct {
	Level  string
	Format string
}

type TenantInitConfig struct {
	Services []string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()

	config := &Config{
		App: AppConfig{
			Env:  viper.GetString("app.env"),
			Port: viper.GetString("app.port"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("db.host"),
			Port:     viper.GetString("db.port"),
			User:     viper.GetString("db.user"),
			Password: viper.GetString("db.password"),
			DBName:   viper.GetString("db.name"),
			SSLMode:  viper.GetString("db.sslmode"),
		},
		SuperTokens: SuperTokensConfig{
			ConnectionURI: viper.GetString("supertokens.connection_uri"),
			APIKey:        viper.GetString("supertokens.api_key"),
			APIDomain:     viper.GetString("supertokens.api_domain"),
			WebsiteDomain: viper.GetString("supertokens.website_domain"),
			APIBasePath:   viper.GetString("supertokens.api_base_path"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("redis.host"),
			Port:     viper.GetString("redis.port"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		},
		Asynq: AsynqConfig{
			Concurrency: viper.GetInt("asynq.concurrency"),
			Queues:      parseQueues(viper.GetString("asynq.queues")),
		},
		Email: EmailConfig{
			Provider:     viper.GetString("email.provider"),
			SMTPHost:     viper.GetString("email.smtp_host"),
			SMTPPort:     viper.GetString("email.smtp_port"),
			SMTPUser:     viper.GetString("email.smtp_user"),
			SMTPPassword: viper.GetString("email.smtp_password"),
			FromAddress:  viper.GetString("email.from_address"),
		},
		Invitation: InvitationConfig{
			ExpiryHours: viper.GetInt("invitation.expiry_hours"),
			BaseURL:     viper.GetString("invitation.base_url"),
		},
		Log: LogConfig{
			Level:  viper.GetString("log.level"),
			Format: viper.GetString("log.format"),
		},
		TenantInit: TenantInitConfig{
			Services: parseServices(viper.GetString("tenant_init.services")),
		},
	}

	return config, nil
}

func setDefaults() {
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.port", "8080")

	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", "5432")
	viper.SetDefault("db.user", "utmuser")
	viper.SetDefault("db.password", "utmpassword")
	viper.SetDefault("db.name", "utm_backend")
	viper.SetDefault("db.sslmode", "disable")

	viper.SetDefault("supertokens.connection_uri", "http://localhost:3567")
	viper.SetDefault("supertokens.api_key", "")
	viper.SetDefault("supertokens.api_domain", "http://localhost:8080")
	viper.SetDefault("supertokens.website_domain", "http://localhost:3000")
	viper.SetDefault("supertokens.api_base_path", "/api/auth")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("asynq.concurrency", 10)
	viper.SetDefault("asynq.queues", "critical:6,default:3,low:1")

	viper.SetDefault("email.provider", "smtp")
	viper.SetDefault("email.smtp_host", "localhost")
	viper.SetDefault("email.smtp_port", "1025")
	viper.SetDefault("email.smtp_user", "")
	viper.SetDefault("email.smtp_password", "")
	viper.SetDefault("email.from_address", "noreply@utm.local")

	viper.SetDefault("invitation.expiry_hours", 72)
	viper.SetDefault("invitation.base_url", "http://localhost:3000/accept-invite")

	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.format", "json")

	viper.SetDefault("tenant_init.services", "")

	// Bind environment variables
	viper.BindEnv("app.env", "APP_ENV")
	viper.BindEnv("app.port", "APP_PORT")
	viper.BindEnv("db.host", "DB_HOST")
	viper.BindEnv("db.port", "DB_PORT")
	viper.BindEnv("db.user", "DB_USER")
	viper.BindEnv("db.password", "DB_PASSWORD")
	viper.BindEnv("db.name", "DB_NAME")
	viper.BindEnv("db.sslmode", "DB_SSL_MODE")
	viper.BindEnv("supertokens.connection_uri", "SUPERTOKENS_CONNECTION_URI")
	viper.BindEnv("supertokens.api_key", "SUPERTOKENS_API_KEY")
	viper.BindEnv("supertokens.api_domain", "API_DOMAIN")
	viper.BindEnv("supertokens.website_domain", "WEBSITE_DOMAIN")
	viper.BindEnv("supertokens.api_base_path", "API_BASE_PATH")
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("asynq.concurrency", "ASYNQ_CONCURRENCY")
	viper.BindEnv("asynq.queues", "ASYNQ_QUEUES")
	viper.BindEnv("email.provider", "EMAIL_PROVIDER")
	viper.BindEnv("email.smtp_host", "SMTP_HOST")
	viper.BindEnv("email.smtp_port", "SMTP_PORT")
	viper.BindEnv("email.smtp_user", "SMTP_USER")
	viper.BindEnv("email.smtp_password", "SMTP_PASSWORD")
	viper.BindEnv("email.from_address", "EMAIL_FROM")
	viper.BindEnv("invitation.expiry_hours", "INVITATION_EXPIRY_HOURS")
	viper.BindEnv("invitation.base_url", "INVITATION_BASE_URL")
	viper.BindEnv("log.level", "LOG_LEVEL")
	viper.BindEnv("log.format", "LOG_FORMAT")
	viper.BindEnv("tenant_init.services", "TENANT_INIT_SERVICES")
}

func parseQueues(queueStr string) map[string]int {
	queues := make(map[string]int)
	if queueStr == "" {
		queues["default"] = 3
		return queues
	}

	pairs := strings.Split(queueStr, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) == 2 {
			var priority int
			fmt.Sscanf(parts[1], "%d", &priority)
			queues[parts[0]] = priority
		}
	}
	return queues
}

func parseServices(servicesStr string) []string {
	if servicesStr == "" {
		return []string{}
	}
	services := strings.Split(servicesStr, ",")
	for i, s := range services {
		services[i] = strings.TrimSpace(s)
	}
	return services
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func (c *Config) GetInvitationExpiry() time.Duration {
	return time.Duration(c.Invitation.ExpiryHours) * time.Hour
}

func IsDevelopment() bool {
	return os.Getenv("APP_ENV") == "development"
}

func IsProduction() bool {
	return os.Getenv("APP_ENV") == "production"
}
