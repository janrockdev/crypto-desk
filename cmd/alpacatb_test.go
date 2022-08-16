package cmd

import (
	"os"
	"testing"
)

func TestVarAlpaca(t *testing.T) {
	apiKey := os.Getenv("ALPACAKEY")
	apiSecret := os.Getenv("ALPACASEC")
	if apiKey == "" || apiSecret == "" {
		t.Fatalf("Missing environmental variables for connection!")
	}
}
