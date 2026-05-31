package dfi

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/types"
	"github.com/d-fi/GoFi/utils"
)

const Version = "2.3.1-go"

type options struct {
	quality         string
	output          string
	url             string
	inputFile       string
	concurrency     int
	setARL          string
	headless        bool
	configFile      string
	resolveFullPath bool
	createPlaylist  bool
	update          bool
}

// Run starts the d-fi compatible CLI.
func Run(ctx context.Context, args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	printBanner()
	if _, err := CleanupStaleDownloadTemps(".", time.Hour); err != nil {
		fmt.Fprintln(os.Stderr, warn("Unable to clean stale resumable download files: "+err.Error()))
	}

	if opts.update {
		fmt.Println(info("Binary self-update is not available for the Go build yet."))
		return nil
	}

	if opts.headless && opts.quality == "" {
		return fmt.Errorf("missing parameters --quality\n%s", note("Quality must be provided with headless mode"))
	}
	if opts.headless && opts.url == "" && opts.inputFile == "" {
		return fmt.Errorf("missing parameters --url\n%s", note("URL must be provided with headless mode"))
	}

	cfg := LoadConfig(opts.configFile)
	if cfg.UserConfigLocation != "" {
		fmt.Println(info("Config loaded --> " + cfg.UserConfigLocation))
	}

	if opts.setARL != "" {
		if err := cfg.Set("cookies.arl", opts.setARL); err != nil {
			return err
		}
		fmt.Println(info("cookies.arl set to --> " + opts.setARL))
		fmt.Println(note(opts.configFile))
		return nil
	}

	fmt.Println(pending("Initializing session..."))
	arl := resolveARL(cfg)
	if arl == "" {
		return fmt.Errorf("missing Deezer ARL. Set DEEZER_ARL or run d-fi --set-arl <arl>")
	}
	fmt.Println(pending("Verifying session..."))
	if _, err := request.InitDeezerAPI(arl); err != nil {
		return err
	}
	user, err := api.GetUser()
	if err != nil {
		return err
	}
	fmt.Println(success("Logged in as " + user.BlogName))

	if opts.inputFile != "" {
		data, err := os.ReadFile(opts.inputFile)
		if err != nil {
			return err
		}
		for line := range strings.SplitSeq(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || !LooksLikeURL(line) {
				continue
			}
			fmt.Println(info("Starting download: " + line))
			if err := startDownload(ctx, cfg, opts, line, true); err != nil {
				fmt.Fprintln(os.Stderr, failure(err.Error()))
			}
		}
		return nil
	}

	return startDownload(ctx, cfg, opts, opts.url, false)
}

func resolveARL(cfg Config) string {
	arl := strings.TrimSpace(os.Getenv("DEEZER_ARL"))
	if arl != "" {
		return arl
	}
	return strings.TrimSpace(cfg.Cookies.ARL)
}

