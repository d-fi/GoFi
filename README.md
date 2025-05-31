# GoFi

GoFi is a Go implementation of a music download tool. It allows you to download music from both Deezer and Spotify URLs.

## Features

- **Automatic URL detection** - Works with both Spotify and Deezer URLs
- **Beautiful CLI interface** - Clean progress bars, colored output, and organized display
- Download high-quality music (MP3 128/320kbps, FLAC lossless)
- Support for tracks, albums, and playlists from both services
- **Spotify integration** - Spotify content is matched and downloaded from Deezer
- **Direct Deezer downloads** - No Spotify authentication needed for Deezer URLs
- **Smart file management** - Skips already downloaded files, organized folder structure
- Command-line interface for easy automation
- Written in Go for high performance and cross-platform compatibility
- Support for configuration files and environment variables
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
# Build the new CLI with URL auto-detection
go build -o gofi ./cmd/gofi

# Or use the Makefile for the legacy CLI
make build
```

This will create a binary named `gofi` in the project root directory.

## Usage

### Setting Up Authentication

#### Deezer Authentication

To use GoFi with Deezer, you need a Deezer ARL token. This token is used to authenticate with Deezer's API.

**Option 1: Automatic Browser Cookie Detection (Recommended)**

```bash
./gofi auth deezer
```

This command will automatically:
- Search for the ARL cookie in your installed browsers (Chrome, Firefox, Edge, Safari)
- Extract and validate the token
- Save it to your `.env` file

**Option 2: Manual Token Retrieval**

1. Log in to [Deezer](https://www.deezer.com) in your web browser
2. Open developer tools (F12 or right-click and select "Inspect")
3. Go to the "Application" tab (in Chrome) or "Storage" tab (in Firefox)
4. Look for cookies (under "Storage" -> "Cookies" -> "https://www.deezer.com")
5. Find the cookie named "arl" and copy its value (should be 192 characters)

#### Spotify Authentication

To use GoFi with Spotify, you need to register an application with Spotify:

1. Go to the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Create a new application
3. Set the redirect URI to `http://localhost:8888/callback`
4. Note your Client ID and Client Secret
5. Add these credentials to your `.env` file (see below)

### Environment Configuration

Create a `.env` file in the project root with the following content:

```
# Spotify API credentials
SPOTIFY_CLIENT_ID=your_spotify_client_id_here
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret_here

# Deezer authentication
DEEZER_ARL=your_deezer_arl_here
```

Alternatively, you can set these as environment variables:

```bash
export SPOTIFY_CLIENT_ID=your_spotify_client_id_here
export SPOTIFY_CLIENT_SECRET=your_spotify_client_secret_here
export DEEZER_ARL=your_deezer_arl_here
```

### Authenticating with Services

#### Spotify Authentication

Before using Spotify features, you need to authenticate:

```bash
./gofi auth spotify
```

This will start the OAuth flow and open a browser window for you to authorize the application.

#### Deezer Authentication

To set up Deezer authentication automatically:

```bash
./gofi auth deezer
```

This will find and save your ARL token from your browser cookies.

### Basic Commands

#### Downloading with URLs

GoFi automatically detects whether you're using a Spotify or Deezer URL:

```bash
# Download from Spotify URLs (requires Spotify authentication)
./gofi download https://open.spotify.com/track/2YarjDYjBJuH63dUIh9OWv
./gofi download https://open.spotify.com/album/1DFixLWuPkv3KT3TnV35m3
./gofi download https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M

# Download from Deezer URLs (no Spotify auth needed)
./gofi download https://www.deezer.com/track/3135556
./gofi download https://www.deezer.com/album/302127
./gofi download https://www.deezer.com/playlist/1234567890

# Specify output directory and quality
./gofi -o ~/Music/Downloads -q 9 download https://www.deezer.com/playlist/13872511521
```

For Spotify URLs, the app will search for matching content on Deezer and download it.
For Deezer URLs, content is downloaded directly.

### User Interface

GoFi features a polished command-line interface with:

- **Progress Bars**: Real-time download progress with speed and ETA
- **Colored Output**: Easy-to-read color-coded messages
- **Status Icons**: Visual feedback with ✓ for success, ✗ for errors, ℹ for info
- **Smart Display**: Clean, organized output that doesn't clutter your terminal
- **Download Summary**: Clear summary of successful and failed downloads

### Quality Settings

GoFi supports different quality settings for music downloads:

- **FLAC (9)**: Lossless audio format, highest quality
- **MP3 320kbps (3)**: High quality MP3 (default)
- **MP3 128kbps (1)**: Standard quality MP3

Specify quality with the `-q` flag:

```bash
./gofi -q 9 download https://open.spotify.com/track/2YarjDYjBJuH63dUIh9OWv
```

### Advanced Options

- `-o, --output`: Set the download directory
- `-q, --quality`: Set the audio quality (1, 3, or 9)
- `-l, --log-level`: Set the log level (debug, info, warn, error)

```bash
./gofi -o ./my-music -q 9 -l debug download https://open.spotify.com/album/1DFixLWuPkv3KT3TnV35m3
```

## Troubleshooting

- If you get Spotify authentication errors, run `./gofi auth spotify` to re-authenticate
- If downloads fail, try increasing the log level with `-l debug` for more information
- ARL tokens expire periodically, so you may need to obtain a new one if you haven't used the tool in a while

## Contributing

Contributions to this project are welcome. If you find any bugs or have suggestions for new features, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).