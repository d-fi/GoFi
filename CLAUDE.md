# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoFi is a Go implementation of a music download tool, originally for Deezer but expanding to support multiple services including Spotify. It allows users to search and download music from these platforms through a command-line interface.

## Build Commands

```bash
# Build the new CLI with Spotify support
make build-cli

# Build and install locally
make install                # Install to /usr/local/bin
make install PREFIX=$HOME/.local  # Install to custom directory
make install-dev           # Create symlink for development

# Build the legacy CLI (Deezer only)
make build

# Build all binaries
make build-all

# Run tests
make test

# Clean build artifacts
make clean

# Uninstall
make uninstall

# Build with version injection
go build -ldflags "-X github.com/d-fi/GoFi/cmd/gofi/cmd.version=$(git describe --tags --always)" -o gofi ./cmd/gofi
```

## Installation Methods

### For End Users

```bash
# One-liner install (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/d-fi/GoFi/main/scripts/install.sh | bash

# One-liner install (Windows PowerShell)
iwr -useb https://raw.githubusercontent.com/d-fi/GoFi/main/scripts/install.ps1 | iex
```

### Installation Scripts

- `scripts/install.sh`: Cross-platform bash script that detects OS/arch and downloads appropriate binary
- `scripts/install.ps1`: PowerShell script for Windows installation with PATH management

## Usage Examples

```bash
# Authenticate with Deezer (automatic browser cookie detection)
gofi auth deezer

# Authenticate with Spotify (required before using Spotify features)
gofi auth spotify

# Download a Spotify track and convert to FLAC
gofi -q 9 download https://open.spotify.com/track/2YarjDYjBJuH63dUIh9OWv

# Download a Spotify album to a specific directory
gofi -o ~/Music/Downloads -q 3 download https://open.spotify.com/album/1DFixLWuPkv3KT3TnV35m3

# Download a Spotify playlist in FLAC quality to a specific directory
gofi -o ~/Music/Playlists -q 9 download https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M

# Download from Deezer URL (no Spotify auth needed)
gofi -q 9 download https://www.deezer.com/track/3135556

# Check version
gofi --version
```

## Code Architecture

### Core Components

1. **API Clients**: Service-specific API clients for fetching music data
   - Deezer: `api/api.go`
   - Spotify: `internal/services/spotify/spotify.go`
   - Service Matching: `api/spotify_deezer_match.go` matches content between services

2. **Authentication**:
   - Deezer: 
     - Automatic browser cookie detection: `internal/auth/browser_cookies.go`
     - Supports Chrome, Firefox, Safari, Edge, and Arc browsers
     - ARL helper functions: `internal/auth/arl_helper.go`
     - CLI command: `cmd/gofi/cmd/auth_deezer.go`
     - No longer requires manual DEEZER_ARL environment variable
   - Spotify: 
     - OAuth2 flow in `internal/services/spotify/auth.go`
     - Stores tokens in `~/.config/gofi/spotify_token.json`
     - CLI command: `cmd/gofi/cmd/auth_spotify.go`

3. **Download Engine**: `download/download.go` handles the actual download of music files
   - Concurrent downloads for albums/playlists
   - Quality selection (FLAC, MP3 320kbps, MP3 128kbps)
   - File existence checking to avoid re-downloads
   - Improved error handling with retry logic

4. **Metadata Management**: Adds appropriate metadata to downloaded files
   - ID3 tags for MP3: `metadata/id3_tag.go`
   - FLAC metadata: `metadata/flac_meta.go`
   - Album cover art: `metadata/album_cover.go`

5. **CLI Interface**:
   - Legacy flag-based CLI: `cmd/main.go` (Deezer only)
   - New Cobra-based CLI: `cmd/gofi/cmd/` (includes Spotify support)
   - Improved download handler: `cmd/gofi/cmd/download_handler_improved.go`
   - Version support with automatic injection during build

6. **User Interface**: Beautiful terminal output
   - Display manager: `internal/ui/display.go`
   - Custom progress bars: `internal/ui/simple_progress.go`
   - Color-coded output using `github.com/fatih/color`
   - Icons and visual feedback for better UX

### Data Flow