func parseOptions(args []string) (options, error) {
	var opts options
	fs := flag.NewFlagSet("d-fi", flag.ContinueOnError)
	fs.Usage = func() {
		printUsage(fs.Output())
	}
	fs.StringVar(&opts.quality, "quality", "", "The quality of the files to download: 128/320/flac")
	fs.StringVar(&opts.quality, "q", "", "The quality of the files to download: 128/320/flac")
	fs.StringVar(&opts.output, "output", "", "Output filename template")
	fs.StringVar(&opts.output, "o", "", "Output filename template")
	fs.StringVar(&opts.url, "url", "", "Deezer album/artist/playlist/track url")
	fs.StringVar(&opts.url, "u", "", "Deezer album/artist/playlist/track url")
	fs.StringVar(&opts.inputFile, "input-file", "", "Downloads all urls listed in text file")
	fs.StringVar(&opts.inputFile, "i", "", "Downloads all urls listed in text file")
	fs.IntVar(&opts.concurrency, "concurrency", 0, "Download concurrency for album, artists and playlist")
	fs.IntVar(&opts.concurrency, "c", 0, "Download concurrency for album, artists and playlist")
	fs.StringVar(&opts.setARL, "set-arl", "", "Set arl cookie")
	fs.StringVar(&opts.setARL, "a", "", "Set arl cookie")
	fs.BoolVar(&opts.headless, "headless", false, "Run in headless mode for scripting automation")
	fs.BoolVar(&opts.headless, "d", false, "Run in headless mode for scripting automation")
	fs.StringVar(&opts.configFile, "config-file", "d-fi.config.json", "Custom location to your config file")
	fs.StringVar(&opts.configFile, "conf", "d-fi.config.json", "Custom location to your config file")
	fs.BoolVar(&opts.resolveFullPath, "resolve-full-path", false, "Use absolute path for playlists")
	fs.BoolVar(&opts.resolveFullPath, "rfp", false, "Use absolute path for playlists")
	fs.BoolVar(&opts.createPlaylist, "create-playlist", false, "Force create a playlist file for non playlists")
	fs.BoolVar(&opts.createPlaylist, "cp", false, "Force create a playlist file for non playlists")
	fs.BoolVar(&opts.update, "update", false, "Update this program to latest version")
	fs.BoolVar(&opts.update, "U", false, "Update this program to latest version")
	if err := fs.Parse(args); err != nil {
		return opts, err
	}
	if opts.url == "" && fs.NArg() > 0 {
		opts.url = fs.Arg(0)
	}
	return opts, nil
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage of d-fi:")
	fmt.Fprintln(w, "  -q, --quality <quality>       The quality of the files to download: 128/320/flac")
	fmt.Fprintln(w, "  -o, --output <template>       Output filename template")
	fmt.Fprintln(w, "  -u, --url <url>               Deezer album/artist/playlist/track url")
	fmt.Fprintln(w, "  -i, --input-file <file>       Downloads all urls listed in text file")
	fmt.Fprintln(w, "  -c, --concurrency <number>    Download concurrency for album, artists and playlist")
	fmt.Fprintln(w, "  -a, --set-arl <string>        Set arl cookie")
	fmt.Fprintln(w, "  -d, --headless                Run in headless mode for scripting automation")
	fmt.Fprintln(w, "  -conf, --config-file <file>   Custom location to your config file")
	fmt.Fprintln(w, "  -rfp, --resolve-full-path     Use absolute path for playlists")
	fmt.Fprintln(w, "  -cp, --create-playlist        Force create a playlist file for non playlists")
	fmt.Fprintln(w, "  -U, --update                  Update this program to latest version")
	fmt.Fprintln(w, "  -h, --help                    Shows this help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  web                           Start the web UI")
}

func printBanner() {
	fmt.Println("             ♥ d-fi - " + Version + " ♥")
	fmt.Println(" ──────────────────────────────────────────────")
	fmt.Println(" │ github   https://github.com/d-fi           │")
	fmt.Println(" │ telegram https://t.me/dFiCommunity         │")
	fmt.Println(" ──────────────────────────────────────────────")
}

