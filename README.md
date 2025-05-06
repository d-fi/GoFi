# GoFi

GoFi is a Go implementation of a Deezer music download tool. It allows you to search and download music from Deezer.

## Features

- Search for tracks, albums, artists and playlists
- Download high-quality music (MP3, FLAC)
- Support for downloading by name or ID
- Command-line interface for easy automation
- Written in Go for high performance and cross-platform compatibility
- Support for configuration files
- Concurrent downloads for albums and playlists

## Installation

### Prerequisites

- Go 1.23 or later

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/d-fi/GoFi.git
cd GoFi
```

2. Build the application:
```bash
make build
```

This will create a binary named `d-fi` in the project root directory.

## Usage

To use GoFi, you need a Deezer ARL token. This token is used to authenticate with Deezer's API.

### Getting an ARL Token

1. Log in to [Deezer](https://www.deezer.com) in your web browser
2. Open developer tools (F12 or right-click and select "Inspect")
3. Go to the "Application" tab (in Chrome) or "Storage" tab (in Firefox)
4. Look for cookies (under "Storage" -> "Cookies" -> "https://www.deezer.com")
5. Find the cookie named "arl" and copy its value (should be 192 characters)

### Setting up the ARL Token

You can provide the ARL token in one of three ways:

1. Using the `-arl` command-line flag:
```bash
./d-fi -arl=YOUR_ARL_TOKEN search "artist name"
```

2. Using the `DEEZER_ARL` environment variable:
```bash
export DEEZER_ARL=YOUR_ARL_TOKEN
./d-fi search "artist name"
```

3. Using a configuration file (see Configuration section below)

### Configuration File

GoFi supports using a configuration file for more advanced settings. Create a JSON file with the following structure:

```json
{
  "concurrency": 3,
  "saveLayout": {
    "track": "/path/to/save/Tracks/{ART_NAME}/{ART_NAME} - {SNG_TITLE}",
    "album": "/path/to/save/Albums/{ART_NAME}/{ALB_TITLE}/{TRACK_NUMBER} - {SNG_TITLE}",
    "artist": "/path/to/save/Artists/{ART_NAME}/{SNG_TITLE}",
    "playlist": "/path/to/save/Playlists/{TITLE}/{ART_NAME} - {SNG_TITLE}"
  },
  "playlist": {
    "resolveFullPath": false
  },
  "trackNumber": false,
  "fallbackTrack": true,
  "fallbackQuality": false,
  "coverSize": {
    "128": 500,
    "320": 500,
    "flac": 1000
  },
  "cookies": {
    "arl": "YOUR_ARL_TOKEN_HERE"
  }
}
```

Path templates support the following placeholders:
- `{ART_NAME}`: Artist name
- `{SNG_TITLE}`: Song title
- `{ALB_TITLE}`: Album title
- `{TRACK_NUMBER}`: Track number (formatted as 01, 02, etc.)
- `{SNG_ID}`: Song ID
- `{ALB_ID}`: Album ID
- `{TITLE}`: Playlist title

To use a configuration file:
```bash
./d-fi -config=path/to/config.json search "artist name"
```

### Basic Commands

1. Search for music:
```bash
./d-fi search "artist or track name"
```

2. Download a track (by name or ID):
```bash
# By name (will search and let you select from results)
./d-fi download track "Harder Better Faster Stronger"

# By ID (direct download)
./d-fi download track 3135556
```

3. Download an album (by name or ID):
```bash
# By name (will search and let you select from results)
./d-fi download album "Discovery"

# By ID (direct download)
./d-fi download album 302127
```

4. Download a playlist (by name or ID):
```bash
# By name (will search and let you select from results)
./d-fi download playlist "Top 50 Global"

# By ID (direct download)
./d-fi download playlist 1234567890
```

### Quality Settings

GoFi supports different quality settings for music downloads:

- **FLAC (9)**: Lossless audio format, highest quality (default)
- **MP3 320kbps (3)**: High quality MP3
- **MP3 128kbps (1)**: Standard quality MP3

The application currently defaults to FLAC quality (9). Different account tiers on Deezer have different capabilities:

- Free accounts: Can only download MP3 128kbps
- Premium accounts: Can download MP3 320kbps
- HiFi accounts: Can download FLAC

If a track isn't available in the requested quality, the app can fall back to a lower quality if `fallbackQuality` is set to `true` in your config file.

### Advanced Options

- `-log-level`: Set the log level (debug, info, warn, error)
```bash
./d-fi -log-level=debug search "artist name"
```

## Troubleshooting

- If you get authentication errors, check that your ARL token is valid and correctly entered
- ARL tokens expire periodically, so you may need to obtain a new one if you haven't used the tool in a while
- For detailed logging, use the debug log level: `./d-fi -log-level=debug`
- If downloads fail, try increasing the concurrency value in your config file

## Contributing

Contributions to this project are welcome. If you find any bugs or have suggestions for new features, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).

🔄 TODO: Add more details about quality settings and advanced usage scenarios.
