package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time" // Added for shutdown timeout

	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	// redirectURI is the endpoint Spotify will redirect to after authorization.
	// Needs to match the one registered in the Spotify Developer Dashboard.
	redirectURI = "http://localhost:8888/callback"
	// tokenFileName is the name of the file to store OAuth tokens.
	tokenFileName = "spotify_token.json"
)

var (
	// scopes define the permissions the application requests from the user.
	// Adjust these based on the specific Spotify data you need to access.
	scopes = []string{
		spotifyauth.ScopeUserReadPrivate, // Read user's private information
		// Add other scopes as needed, e.g.:
		// spotifyauth.ScopePlaylistReadPrivate,   // Read user's private playlists
		// spotifyauth.ScopePlaylistReadCollaborative, // Read user's collaborative playlists
		// spotifyauth.ScopeUserLibraryRead,       // Read user's saved tracks/albums
		// spotifyauth.ScopeUserReadEmail,         // Read user's email address
	}
	// state is a random string to protect against CSRF attacks.
	state = uuid.NewString()
)

// AuthService handles Spotify OAuth2 authentication.
type AuthService struct {
	authenticator *spotifyauth.Authenticator
	ch            chan *spotify.Client // Channel to receive the authenticated client
	server        *http.Server
	configDir     string
	tokenPath     string
}

// Config holds Spotify API credentials.
// 🔄 TODO: Load these securely, e.g., from environment variables or a config file.
type Config struct {
	ClientID     string
	ClientSecret string
}

// OAuthConfig returns the OAuth2 config for Spotify authentication.
func (c *Config) OAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  spotifyauth.AuthURL,
			TokenURL: spotifyauth.TokenURL,
		},
		RedirectURL: redirectURI,
		Scopes:      scopes,
	}
}

// NewAuthService creates a new Spotify authentication service.
func NewAuthService(cfg Config) (*AuthService, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}
	tokenPath := filepath.Join(configDir, tokenFileName)

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(scopes...),
		spotifyauth.WithClientID(cfg.ClientID),
		spotifyauth.WithClientSecret(cfg.ClientSecret),
	)

	return &AuthService{
		authenticator: auth,
		ch:            make(chan *spotify.Client),
		configDir:     configDir,
		tokenPath:     tokenPath,
	}, nil
}

// GetAuthURL generates the Spotify authorization URL for the user to visit.
func (s *AuthService) GetAuthURL() string {
	return s.authenticator.AuthURL(state)
}

// getConfigDir finds or creates the application's configuration directory.
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	configDir := filepath.Join(homeDir, ".config", "gofi")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory '%s': %w", configDir, err)
	}
	return configDir, nil
}

// saveToken saves the OAuth2 token to the configuration file.
func (s *AuthService) saveToken(token *oauth2.Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	if err := os.WriteFile(s.tokenPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file '%s': %w", s.tokenPath, err)
	}
	log.Printf("Token saved to %s\n", s.tokenPath)
	return nil
}

// loadToken loads the OAuth2 token from the configuration file.
func (s *AuthService) loadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(s.tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No token file exists yet
		}
		return nil, fmt.Errorf("failed to read token file '%s': %w", s.tokenPath, err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token from '%s': %w", s.tokenPath, err)
	}
	return &token, nil
}

// StartAuthentication initiates the OAuth flow if no token exists or prompts the user.
// It starts a local server to handle the callback.
func (s *AuthService) StartAuthentication(ctx context.Context) (*spotify.Client, error) {
	// Try loading existing token
	token, err := s.loadToken()
	if err != nil {
		log.Printf("Warning: could not load existing token: %v", err)
	}

	// If token exists and is valid (or refreshable), create client
	if token != nil {
		// ‼️ FIXME: The underlying http client in `spotify.New` created via `authenticator.Client`
		// should handle token refreshing automatically using the `oauth2.TokenSource`.
		// We still might want an explicit check here or in GetClient later.
		client := spotify.New(s.authenticator.Client(ctx, token))
		log.Println("Using existing Spotify token.")
		return client, nil
	}

	// No valid token, start the authentication flow
	log.Println("No valid Spotify token found. Starting authentication flow...")
	fmt.Printf(`GoFi needs permission to access your Spotify account.
`)
	fmt.Printf(`Please open the following URL in your browser:

%s

`, s.GetAuthURL())
	fmt.Println("Waiting for authorization...")

	// Start the callback server
	if err := s.startCallbackServer(); err != nil {
		return nil, err
	}
	defer s.stopCallbackServer(ctx) // Ensure server is stopped

	// Wait for the callback handler to send the authenticated client
	select {
	case client := <-s.ch:
		if client == nil {
			return nil, fmt.Errorf("authentication failed during callback")
		}
		log.Println("Spotify authentication successful.")
		return client, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("authentication timed out or was cancelled: %w", ctx.Err())
	}
}