func startDownload(ctx context.Context, cfg Config, opts options, rawURL string, skipPrompt bool) error {
	reader := bufio.NewReader(os.Stdin)
	if opts.quality == "" {
		quality, err := promptQuality(reader)
		if err != nil {
			return err
		}
		opts.quality = quality
	} else {
		_, _, quality, err := ParseQualityStrict(opts.quality)
		if err != nil {
			return err
		}
		opts.quality = quality
	}
	if rawURL == "" {
		fmt.Print("Enter URL or search: ")
		value, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		rawURL = strings.TrimSpace(value)
	}

	data, err := resolveInput(rawURL, opts.headless, reader)
	if err != nil {
		return err
	}
	if !opts.headless && len(data.Tracks) > 1 {
		data.Tracks, err = promptTracks(reader, data.Tracks)
		if err != nil {
			return err
		}
	}

	if len(data.Tracks) == 0 {
		fmt.Println(info("No items to download!"))
		return nil
	}

	fmt.Println(info(fmt.Sprintf("Proceeding to download %d tracks. Be patient.", len(data.Tracks))))
	if data.LinkType == "playlist" {
		data.Tracks = dedupePlaylistTracks(data.Tracks)
	}

	resolveFullPath := opts.resolveFullPath || cfg.Playlist.ResolveFullPath
	concurrency := opts.concurrency
	if concurrency <= 0 {
		concurrency = cfg.Concurrency
	}
	if concurrency <= 0 {
		concurrency = 1
	}

	pathTemplate := opts.output
	if pathTemplate == "" {
		pathTemplate = cfg.Layout(data.LinkType)
	}

	savedFiles := downloadAll(ctx, data, cfg, opts, pathTemplate, concurrency)
	if len(savedFiles) > 0 {
		fmt.Println(info("Saved in " + strings.Join(uniqueDirs(savedFiles), ", ")))
	}

	if (opts.createPlaylist || data.LinkType == "playlist") && os.Getenv("SIMULATE") == "" && len(savedFiles) > 1 {
		if _, err := WritePlaylistFile(data.LinkInfo, savedFiles, resolveFullPath); err != nil {
			return err
		}
	}

	if !opts.headless && !skipPrompt {
		return startDownload(ctx, cfg, opts, "", skipPrompt)
	}
	return nil
}

func resolveInput(rawURL string, headless bool, reader *bufio.Reader) (ResolvedInput, error) {
	if !LooksLikeURL(rawURL) {
		if headless {
			return ResolvedInput{}, fmt.Errorf("please provide a valid URL. Unknown URL: %s", rawURL)
		}
		return resolveSearch(rawURL, reader)
	}
	if strings.Contains(rawURL, "playlist") || strings.Contains(rawURL, "artist") {
		fmt.Println(info("Fetching data. Please hold on."))
	}
	data, err := ParseResolvedURL(rawURL)
	if err != nil {
		return ResolvedInput{}, err
	}
	return data, nil
}

func resolveSearch(query string, reader *bufio.Reader) (ResolvedInput, error) {
	switch {
	case strings.HasPrefix(query, "artist:"):
		search, err := api.SearchMusic(strings.TrimPrefix(query, "artist:"), SearchOptionLimit, "ARTIST")
		if err != nil {
			return ResolvedInput{}, err
		}
		index, err := promptChoice(reader, fmt.Sprintf("Select one artist. (found %d artists)", len(search.ARTIST.Data)), len(search.ARTIST.Data), func(i int) string {
			item := search.ARTIST.Data[i]
			return fmt.Sprintf("%s - %d fans", item.ART_NAME, item.NB_FAN)
		})
		if err != nil {
			return ResolvedInput{}, err
		}
		fmt.Println(info("Fetching data. Please hold on."))
		return resolveInput("https://deezer.com/us/artist/"+search.ARTIST.Data[index].ART_ID, false, reader)
	case strings.HasPrefix(query, "album:"):
		search, err := api.SearchMusic(strings.TrimPrefix(query, "album:"), SearchOptionLimit, "ALBUM")
		if err != nil {
			return ResolvedInput{}, err
		}
		index, err := promptChoice(reader, fmt.Sprintf("Select one album. (found %d albums)", len(search.ALBUM.Data)), len(search.ALBUM.Data), func(i int) string {
			item := search.ALBUM.Data[i]
			return fmt.Sprintf("%s - by %s, %s tracks", item.ALB_TITLE, item.ART_NAME, item.NUMBER_TRACK)
		})
		if err != nil {
			return ResolvedInput{}, err
		}
		return resolveInput("https://deezer.com/us/album/"+search.ALBUM.Data[index].ALB_ID, false, reader)
	case strings.HasPrefix(query, "playlist:"):
		search, err := api.SearchMusic(strings.TrimPrefix(query, "playlist:"), SearchOptionLimit, "PLAYLIST")
		if err != nil {
			return ResolvedInput{}, err
		}
		index, err := promptChoice(reader, fmt.Sprintf("Select one playlist. (found %d playlists)", len(search.PLAYLIST.Data)), len(search.PLAYLIST.Data), func(i int) string {
			item := search.PLAYLIST.Data[i]
			return fmt.Sprintf("%s - by %s, %d tracks", item.Title, item.ParentUsername, item.NbSong)
		})
		if err != nil {
			return ResolvedInput{}, err
		}
		return resolveInput("https://deezer.com/us/playlist/"+search.PLAYLIST.Data[index].PlaylistID, false, reader)
	default:
		data, err := ResolveTrackSearch(query)
		if err != nil {
			return ResolvedInput{}, err
		}
		return data, nil
	}
}

