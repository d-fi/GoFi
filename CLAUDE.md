# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoFi is a Go implementation of a music download tool, originally for Deezer but expanding to support multiple services including Spotify. It allows users to search and download music from these platforms through a command-line interface.

## Build Commands

```bash
# Build the application
make build

# Run tests
make test

# Clean build artifacts
make clean

# Run a specific test
go test -v ./path/to/package -run TestName
```

## Code Architecture

### Core Components

1. **API Clients**: Service-specific API clients for fetching music data
   - Deezer: `api/api.go`
   - Spotify: `internal/services/spotify/spotify.go`

2. **Authentication**:
   - Deezer: Uses ARL token stored in configuration
   - Spotify: OAuth2 flow in `internal/services/spotify/auth.go`

3. **Download Engine**: `download/download.go` handles the actual download of music files
   - Concurrent downloads for albums/playlists
   - Quality selection (FLAC, MP3 320kbps, MP3 128kbps)

4. **Metadata Management**: Adds appropriate metadata to downloaded files
   - ID3 tags for MP3: `metadata/id3_tag.go`
   - FLAC metadata: `metadata/flac_meta.go`
   - Album cover art: `metadata/album_cover.go`

5. **CLI Interface**:
   - Legacy flag-based CLI: `cmd/main.go`
   - New Cobra-based CLI: `cmd/gofi/cmd/`

### Data Flow

1. User provides a URL or search query
2. URL parsing (`internal/utils/urlparser.go`) identifies service type and content ID
3. Service-specific API fetches metadata (track, album, artist, playlist info)
4. Download engine retrieves audio files with appropriate quality
5. Metadata is added to downloaded files
6. Files are saved according to configured path templates

## Configuration

The application accepts configuration through:
1. Command-line flags
2. Environment variables
3. JSON configuration file

Main configuration options:
- Authentication tokens for music services
- Download quality preferences
- Output path templates
- Concurrency settings

## Feature Flags

- `fallbackQuality`: If true, falls back to lower quality when requested quality is unavailable
- `fallbackTrack`: If true, attempts to find alternative tracks when the requested track is unavailable
- `trackNumber`: If true, includes track numbers in filenames

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

The project is in the process of extending support beyond Deezer to include Spotify, with additional services planned. The `feat/spotify-integration` branch contains the latest work on Spotify integration.