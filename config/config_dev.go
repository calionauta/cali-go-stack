//go:build dev

package config

import "os"

func init() {
	defaultDev := map[string]string{
		"ENVIRONMENT":   "development",
		"LOG_LEVEL":     "DEBUG",
		"DATABASE_PATH": "data/dev.db",
	}

	for k, v := range defaultDev {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}
