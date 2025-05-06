# GoFi

GoFi is a Go implementation of a music download tool. It allows you to search and download music from Deezer, with Spotify integration for finding tracks.

## Features

- Search for tracks, albums, artists and playlists
- Download high-quality music (MP3, FLAC)
- Support for downloading by name or ID
- **Spotify integration** for finding and downloading Spotify content via Deezer
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

### Setting Up Authentication

#### Deezer Authentication

To use GoFi with Deezer, you need a Deezer ARL token. This token is used to authenticate with Deezer's API.

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

### Authenticating with Spotify

Before using Spotify features, you need to authenticate:

```bash
./d-fi auth spotify
```

This will start the OAuth flow and open a browser window for you to authorize the application.

### Basic Commands

#### Downloading from Spotify URLs

You can download tracks, albums, or playlists directly from Spotify URLs:

```bash
# Download a Spotify track
./d-fi download https://open.spotify.com/track/2YarjDYjBJuH63dUIh9OWv

# Download a Spotify album
./d-fi download https://open.spotify.com/album/1DFixLWuPkv3KT3TnV35m3

# Download a Spotify playlist
./d-fi download https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M
```

The application will search for matching content on Deezer and download it.

#### Downloading from Deezer

You can continue to download directly from Deezer:

```bash
# By ID (direct download)
./d-fi download track 3135556
./d-fi download album 302127
./d-fi download playlist 1234567890
```

### Quality Settings

GoFi supports different quality settings for music downloads:

- **FLAC (9)**: Lossless audio format, highest quality
- **MP3 320kbps (3)**: High quality MP3 (default)
- **MP3 128kbps (1)**: Standard quality MP3

Specify quality with the `-q` flag:

```bash
./d-fi -q 9 download https://open.spotify.com/track/2YarjDYjBJuH63dUIh9OWv
```

### Advanced Options

- `-o, --output`: Set the download directory
- `-q, --quality`: Set the audio quality (1, 3, or 9)
- `-l, --log-level`: Set the log level (debug, info, warn, error)

```bash
./d-fi -o ./my-music -q 9 -l debug download https://open.spotify.com/album/1DFixLWuPkv3KT3TnV35m3
```

## Troubleshooting

- If you get Spotify authentication errors, run `./d-fi auth spotify` to re-authenticate
- If downloads fail, try increasing the log level with `-l debug` for more information
- ARL tokens expire periodically, so you may need to obtain a new one if you haven't used the tool in a while

## Contributing

Contributions to this project are welcome. If you find any bugs or have suggestions for new features, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).