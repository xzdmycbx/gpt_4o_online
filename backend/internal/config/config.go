package config

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ai-chat/backend/internal/pkg/crypto"
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
		// OAuth2、Email、AI 配置将从数据库加载，这里使用环境变量作为初始值
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

	if c.Encryption.Key != "" && !isValidEncryptionKey(c.Encryption.Key) {
		return fmt.Errorf("ENCRYPTION_KEY must be 32 chars (raw) or 64 chars (hex) for AES-256")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD must be set")
	}

	return nil
}

func isValidEncryptionKey(key string) bool {
	if len(key) == 32 {
		return true
	}

	if len(key) == 64 {
		_, err := hex.DecodeString(key)
		return err == nil
	}

	return false
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

// LoadDynamicConfig loads dynamic configuration from database
func (c *Config) LoadDynamicConfig(db *sql.DB) error {
	ctx := context.Background()

	// 查询所有系统设置
	query := `SELECT setting_key, setting_value FROM system_settings`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query system settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		settings[key] = value
	}

	// 加载 OAuth2 配置
	if val, ok := settings["oauth2_twitter_client_id"]; ok && val != "" {
		c.OAuth2.TwitterClientID = val
	}
	if val, ok := settings["oauth2_twitter_client_secret"]; ok && val != "" {
		// 解密
		if decrypted, err := crypto.Decrypt(val, c.Encryption.Key); err == nil {
			c.OAuth2.TwitterClientSecret = decrypted
		}
	}
	if val, ok := settings["oauth2_twitter_redirect_url"]; ok && val != "" {
		c.OAuth2.TwitterRedirectURL = val
	}

	// 加载邮件配置
	if val, ok := settings["email_provider"]; ok && val != "" {
		c.Email.Provider = val
	}
	if val, ok := settings["email_smtp_host"]; ok && val != "" {
		c.Email.SMTPHost = val
	}
	if val, ok := settings["email_smtp_port"]; ok && val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			c.Email.SMTPPort = port
		}
	}
	if val, ok := settings["email_smtp_user"]; ok && val != "" {
		c.Email.SMTPUser = val
	}
	if val, ok := settings["email_smtp_password"]; ok && val != "" {
		// 解密
		if decrypted, err := crypto.Decrypt(val, c.Encryption.Key); err == nil {
			c.Email.SMTPPassword = decrypted
		}
	}
	if val, ok := settings["email_from"]; ok && val != "" {
		c.Email.EmailFrom = val
	}
	if val, ok := settings["email_resend_api_key"]; ok && val != "" {
		// 解密
		if decrypted, err := crypto.Decrypt(val, c.Encryption.Key); err == nil {
			c.Email.ResendAPIKey = decrypted
		}
	}

	// 加载 AI 配置
	if val, ok := settings["ai_default_memory_model"]; ok && val != "" {
		c.AI.DefaultMemoryModel = val
	}
	if val, ok := settings["ai_memory_extraction_enabled"]; ok {
		c.AI.MemoryExtractionEnabled = (val == "true")
	}

	// 加载速率限制配置
	if val, ok := settings["rate_limit_default_per_minute"]; ok && val != "" {
		if limit, err := strconv.Atoi(val); err == nil {
			c.RateLimit.DefaultPerMinute = limit
		}
	}

	return nil
}
