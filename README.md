# GoFi

GoFi is a Go port of the d-fi / d-fi-core Deezer tooling. It can be used as a CLI for downloading Deezer content, or as a Go library for resolving metadata, searching, converting supported links, and downloading tracks.

## Install

Download prebuilt CLI binaries from:

https://github.com/d-fi/releases/releases

On Windows, open `d-fi.bat` from the release zip to start the CLI, start the web UI, or save your Deezer ARL.

Or install from source with Go:

```sh
go install github.com/d-fi/GoFi/cmd/d-fi@latest
```

Or build from a local checkout:

```sh
go build ./cmd/d-fi
```

## Authentication

Most Deezer API and download calls require an ARL cookie. The CLI reads it from `DEEZER_ARL` first, then from `d-fi.config.json`.

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

## Web UI

Start the local web UI:

```sh
d-fi web
```

Then open:

```text
http://127.0.0.1:8080
```

Web options:

```sh
d-fi web --addr 127.0.0.1:8080 --config d-fi.config.json
```

The web UI uses the same config format as the CLI. It reads `DEEZER_ARL` first, then `cookies.arl` from the config file. If an ARL is already available, the web server tries to connect automatically on startup. You can also save an ARL from the web UI.

The download flow is:

1. Choose a source type: auto, track, album, artist, or playlist.
2. Enter a Deezer URL, Spotify URL/URI, or search text.
3. Preview the resolved tracks.
4. Select the tracks to download.
5. Choose quality and start the download.

Downloads use the configured `saveLayout`, `trackNumber`, fallback, cover size, and playlist settings. Playlist downloads create `.m3u8` files using `playlist.resolveFullPath`.

The Downloads panel shows progress for active jobs. Active jobs can be canceled. `Clear History` removes finished, failed, and canceled job rows from the web UI. It does not delete downloaded files.

## Config

The config file is optional. By default, GoFi reads `d-fi.config.json` from the directory where you run `d-fi`. You can use another file with:

```sh
d-fi --config-file custom-config.json
```

The file is also created automatically when you save an ARL:

```sh
d-fi --set-arl "your_deezer_arl"
```

