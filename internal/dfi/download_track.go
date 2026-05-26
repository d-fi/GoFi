package dfi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/d-fi/GoFi/decrypt"
	"github.com/d-fi/GoFi/download"
	"github.com/d-fi/GoFi/metadata"
	"github.com/d-fi/GoFi/types"
)

type DownloadTrackOptions struct {
	Track             types.TrackType
	Quality           any
	Info              any
	CoverSizes        CoverSizes
	Path              string
	TotalTracks       int
	TrackNumber       bool
	FallbackTrack     bool
	FallbackQuality   bool
	IsFallback        bool
	IsQualityFallback bool
	Message           string
}

func downloadTrack(ctx context.Context, options DownloadTrackOptions) (string, error) {
	track := options.Track
	terminalStatus.Println(pending(track.SNG_TITLE + " by " + track.ART_NAME + " from " + track.ALB_TITLE))

	quality, ext, label := parseQuality(options.Quality)
	coverSize := coverSizeForQuality(options.CoverSizes, label)
	if os.Getenv("SIMULATE") != "" {
		coverSize = 56
	}

	savePath := saveLayout(track, options.Info, options.Path, options.TrackNumber, options.TotalTracks) + ext
	if _, err := os.Stat(savePath); err == nil {
		terminalStatus.Println(info(fmt.Sprintf("Skipped %q, track already exists.", track.SNG_TITLE)))
		terminalStatus.Println(note(savePath))
		return savePath, nil
	}

	trackData, err := download.GetTrackDownloadUrl(track, quality)
	if err != nil {
		var geoBlocked *download.GeoBlocked
		if !errors.As(err, &geoBlocked) || track.FALLBACK == nil {
			return "", err
		}
	}

	if trackData == nil {
		if options.FallbackTrack && track.FALLBACK != nil && !options.IsFallback && track.ART_ID == track.FALLBACK.ART_ID {
			fallback := track
			fallback.SongType = *track.FALLBACK
			fallback.FALLBACK = nil
			fallback.TRACK_POSITION = track.TRACK_POSITION
			options.Track = fallback
			options.FallbackTrack = false
			options.IsFallback = true
			return downloadTrack(ctx, options)
		}
		if options.FallbackQuality && quality != 1 {
			if quality == 9 {
				options.Quality = 3
			} else {
				options.Quality = 1
			}
			options.IsQualityFallback = true
			return downloadTrack(ctx, options)
		}
		terminalStatus.Println(warn(fmt.Sprintf("Skipped %q, track not available.", track.SNG_TITLE)))
		return "", nil
	}

	tmpFile := fmt.Sprintf("d-fi_%d_%s_%s", quality, track.SNG_ID, track.MD5_ORIGIN)
	if os.Getenv("SIMULATE") != "" {
		tmpFile = fmt.Sprintf("d-fi_%d_%s_simulate", quality, track.SNG_ID)
	}
	if err := downloadToTemp(ctx, trackData, tmpFile, track, options.Message); err != nil {
		return "", err
	}
	defer os.Remove(tmpFile)

	raw, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", err
	}
	if trackData.IsEncrypted {
		terminalStatus.Update(pending("Decrypting " + track.SNG_TITLE + " by " + track.ART_NAME))
		raw = decrypt.DecryptDownload(raw, track.SNG_ID)
	}

	terminalStatus.Update(pending("Tagging " + track.SNG_TITLE + " by " + track.ART_NAME))
	tagged, err := metadata.AddTrackTags(raw, track, coverSize)
	if err != nil {
		return "", err
	}

	terminalStatus.Update(pending("Saving " + track.SNG_TITLE + " by " + track.ART_NAME))
	if os.Getenv("SIMULATE") == "" {
		if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(savePath, tagged, 0644); err != nil {
			return "", err
		}
	}

	prefix := ""
	if options.IsFallback {
		prefix = "[Fallback] "
	}
	terminalStatus.Done(success(prefix + track.SNG_TITLE + " by " + track.ART_NAME))
	if options.IsQualityFallback {
		if quality == 3 {
			terminalStatus.Println(note("Used 320kbps as other formats were unavailable"))
		} else {
			terminalStatus.Println(note("Used 128kbps as other formats were unavailable"))
		}
	}
	return savePath, nil
}

func downloadToTemp(ctx context.Context, trackData *download.TrackDownloadUrl, tmpFile string, track types.TrackType, message string) error {
	var downloaded int64
	resuming := false
	headers := http.Header{}
	if os.Getenv("SIMULATE") != "" {
		headers.Set("Range", "bytes=0-1023")
	} else if stat, err := os.Stat(tmpFile); err == nil {
		downloaded = stat.Size()
		if downloaded > 0 {
			resuming = true
			headers.Set("Range", fmt.Sprintf("bytes=%d-", downloaded))
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, trackData.TrackUrl, nil)
	if err != nil {
		return err
	}
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	if resuming && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		if err := os.Remove(tmpFile); err != nil && !os.IsNotExist(err) {
			return err
		}
		downloaded = 0
		return downloadToTemp(ctx, trackData, tmpFile, track, message)
	}

	out, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	total := trackData.FileSize
	bar := progressBar(total, 40)
	humanSizeTotal := float64(total) / 1024 / 1024
	transferred := downloaded
	lastPrinted := transferred
	lastLogged := transferred
	lastProgressUpdate := time.Now().Add(-500 * time.Millisecond)
	buffer := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buffer)
		if n > 0 {
			written, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}
			transferred += int64(written)
			now := time.Now()
			completed := transferred >= total
			if (transferred-lastPrinted > 50000 && now.Sub(lastProgressUpdate) >= 500*time.Millisecond) || completed {
				lastPrinted = transferred
				lastProgressUpdate = now
				progress := info(fmt.Sprintf("Downloading %s %s  %s | %.2fMiB", track.SNG_TITLE, message, bar(transferred), humanSizeTotal))
				if terminalStatus.interactive {
					terminalStatus.Update(progress)
				} else if transferred-lastLogged > 5*1024*1024 || transferred >= total {
					lastLogged = transferred
					terminalStatus.Println(progress)
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			if !strings.Contains(readErr.Error(), "context canceled") {
				_ = os.Remove(tmpFile)
			}
			return readErr
		}
	}
	return nil
}
