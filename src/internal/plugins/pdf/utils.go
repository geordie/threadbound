package pdf

import (
	"os"
)

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// writeToFile writes data to a file
func writeToFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// readFile reads data from a file
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// removeFile removes a file if it exists
func removeFile(path string) {
	os.Remove(path) // Ignore errors for cleanup
}