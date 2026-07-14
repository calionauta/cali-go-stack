//go:build !dev

// SCOPE:core - DO NOT REMOVE - Production defaults.
package config

import "os"

func init() {
	if os.Getenv("ENVIRONMENT") == "" {
		os.Setenv("ENVIRONMENT", "production")
	}
}
