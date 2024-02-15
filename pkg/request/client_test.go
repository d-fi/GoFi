package request

import (
	"os"
	"testing"
)

func TestInitDeezerAPI(t *testing.T) {
	arl := os.Getenv("DEEZER_ARL")

	session, err := InitDeezerAPI(arl)
	if err != nil {
		t.Fatalf("Error initializing Deezer API: %v", err)
	}

	if session == "" {
		t.Error("Session is empty")
	}

	t.Logf("Deezer API session: %s", session)
}
