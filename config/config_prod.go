//go:build !dev

package config

import "os"

func init() {
	if os.Getenv("ENVIRONMENT") == "" {
		os.Setenv("ENVIRONMENT", "production")
	}
}
