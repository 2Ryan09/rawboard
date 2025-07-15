package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== Database Environment Variables Debug ===")

	envVars := []string{
		"VALKEY_URI",
		"REDIS_URL",
		"DATABASE_URL",
		"VALKEY_URL",
		"REDIS_HOST",
		"REDIS_PORT",
	}

	found := false
	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if value != "" {
			// Mask credentials for security
			maskedValue := maskCredentials(value)
			fmt.Printf("✅ %s = %s\n", envVar, maskedValue)
			found = true
		} else {
			fmt.Printf("❌ %s = (not set)\n", envVar)
		}
	}

	if !found {
		fmt.Println("\n⚠️  No database environment variables found!")
		fmt.Println("Expected one of: VALKEY_URI, REDIS_URL, DATABASE_URL, VALKEY_URL, or REDIS_HOST+REDIS_PORT")
	}

	fmt.Println("\n=== End Debug Info ===")
}

func maskCredentials(value string) string {
	// Simple masking - replace everything between :// and @ with ***
	if len(value) < 10 {
		return "***"
	}

	// Find protocol separator
	protocolEnd := -1
	for i := 0; i < len(value)-3; i++ {
		if value[i:i+3] == "://" {
			protocolEnd = i + 3
			break
		}
	}

	if protocolEnd == -1 {
		return "***"
	}

	// Find @ symbol for credentials
	atIndex := -1
	for i := protocolEnd; i < len(value); i++ {
		if value[i] == '@' {
			atIndex = i
			break
		}
	}

	if atIndex == -1 {
		// No credentials, just show host part
		return value[:protocolEnd] + "***"
	}

	// Has credentials, mask them
	return value[:protocolEnd] + "***" + value[atIndex:]
}