You can omit fields you do not need. GoFi merges your config with the defaults below.

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
  "cover": {
    "mode": "embed",
    "fileName": "cover.jpg"
  },
  "cookies": {
    "arl": ""
  }
}
```

Config fields:

### `concurrency`

Number of tracks to download at the same time for album, artist, and playlist downloads. Original d-fi documents this as `1` to `50`. Higher values can be faster on a good connection, but very high values may be less reliable.

### `saveLayout`

Output path templates. The selected template depends on the resolved input type:

```text
saveLayout.track       Single-track downloads
saveLayout.album       Album downloads
saveLayout.artist      Artist downloads
saveLayout.playlist    Playlist downloads
```

You can override the layout for one command with `--output`:

```sh
d-fi --output "{ART_NAME} - {SNG_TITLE}" "https://www.deezer.com/track/3135556"
```

Common placeholders for track, album, artist, and playlist layouts:

```text
{ALB_TITLE}        Album title
{ART_NAME}         Artist name
{SNG_TITLE}        Track title
{DISK_FOLDER}      CD1, CD2, etc. for multi-disc albums, empty for single-disc albums
{DISK_NUMBER}      Disc number from the track metadata
{RELEASE_DATE}     Album release date, such as 2001-03-07. Prefers physical/original dates when Deezer provides them.
{RELEASE_YEAR}     Album release year, such as 2001
{TRACK_NUMBER}     Force track number in this position
{NO_TRACK_NUMBER}  Disable automatic track number for this layout
{TITLE}            Playlist title, only available for playlist layout
```

`{TRACK_NUMBER}` forces the track number at that position. `{NO_TRACK_NUMBER}` disables the automatic number prefix for that layout.

By default, multi-disc album folders keep the previous behavior and append the disc to `{ALB_TITLE}`, such as `Album Name (Disc 01)`. Use `{DISK_FOLDER}` in the layout to opt into a shared album folder with disc subfolders, such as `Album Name/CD1`.

Any field from the track or album metadata can be used as a placeholder. Nested values can be accessed with dot notation, including array indexes, like `{ARTISTS.0.ART_NAME}` or `{SNG_CONTRIBUTORS.main_artist.0}`.

Fallback placeholders are also supported. Use `|` inside a placeholder to use the first non-empty value:

```text
{ALB_TITLE|TITLE}              Album title, or playlist title if album title is empty
{TRACK_POSITION|TRACK_NUMBER}  Track position, or track number if position is empty
```

Example:

```json
{
  "saveLayout": {
    "album": "Music/{ART_NAME|ARTISTS.0.ART_NAME}/{RELEASE_YEAR} - {ALB_TITLE}/{DISK_FOLDER}/{TRACK_POSITION|TRACK_NUMBER} - {SNG_TITLE}",
    "playlist": "Playlist/{TITLE}/{SNG_TITLE}"
  }
}
```

### `playlist.resolveFullPath`

When `true`, generated `.m3u8` playlists contain absolute file paths:

```text
/home/sayem/Playlist/My Playlist/01 - A song.mp3
```

When `false`, playlist entries are relative to the playlist file location.

### `trackNumber`

When `true`, GoFi prefixes saved tracks with track position, such as `01 - Title` or `02 - Title`. When `false`, the number prefix is omitted unless the layout explicitly uses a track-number placeholder.

### `fallbackTrack`

When `true`, GoFi can download Deezer's fallback track when the requested track was moved or replaced and the fallback is available. This matches how Deezer handles some moved or deleted songs.

### `fallbackQuality`

When `true`, GoFi falls back to a lower available quality when the requested quality is unavailable. For example, FLAC may fall back to MP3 320, or MP3 320 may fall back to MP3 128. Set this to `false` if you want unavailable qualities to be skipped instead.

### `coverSize`

Album cover size used for metadata tagging and saved cover files. Valid values are between `50` and `1800`. The web UI shows common presets: `56`, `250`, `500`, `1000`, `1200`, `1400`, `1500`, and `1800`. If the config contains another valid value, such as `1234`, the web UI keeps it as a custom option.

```text
coverSize.128     Cover size for MP3 128 downloads
coverSize.320     Cover size for MP3 320 downloads
coverSize.flac    Cover size for FLAC downloads
```

### `cover.mode`

Controls how album artwork is handled. The default is `embed`, so existing configs keep the previous behavior.

```text
embed    Embed album art in each downloaded track
file     Save the configured cover image file next to tracks, without embedding artwork
both     Embed album art and save the configured cover image file
none     Do not embed or save album artwork
```

When saving a cover file, GoFi only creates it in a folder where all selected tracks share the same album cover. This avoids writing a misleading cover file into a mixed playlist folder.

### `cover.fileName`

File name used when `cover.mode` is `file` or `both`. The default is `cover.jpg`.

Deezer returns JPEG cover bytes, so GoFi keeps the file extension as `.jpg` or `.jpeg`. Path-like names are reduced to a safe file name.

### `cookies.arl`

Saved Deezer ARL cookie. GoFi also supports `DEEZER_ARL`. When both are present, the environment variable takes priority over `cookies.arl`.

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

Supported converter inputs include Deezer, Spotify, Tidal, YouTube, ISRC, and UPC helpers.

Download a tagged track to a file:

```go
path, err := download.DownloadTrack(context.Background(), download.DownloadTrackOptions{
	SngID:     "3135556",
	Quality:   3, // 1 = MP3 128, 3 = MP3 320, 9 = FLAC
	CoverSize: 500,
	CoverMode: "embed", // optional: embed, file, both, or none
	CoverName: "cover.jpg",
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
