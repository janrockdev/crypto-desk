package cmd

import (
	"os"
	"testing"
)

func TestVarAlpaca(t *testing.T) {
	baseURL := os.Getenv("ALPACAAPI")
	apiKey := os.Getenv("ALPACAKEY")
	apiSecret := os.Getenv("ALPACASEC")
	if apiKey == "" || apiSecret == "" || baseURL == "" {
		t.Fatalf("Missing environmental variables for connection!")
	}
}