// startCallbackServer starts the local HTTP server to listen for the Spotify callback.
func (s *AuthService) startCallbackServer() error {
	// ‼️ FIXME: Ensure this doesn't clash if another instance is running or port is taken.
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)
	s.server = &http.Server{
		Addr:    ":8888", // Listen on the port specified in redirectURI
		Handler: mux,
	}

	// Channel to signal server start or error
	startErrCh := make(chan error, 1)

	// Run the server in a separate goroutine so it doesn't block.
	go func() {
		log.Println("Starting callback server on http://localhost:8888")
		startErrCh <- nil // Signal successful start attempt (ListenAndServe blocks)
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Error starting or running callback server: %v", err)
			// Try to send error back if channel is still listened to
			select {
			case startErrCh <- fmt.Errorf("callback server error: %w", err):
			default:
			}
			// Signal failure via the main client channel if server crashes later
			select {
			case s.ch <- nil:
			default: // Avoid blocking if channel is already closed or full
			}
		}
	}()

	// Wait for server to start or error out immediately
	select {
	case err := <-startErrCh:
		if err != nil {
			return err // Return error if server failed to start listening
		}
		// Server start attempted (doesn't guarantee it's fully ready, but ListenAndServe was called)
		log.Println("Callback server listener started.")
		return nil
	case <-time.After(2 * time.Second): // Timeout for server start
		return fmt.Errorf("callback server failed to start within timeout")

	}
}

// stopCallbackServer gracefully shuts down the callback server.
func (s *AuthService) stopCallbackServer(ctx context.Context) {
	if s.server != nil {
		log.Println("Shutting down callback server...")
		// Create a context with timeout for shutdown
		// Use a background context for shutdown if the original ctx is already done.
		shutdownCtx := context.Background()
		if ctx.Err() == nil {
			var cancel context.CancelFunc
			shutdownCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
		} else {
			// If parent context is done, give shutdown a fixed timeout.
			var cancel context.CancelFunc
			shutdownCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
		}

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Callback server shutdown error: %v", err)
		} else {
			log.Println("Callback server stopped.")
		}
		s.server = nil
		close(s.ch) // Close channel after server stops
	}
}

// handleCallback is the HTTP handler for the Spotify redirect URI.
func (s *AuthService) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // Use request context

	// Verify the state matches to prevent CSRF
	receivedState := r.FormValue("state")
	if receivedState != state {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		log.Printf("State mismatch: expected %s, got %s", state, receivedState)
		// Don't send nil to channel here, let the main flow time out or handle missing client
		return
	}

	// Exchange the authorization code for a token
	token, err := s.authenticator.Token(ctx, state, r) // Pass received state back for validation by library? check docs
	if err != nil {
		http.Error(w, "Couldn't get token: "+err.Error(), http.StatusForbidden)
		log.Printf("Error getting token: %v", err)
		// Don't send nil to channel here
		return
	}

	// Save the token
	if err := s.saveToken(token); err != nil {
		log.Printf("Critical Error: Failed to save token: %v", err)
		// This is more critical, maybe return error to user?
		http.Error(w, "Failed to save token", http.StatusInternalServerError)
		// Don't send nil to channel here
		return
	}

	// Create a client using the new token
	client := spotify.New(s.authenticator.Client(ctx, token))

	// Send success message to browser
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html><body><h1>Authentication successful!</h1><p>You can close this window now.</p></body></html>`)
	log.Println("Callback handled successfully.")

	// Send the authenticated client back to the main flow
	// Use non-blocking send in case receiver is gone (e.g., timeout)
	select {
	case s.ch <- client:
		log.Println("Authenticated client sent to channel.")
	default:
		log.Println("Warning: Failed to send authenticated client to channel (receiver not ready or channel closed).")
	}
}

// GetClient provides an authenticated Spotify client.
// It attempts to load a token, refresh it if necessary.
// It does NOT initiate the full authentication flow here; that's handled by StartAuthentication.
func (s *AuthService) GetClient(ctx context.Context) (*spotify.Client, error) {
	token, err := s.loadToken()
	if err != nil {
		// Don't wrap os.IsNotExist, let caller handle that if needed
		return nil, err
	}

	if token == nil {
		// No token exists, authentication is required.
		return nil, fmt.Errorf("spotify token not found, authentication required (run 'gofi auth spotify')")
	}

	// Create a client which will automatically handle refreshing
	client := spotify.New(s.authenticator.Client(ctx, token))

	// 🚀 TODO (OPTIMIZATION): Optional: Verify token validity with a lightweight API call.
	// _, err = client.CurrentUser(ctx)
	// if err != nil {
	//     log.Printf("Warning: Spotify token validation check failed: %v. Token might be expired or invalid.", err)
	//     // Depending on the error, could indicate need for re-auth
	//     return nil, fmt.Errorf("spotify token may be invalid, re-run 'gofi auth spotify': %w", err)
	// }

	log.Println("Using authenticated Spotify client from stored token.")
	return client, nil
}
