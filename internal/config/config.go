package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env                   string
	PublicBaseURL         string
	InternalToken         string
	InternalAPIBaseURL    string
	APIAddr               string
	APIDBPath             string
	WorkerDBPath          string
	NATSURL               string
	NATSStream            string
	NATSJobSubject        string
	PayloadInlineMaxBytes int64
	DefaultTTLSeconds     int
	AllowedTTLSeconds     []int
}

func Load() (Config, error) {
	cfg := Config{
		Env:                   getenv("FLICK_ENV", "development"),
		PublicBaseURL:         getenv("FLICK_PUBLIC_BASE_URL", "http://localhost:5173"),
		InternalToken:         getenv("FLICK_INTERNAL_TOKEN", ""),
		InternalAPIBaseURL:    getenv("FLICK_INTERNAL_API_BASE_URL", "http://localhost:8080"),
		APIAddr:               getenv("FLICK_API_ADDR", ":8080"),
		APIDBPath:             getenv("FLICK_API_DB_PATH", "./var/api.db"),
		WorkerDBPath:          getenv("FLICK_WORKER_DB_PATH", "./var/worker.db"),
		NATSURL:               getenv("FLICK_NATS_URL", "nats://127.0.0.1:4222"),
		NATSStream:            getenv("FLICK_NATS_STREAM", "FLICK_JOBS"),
		NATSJobSubject:        getenv("FLICK_NATS_JOB_SUBJECT", "flick.jobs"),
		PayloadInlineMaxBytes: 1048576,
		DefaultTTLSeconds:     3600,
		AllowedTTLSeconds:     []int{600, 3600, 86400},
	}

	var err error
	if raw := os.Getenv("FLICK_PAYLOAD_INLINE_MAX_BYTES"); raw != "" {
		cfg.PayloadInlineMaxBytes, err = parsePositiveInt64("FLICK_PAYLOAD_INLINE_MAX_BYTES", raw)
		if err != nil {
			return Config{}, err
		}
	}
	if raw := os.Getenv("FLICK_DEFAULT_TTL_SECONDS"); raw != "" {
		cfg.DefaultTTLSeconds, err = parsePositiveInt("FLICK_DEFAULT_TTL_SECONDS", raw)
		if err != nil {
			return Config{}, err
		}
	}
	if raw := os.Getenv("FLICK_ALLOWED_TTL_SECONDS"); raw != "" {
		cfg.AllowedTTLSeconds, err = parseAllowedTTLs(raw)
		if err != nil {
			return Config{}, err
		}
	}

	if !containsInt(cfg.AllowedTTLSeconds, cfg.DefaultTTLSeconds) {
		return Config{}, fmt.Errorf("FLICK_DEFAULT_TTL_SECONDS must be present in FLICK_ALLOWED_TTL_SECONDS")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parsePositiveInt(name, raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}

func parsePositiveInt64(name, raw string) (int64, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}

func parseAllowedTTLs(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	values := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, fmt.Errorf("FLICK_ALLOWED_TTL_SECONDS must not contain empty entries")
		}
		value, err := parsePositiveInt("FLICK_ALLOWED_TTL_SECONDS", trimmed)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[value]; ok {
			return nil, fmt.Errorf("FLICK_ALLOWED_TTL_SECONDS contains duplicate value %d", value)
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}

	return values, nil
}

func containsInt(values []int, needle int) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
