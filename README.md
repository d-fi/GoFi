# GoFi

GoFi is a Go port of the d-fi / d-fi-core Deezer tooling. It can be used as a
CLI for downloading Deezer content, or as a Go library for resolving metadata,
searching, converting supported links, and downloading tracks.

## Install

```sh
go install github.com/d-fi/GoFi/cmd/d-fi@latest
```

Or build from a local checkout:

```sh
go build ./cmd/d-fi
```

## Authentication

Most Deezer API and download calls require an ARL cookie. The CLI reads it from
`DEEZER_ARL` first, then from `d-fi.config.json`.

```sh
export DEEZER_ARL="your_deezer_arl"
```

You can also write it to the config file:

```sh
d-fi --set-arl "your_deezer_arl"
```

## CLI Usage

Download from a Deezer URL:

```sh
d-fi "https://www.deezer.com/album/302127"
```

Choose quality non-interactively:

```sh
d-fi --quality flac --url "https://www.deezer.com/track/3135556" --headless
```

Download multiple URLs from a file:

```sh
d-fi --quality 320 --input-file urls.txt --headless
```

Search interactively:

```sh
d-fi "artist:Daft Punk"
d-fi "album:Discovery"
d-fi "playlist:Workout"
d-fi "Harder Better Faster Stronger"
```

Useful flags:

```text
-q, --quality <quality>       128, 320, or flac
-o, --output <template>       Output filename template
-u, --url <url>               Deezer album/artist/playlist/track URL
-i, --input-file <file>       Download all URLs listed in a text file
-c, --concurrency <number>    Parallel downloads for albums, artists, playlists
-a, --set-arl <string>        Save ARL cookie to config
-d, --headless                Run without interactive prompts
-conf, --config-file <file>   Config file path
-rfp, --resolve-full-path     Use absolute paths in generated playlists
-cp, --create-playlist        Force .m3u8 creation for non-playlist downloads
```

## Config

Default config file: `d-fi.config.json`

```json
{
  "concurrency": 4,
  "saveLayout": {
    "track": "Music/{ALB_TITLE}/{SNG_TITLE}",
    "album": "Music/{ALB_TITLE}/{SNG_TITLE}",
    "artist": "Music/{ALB_TITLE}/{SNG_TITLE}",
    "playlist": "Playlist/{TITLE}/{SNG_TITLE}"
  },
  "playlist": {
    "resolveFullPath": false
  },
  "trackNumber": true,
  "fallbackTrack": true,
  "fallbackQuality": true,
  "coverSize": {
    "128": 500,
    "320": 500,
    "flac": 1000
  },
  "cookies": {
    "arl": ""
  }
}
```

Config fields:

- `concurrency`: number of tracks to download at the same time for album,
  artist, and playlist downloads. Higher values can be faster, but very high
  values may be less reliable.
- `saveLayout`: output path templates. The selected template depends on the
  resolved input type:
  - `track`: single-track downloads
  - `album`: album downloads
  - `artist`: artist downloads
  - `playlist`: playlist downloads
- `playlist.resolveFullPath`: when `true`, generated `.m3u8` playlists contain
  absolute file paths. When `false`, playlist entries are relative to the
  playlist file location.
- `trackNumber`: when `true`, GoFi prefixes saved tracks with track position,
  such as `01 - Title`. When `false`, the number prefix is omitted unless the
  layout explicitly uses a track-number placeholder.
- `fallbackTrack`: when `true`, GoFi can download Deezer's fallback track when
  the requested track was moved or replaced and the fallback is available.
- `fallbackQuality`: when `true`, GoFi falls back to a lower available quality
  when the requested quality is unavailable. For example, FLAC may fall back to
  MP3 320, or MP3 320 may fall back to MP3 128.
- `coverSize`: album cover size used for metadata tagging:
  - `128`: cover size for MP3 128 downloads
  - `320`: cover size for MP3 320 downloads
  - `flac`: cover size for FLAC downloads
- `cookies.arl`: saved Deezer ARL cookie. `DEEZER_ARL` takes priority over this
  value when both are present.

Common `saveLayout` placeholders:

```text
{ALB_TITLE}        Album title
{ART_NAME}         Artist name
{SNG_TITLE}        Track title
{TRACK_NUMBER}     Force track number in this position
{NO_TRACK_NUMBER}  Disable automatic track number for this layout
{TITLE}            Playlist title, only available for playlist layout
```

Nested values can be accessed with dot notation, including array indexes, like
`{ARTISTS.0.ART_NAME}` or `{SNG_CONTRIBUTORS.main_artist.0}`.

Example:

```json
{
  "saveLayout": {
    "album": "Music/{ARTISTS.0.ART_NAME}/{ALB_TITLE}/{TRACK_NUMBER} - {SNG_TITLE}",
    "playlist": "Playlist/{TITLE}/{SNG_TITLE}"
  }
}
```

## Library Usage

Install the module:

```sh
go get github.com/d-fi/GoFi
```

Initialize the Deezer API session:

```go
package main

import (
	"log"
	"os"

	"github.com/d-fi/GoFi/request"
)

func main() {
	if _, err := request.InitDeezerAPI(os.Getenv("DEEZER_ARL")); err != nil {
		log.Fatal(err)
	}
}
```

Fetch metadata:

```go
track, err := api.GetTrackInfo("3135556")
album, err := api.GetAlbumInfo("302127")
playlist, err := api.GetPlaylistInfo("908622995")
artist, err := api.GetArtistInfo("27")
```

Search:

```go
results, err := api.SearchMusic("Daft Punk", 10, "TRACK", "ALBUM", "ARTIST")
```

Resolve supported links into Deezer tracks:

```go
parsed, err := converter.ParseInfo("https://www.deezer.com/album/302127")
if err != nil {
	log.Fatal(err)
}
for _, track := range parsed.Tracks {
	log.Println(track.SNG_TITLE)
}
```

Supported converter inputs include Deezer, Spotify, Tidal, YouTube, ISRC, and
UPC helpers.

Download a tagged track to a file:

```go
path, err := download.DownloadTrack(download.DownloadTrackOptions{
	SngID:     "3135556",
	Quality:   3, // 1 = MP3 128, 3 = MP3 320, 9 = FLAC
	CoverSize: 500,
	SaveToDir: "Music",
})
if err != nil {
	log.Fatal(err)
}
log.Println(path)
```

Download to memory with tags:

```go
buf, err := download.DownloadTrackToBuffer(context.Background(), download.DownloadTrackToBufferOptions{
	SngID:     "3135556",
	Quality:   3,
	CoverSize: 500,
})
```

Download to memory without metadata:

```go
buf, err := download.DownloadTrackWithoutMetadata(context.Background(), download.DownloadTrackWithoutMetadataOptions{
	SngID:   "3135556",
	Quality: 3,
})
```

Quality values:

```text
1 = MP3 128 kbps
3 = MP3 320 kbps
9 = FLAC
```

## Development

```sh
go test ./...
go build ./cmd/d-fi
staticcheck ./...
```

## License

This project is licensed under the [MIT License](LICENSE).