#### For Spotify URLs:
1. User provides a Spotify URL
2. URL parsing (`internal/utils/urlparser.go`) identifies the content type (track, album, playlist)
3. Spotify API fetches metadata through OAuth2 authentication
4. The Spotify content is matched to equivalent Deezer content (`api/spotify_deezer_match.go`)
5. Deezer API is used to download the matched tracks
6. Metadata is added to downloaded files
7. Files are saved according to the format: `<artist> - <track>.<ext>`

#### For Deezer URLs:
1. User provides a Deezer URL
2. URL parsing (`internal/utils/urlparser.go`) identifies the content type (track, album, playlist)
3. Deezer API fetches content directly (no Spotify auth needed)
4. Content is downloaded using the Deezer API
5. Metadata is added to downloaded files
6. Files are saved according to the format: `<artist> - <track>.<ext>`

## Configuration

The application accepts configuration through:
1. Command-line flags
2. Environment variables in a `.env` file
3. JSON configuration file

Main environment variables:

Authentication:
- `SPOTIFY_CLIENT_ID`: Spotify API client ID
- `SPOTIFY_CLIENT_SECRET`: Spotify API client secret
- `DEEZER_ARL`: Deezer authentication token (optional - can be auto-detected)

Configuration (Priority: CLI flags > Environment variables > Default values):
- `GOFI_OUTPUT_DIR`: Default download directory (default: "./downloads")
- `GOFI_QUALITY`: Default audio quality - 1, 3, or 9 (default: 3)
- `GOFI_LOG_LEVEL`: Default log level - debug, info, warn, error (default: "info")

## Important File Paths

- **Spotify Authentication**: `~/.config/gofi/spotify_token.json` - Stores OAuth2 tokens
- **Environment Variables**: `.env` file in project root directory
- **Binary Location**: `gofi` in project root (excluded from git)

## Cover Size Settings

When downloading music, ensure appropriate cover sizes are used based on quality:
- MP3 128kbps (quality=1): 500px
- MP3 320kbps (quality=3): 500px
- FLAC (quality=9): 1000px

## Testing

Tests no longer require DEEZER_ARL environment variable. They will:
1. Try to get ARL from environment variable
2. Try to get ARL from browser cookies
3. Skip tests if no valid ARL is available

Run all tests:
```bash
make test
```

Run specific tests:
```bash
go test -v ./path/to/package -run TestName
```

Current test coverage focuses on:
- URL handling and parsing
- API client functionality
- Metadata management
- Browser cookie extraction
- ARL token validation

## CI/CD

### GitHub Actions Workflows

1. **test.yml**: Runs on all pushes and PRs to main
   - Sets up Go 1.23
   - Runs all tests (no longer needs DEEZER_ARL secret)
   
2. **release.yml**: Triggered by version tags (e.g., v1.0.0)
   - Builds binaries for multiple platforms:
     - macOS (Intel and Apple Silicon)
     - Linux (AMD64 and ARM64)
     - Windows (AMD64)
   - Injects version number into binaries
   - Creates GitHub release with all binaries

## Recent Improvements

1. **Automatic ARL Detection**: 
   - No manual token configuration needed
   - Supports multiple browsers
   - Secure cookie extraction with proper decryption

2. **Installation Scripts**:
   - Platform detection and automatic binary selection
   - Proper permission handling
   - PATH management on Windows

3. **Build System**:
   - Version injection during build
   - Multiple installation targets
   - Development-friendly symlink option

4. **Testing**:
   - Tests work without environment variables
   - Better error handling for invalid tokens
   - Graceful test skipping

5. **Error Handling**:
   - Improved error messages throughout
   - Better handling of API responses
   - Retry logic for network failures

## Development Tips

1. Use `make install-dev` during development to create a symlink that updates automatically
2. The `gofi` binary is git-ignored, so it won't be committed
3. Version is automatically set from git tags during build
4. Tests will skip if no valid ARL is available (from env or cookies)
5. Always clean ARL tokens to remove control characters before use

## Current Status

The project now includes:
- Full Spotify integration alongside Deezer support
- Automatic URL detection for both Spotify and Deezer URLs
- Beautiful CLI interface with progress bars and colored output
- Smart file management that skips already downloaded files
- Improved error handling and user feedback
- Automatic browser-based authentication for Deezer
- Cross-platform installation scripts
- GitHub Actions for automated testing and releases

Users can authenticate with Spotify, browse Spotify content, and download matched content from Deezer in high quality formats. Deezer URLs work directly without requiring Spotify authentication.