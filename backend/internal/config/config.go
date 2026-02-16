package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	Env      string
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OAuth2   OAuth2Config
	GeoIP    GeoIPConfig
	RateLimit RateLimitConfig
	Email    EmailConfig
	Encryption EncryptionConfig
	AI       AIConfig
}

type ServerConfig struct {
	Host           string
	Port           int
	FrontendURL    string
	TrustedProxies []string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type OAuth2Config struct {
	TwitterClientID     string
	TwitterClientSecret string
	TwitterRedirectURL  string
}

type GeoIPConfig struct {
	DBPath      string
	BlockChina  bool
}

type RateLimitConfig struct {
	DefaultPerMinute int
}

type EmailConfig struct {
	Provider      string // "smtp" or "resend"
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPassword  string
	EmailFrom     string
	ResendAPIKey  string
}

type EncryptionConfig struct {
	Key string // 32 bytes for AES-256
}

type AIConfig struct {
	DefaultMemoryModel      string
	MemoryExtractionEnabled bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	jwtExpirationHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		jwtExpirationHours = 24
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		dbPort = 5432
	}

	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		redisPort = 6379
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		smtpPort = 587
	}

	rateLimitDefault, err := strconv.Atoi(getEnv("RATE_LIMIT_DEFAULT_PER_MINUTE", "20"))
	if err != nil {
		rateLimitDefault = 20
	}

	blockChina, err := strconv.ParseBool(getEnv("GEOIP_BLOCK_CHINA", "true"))
	if err != nil {
		blockChina = true
	}

	memoryExtractionEnabled, err := strconv.ParseBool(getEnv("MEMORY_EXTRACTION_ENABLED", "true"))
	if err != nil {
		memoryExtractionEnabled = true
	}

	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		serverPort = 8080
	}

	cfg := &Config{
		Env: getEnv("ENV", "development"),
		Server: ServerConfig{
			Host:           getEnv("SERVER_HOST", "0.0.0.0"),
			Port:           serverPort,
			FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:3000"),
			TrustedProxies: parseTrustedProxies(getEnv("TRUSTED_PROXIES", "")),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "ai_chat_user"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "ai_chat_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     redisPort,
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "change_this_secret_key"),
			Expiration: time.Duration(jwtExpirationHours) * time.Hour,
		},
		OAuth2: OAuth2Config{
			TwitterClientID:     getEnv("OAUTH2_TWITTER_CLIENT_ID", ""),
			TwitterClientSecret: getEnv("OAUTH2_TWITTER_CLIENT_SECRET", ""),
			TwitterRedirectURL:  getEnv("OAUTH2_TWITTER_REDIRECT_URL", ""),
		},
		GeoIP: GeoIPConfig{
			DBPath:     getEnv("GEOIP_DB_PATH", "/app/data/GeoLite2-Country.mmdb"),
			BlockChina: blockChina,
		},
		RateLimit: RateLimitConfig{
			DefaultPerMinute: rateLimitDefault,
		},
		Email: EmailConfig{
			Provider:     getEnv("EMAIL_PROVIDER", "smtp"),
			SMTPHost:     getEnv("SMTP_HOST", ""),
			SMTPPort:     smtpPort,
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			EmailFrom:    getEnv("EMAIL_FROM", "noreply@example.com"),
			ResendAPIKey: getEnv("RESEND_API_KEY", ""),
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", ""),
		},
		AI: AIConfig{
			DefaultMemoryModel:      getEnv("DEFAULT_MEMORY_MODEL", "gpt-3.5-turbo"),
			MemoryExtractionEnabled: memoryExtractionEnabled,
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.Secret == "" || c.JWT.Secret == "change_this_secret_key" {
		if c.Env == "production" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
	}

	if c.Encryption.Key == "" {
		if c.Env == "production" {
			return fmt.Errorf("ENCRYPTION_KEY must be set")
		}
	}

	if len(c.Encryption.Key) != 32 && c.Encryption.Key != "" {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes for AES-256")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD must be set")
	}

	return nil
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// GetRedisAddr returns the Redis connection address
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseTrustedProxies parses comma-separated trusted proxy IP addresses
func parseTrustedProxies(value string) []string {
	if value == "" {
		return nil
	}

	var proxies []string
	for _, proxy := range strings.Split(value, ",") {
		proxy = strings.TrimSpace(proxy)
		if proxy != "" {
			proxies = append(proxies, proxy)
		}
	}
	return proxies
}