func promptQuality(reader *bufio.Reader) (string, error) {
	for {
		fmt.Println("Select music quality:")
		fmt.Println("1) MP3  - 128 kbps")
		fmt.Println("2) MP3  - 320 kbps")
		fmt.Println("3) FLAC - 1411 kbps")
		fmt.Print("> ")
		value, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		switch strings.TrimSpace(value) {
		case "1":
			return "128", nil
		case "2":
			return "320", nil
		case "3":
			return "flac", nil
		default:
			fmt.Println("Invalid quality. Choose 1, 2, or 3.")
		}
	}
}

func promptChoice(reader *bufio.Reader, message string, count int, describe func(int) string) (int, error) {
	if count == 0 {
		return 0, fmt.Errorf("no items found")
	}
	fmt.Println(message)
	for i := range count {
		fmt.Printf("%d) %s\n", i+1, describe(i))
	}
	fmt.Print("> ")
	value, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	index, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || index < 1 || index > count {
		return 0, fmt.Errorf("invalid selection")
	}
	return index - 1, nil
}

func promptTracks(reader *bufio.Reader, tracks []types.TrackType) ([]types.TrackType, error) {
	fmt.Printf("Select songs to download. Total of %d tracks.\n", len(tracks))
	for i, track := range tracks {
		fmt.Printf("%d) %s - Artist: %s, Album: %s, Duration: %s\n", i+1, track.SNG_TITLE, track.ART_NAME, track.ALB_TITLE, formatSecondsReadable(AsInt(track.DURATION)))
	}
	fmt.Print("Comma separated numbers, ranges, or blank for all: ")
	value, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return tracks, nil
	}
	selected := []types.TrackType{}
	seen := map[int]bool{}
	for part := range strings.SplitSeq(value, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			edges := strings.SplitN(part, "-", 2)
			start, _ := strconv.Atoi(strings.TrimSpace(edges[0]))
			end, _ := strconv.Atoi(strings.TrimSpace(edges[1]))
			for i := start; i <= end; i++ {
				addTrackSelection(&selected, seen, tracks, i)
			}
			continue
		}
		index, _ := strconv.Atoi(part)
		addTrackSelection(&selected, seen, tracks, index)
	}
	return selected, nil
}

func addTrackSelection(selected *[]types.TrackType, seen map[int]bool, tracks []types.TrackType, index int) {
	if index < 1 || index > len(tracks) || seen[index] {
		return
	}
	seen[index] = true
	*selected = append(*selected, tracks[index-1])
}

func dedupePlaylistTracks(tracks []types.TrackType) []types.TrackType {
	filtered, duplicates := dedupePlaylistTrackList(tracks)
	if duplicates > 0 {
		fmt.Println(warn(fmt.Sprintf("Removed %d duplicate %s.", duplicates, plural("track", duplicates))))
	}
	return filtered
}

func DedupePlaylistTracks(tracks []types.TrackType) []types.TrackType {
	filtered, _ := dedupePlaylistTrackList(tracks)
	return filtered
}

func AppendTrackVersionsToTitles(tracks []types.TrackType) []types.TrackType {
	for i := range tracks {
		version := tracks[i].VERSION
		if version != nil && *version != "" && !strings.Contains(tracks[i].SNG_TITLE, *version) {
			tracks[i].SNG_TITLE += " " + *version
		}
	}
	return tracks
}

