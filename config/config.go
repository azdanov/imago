package config

import "strconv"

type Config struct {
	DB     DBConfig
	SMTP   SMTPConfig
	CSRF   CSRFConfig
	Server ServerConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// Returns the PostgreSQL DSN (Data Source Name) for connecting to the database
func (c *DBConfig) GetDSN() string {
	return "host=" + c.Host +
		" port=" + strconv.Itoa(c.Port) +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.Database +
		" sslmode=" + c.SSLMode
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	SSLMode  bool
}

type CSRFConfig struct {
	Key    string
	Secure bool
}

type ServerConfig struct {
	Host    string
	Port    int
	Env     Environment
	SSLMode bool
}

// Returns the address for the server, including protocol based on SSL mode
func (c *ServerConfig) GetAddr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

func (c *ServerConfig) GetURL() string {
	addr := c.Host + ":" + strconv.Itoa(c.Port)
	if c.SSLMode {
		return "https://" + addr
	}
	return "http://" + addr
}

func NewEnvConfig() Config {
	return Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getIntEnv("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Database: getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     getIntEnv("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", "user"),
			Password: getEnv("SMTP_PASSWORD", "password"),
			SSLMode:  getBoolEnv("SMTP_SSLMODE", false),
		},
		CSRF: CSRFConfig{
			Key:    getEnv("CSRF_KEY", "default-key"),
			Secure: getBoolEnv("CSRF_SECURE", false),
		},
		Server: ServerConfig{
			Host:    getEnv("SERVER_HOST", "localhost"),
			Port:    getIntEnv("SERVER_PORT", 8080),
			Env:     GetEnvironment("SERVER_ENV", Dev),
			SSLMode: getBoolEnv("SERVER_SSLMODE", false),
		},
	}
}
