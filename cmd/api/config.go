package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

// A config struct to hold all the configuration settings for our application.”
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}

	// Add a new limiter struct containing fields for the requests-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting
	// altogether.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}

	// Configure CORRS
	cors struct {
		trustedOrigins []string
	}
}

func loadConfig() (config, bool) {
	var cfg config
	flag.IntVar(&cfg.port, "port", getIntEnvVar("SERVER_PORT", 4000), "API server port")
	flag.StringVar(
		&cfg.env,
		"env",
		os.Getenv("ENV"),
		"Environment (development|staging|production)",
	)
	// Read the DSN value from the db-dsn command-line flag into the config struct. We
	// default to using our development DSN if no flag is provided.
	// Use the value of the GREENLIGHT_DB_DSN environment variable as the default value
	// for our db-dsn command-line flag.
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	// Read the connection pool settings from command-line flags into the config struct.
	// Note that the default values we're using are the ones we discussed above.
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(
		&cfg.db.maxIdleTime,
		"db-max-idle-time",
		15*time.Minute,
		"PostgreSQL max connection idle time",
	)

	// Create command-line flags to read the settings values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting.
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", getIntEnvVar("SMTP_PORT", 465), "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(
		&cfg.smtp.sender,
		"smtp-sender",
		os.Getenv("SMTP_SENDER"),
		"SMTP sender",
	)

	// Use the flag.Func() function to process the -cors-trusted-origins command line
	// flag. In this we use the strings.Fields() function to split the flag value into a
	// slice based on whitespace characters and assign it to our config struct.
	// Importantly, if the -cors-trusted-origins flag is not present, contains the empty
	// string, or contains only whitespace, then strings.Fields() will return an empty
	// []string slice.
	var corsFlagSet bool

	flag.Func(
		"cors-trusted-origins",
		"Trusted CORS origins (space separated)",
		func(val string) error {
			cfg.cors.trustedOrigins = strings.Fields(val)
			corsFlagSet = true
			return nil
		},
	)

	// Create a new version boolean flag with the default value of false.
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if !corsFlagSet {
		cfg.cors.trustedOrigins = getStringListEnvVar("CORS_TRUSTED_ORIGINS")
	}

	return cfg, *displayVersion
}

// getIntEnvVar reads the environment variable with the given key,
// converts it to an int, and returns it. If the variable does not exist
// or cannot be converted to an int, it returns the default value.
func getIntEnvVar(key string, defaultValue int) int {
	valStr, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	valInt, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}

	return valInt
}

// getStringListEnvVar reads an environment variable by key, splits it into a list of strings,
// using a separator (defaults to ",") and a default value (defaults to an empty list if not provided).
//
// Usage examples:
//
//	getStringListEnvVar("FOO")                                 // uses default "," and []string{}
//	getStringListEnvVar("FOO", ";")                            // uses ";" and []string{}
//	getStringListEnvVar("FOO", ";", []string{"fallback"})      // uses ";" and fallback list
func getStringListEnvVar(key string, args ...interface{}) []string {
	separator := ","
	defaultValue := []string{}

	// Parse optional arguments
	if len(args) >= 1 {
		if s, ok := args[0].(string); ok && s != "" {
			separator = s
		}
	}
	if len(args) >= 2 {
		if def, ok := args[1].([]string); ok {
			defaultValue = def
		}
	}

	valStr, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(valStr) == "" {
		return defaultValue
	}

	parts := strings.Split(valStr, separator)
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}

	return parts
}