func dedupePlaylistTrackList(tracks []types.TrackType) ([]types.TrackType, int) {
	seen := map[string]bool{}
	filtered := make([]types.TrackType, 0, len(tracks))
	duplicates := 0
	for _, track := range tracks {
		if seen[track.SNG_ID] {
			duplicates++
			continue
		}
		seen[track.SNG_ID] = true
		filtered = append(filtered, track)
	}
	if duplicates > 0 {
		sort.SliceStable(filtered, func(i, j int) bool {
			return trackPosition(filtered[i]) < trackPosition(filtered[j])
		})
		for i := range filtered {
			position := i + 1
			filtered[i].TRACK_POSITION = &position
		}
	}
	return filtered, duplicates
}

func trackPosition(track types.TrackType) int {
	if track.TRACK_POSITION != nil {
		return *track.TRACK_POSITION
	}
	return 0
}

func plural(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

func downloadAll(ctx context.Context, data ResolvedInput, cfg Config, opts options, pathTemplate string, concurrency int) []string {
	type job struct {
		index int
		track types.TrackType
	}

	jobs := make(chan job)
	var wg sync.WaitGroup
	var mu sync.Mutex
	savedFiles := []string{}
	workerCount := min(len(data.Tracks), concurrency)
	coverPolicy := CoverFilePolicy(data.Tracks, data.LinkInfo, pathTemplate, cfg.TrackNumber)

	for range workerCount {
		wg.Go(func() {
			for item := range jobs {
				savedPath, err := downloadTrack(ctx, DownloadTrackOptions{
					Track:           item.track,
					Quality:         opts.quality,
					Info:            data.LinkInfo,
					CoverSizes:      cfg.CoverSize,
					CoverMode:       cfg.Cover.Mode,
					CoverFileName:   cfg.Cover.FileName,
					CoverFilePolicy: coverPolicy,
					Path:            pathTemplate,
					TotalTracks:     len(data.Tracks),
					TrackNumber:     cfg.TrackNumber,
					FallbackTrack:   cfg.FallbackTrack,
					FallbackQuality: cfg.FallbackQuality,
					Message:         fmt.Sprintf("(%d/%d)", item.index, len(data.Tracks)),
				})
				if err != nil {
					fmt.Fprintln(os.Stderr, failure(item.track.SNG_TITLE))
					fmt.Fprintln(os.Stderr, note(err.Error()))
					continue
				}
				if savedPath == "" {
					continue
				}
				mu.Lock()
				savedFiles = append(savedFiles, savedPath)
				mu.Unlock()
			}
		})
	}

	for index, track := range data.Tracks {
		jobs <- job{index: index, track: track}
	}
	close(jobs)
	wg.Wait()
	return savedFiles
}

func WritePlaylistFile(info any, savedFiles []string, resolveFullPath bool) (string, error) {
	playlistDir := commonPath(uniqueDirs(savedFiles))
	if playlistDir == "" {
		playlistDir = "."
	}
	name := playlistName(info)
	if name == "" {
		name = "playlist"
	}

	entries := append([]string(nil), savedFiles...)
	if resolveFullPath {
		for i, file := range entries {
			if abs, err := filepath.Abs(file); err == nil {
				entries[i] = abs
			}
		}
	} else {
		resolvedDir, _ := filepath.Abs(playlistDir)
		for i, file := range entries {
			abs, err := filepath.Abs(file)
			if err != nil {
				continue
			}
			rel, err := filepath.Rel(resolvedDir, abs)
			if err == nil {
				entries[i] = rel
			}
		}
	}
	sort.Strings(entries)
	content := "#EXTM3U\n" + strings.Join(entries, "\n")
	path := filepath.Join(playlistDir, utils.SanitizeFileName(name)+".m3u8")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func playlistName(info any) string {
	data := utils.StructMap(info)
	for _, key := range []string{"TITLE", "ALB_TITLE"} {
		if value, ok := data[key]; ok {
			return fmt.Sprintf("%v", value)
		}
	}
	return ""
}
