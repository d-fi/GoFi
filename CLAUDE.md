# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoFi is a Go implementation of a music download tool, originally for Deezer but expanding to support multiple services including Spotify. It allows users to search and download music from these platforms through a command-line interface.

## Build Commands

```bash
# Build the application (legacy CLI)
make build

# Build the new CLI with Spotify support
go build -o gofi ./cmd/gofi

# Run tests
make test

# Clean build artifacts
make clean

# Run a specific test
go test -v ./path/to/package -run TestName
```

## Usage Examples

```bash
# Authenticate with Spotify (required before using Spotify features)
./gofi auth spotify

# Download a Spotify track and convert to FLAC
./gofi -q 9 download https://open.spotify.com/track/2YarjDYjBJuH63dUIh9OWv

# Download a Spotify album to a specific directory
./gofi -o ~/Music/Downloads -q 3 download https://open.spotify.com/album/1DFixLWuPkv3KT3TnV35m3

# Download a Spotify playlist in FLAC quality to a specific directory
./gofi -o ~/Music/Playlists -q 9 download https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M
```

## Code Architecture

### Core Components

1. **API Clients**: Service-specific API clients for fetching music data
   - Deezer: `api/api.go`
   - Spotify: `internal/services/spotify/spotify.go`
   - Service Matching: `api/spotify_deezer_match.go` matches content between services

2. **Authentication**:
   - Deezer: Uses ARL token stored in configuration or from environment variable
     - Automatic browser cookie detection: `internal/auth/browser_cookies.go`
     - ARL helper functions: `internal/auth/arl_helper.go`
     - CLI command: `cmd/gofi/cmd/auth_deezer.go`
   - Spotify: OAuth2 flow in `internal/services/spotify/auth.go`

3. **Download Engine**: `download/download.go` handles the actual download of music files
   - Concurrent downloads for albums/playlists
   - Quality selection (FLAC, MP3 320kbps, MP3 128kbps)
   - File existence checking to avoid re-downloads

4. **Metadata Management**: Adds appropriate metadata to downloaded files
   - ID3 tags for MP3: `metadata/id3_tag.go`
   - FLAC metadata: `metadata/flac_meta.go`
   - Album cover art: `metadata/album_cover.go`

5. **CLI Interface**:
   - Legacy flag-based CLI: `cmd/main.go` (Deezer only)
   - New Cobra-based CLI: `cmd/gofi/cmd/` (includes Spotify support)
   - Improved download handler: `cmd/gofi/cmd/download_handler_improved.go`

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
- `SPOTIFY_CLIENT_ID`: Spotify API client ID
- `SPOTIFY_CLIENT_SECRET`: Spotify API client secret
- `DEEZER_ARL`: Deezer authentication token

## Important File Paths

- **Spotify Authentication**: `~/.config/gofi/spotify_token.json` - Stores OAuth2 tokens
- **Environment Variables**: `.env` file in project root directory

## Cover Size Settings

When downloading music, ensure appropriate cover sizes are used based on quality:
- MP3 128kbps (quality=1): 500px
- MP3 320kbps (quality=3): 500px
- FLAC (quality=9): 1000px

## Testing

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

## Current Development

The project now includes:
- Full Spotify integration alongside Deezer support
- Automatic URL detection for both Spotify and Deezer URLs
- Beautiful CLI interface with progress bars and colored output
- Smart file management that skips already downloaded files
- Improved error handling and user feedback

Users can authenticate with Spotify, browse Spotify content, and download matched content from Deezer in high quality formats. Deezer URLs work directly without requiring Spotify authentication.