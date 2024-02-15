package request

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Harder, Better, Faster, Stronger by Daft Punk
const SNG_ID = "3135556"

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

	trackInfo, err := GetTrackInfo(SNG_ID)
	if err != nil {
		t.Fatalf("Error getting track info: %v", err)
	}

	assert.NotNil(t, trackInfo)
	log.Printf("Track info: %v", trackInfo)
}
