package utils

import "os"

// GetEnv retrieves the value of the environment variable named by the key.
// If the variable is present in the environment the value (which may be empty) is returned.
// Otherwise it returns the defaultValue.
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
