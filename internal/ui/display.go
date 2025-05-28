package ui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// UI color definitions
var (
	InfoColor    = color.New(color.FgCyan)
	SuccessColor = color.New(color.FgGreen)
	ErrorColor   = color.New(color.FgRed)
	WarningColor = color.New(color.FgYellow)
	HeaderColor  = color.New(color.FgMagenta, color.Bold)
	BoldColor    = color.New(color.Bold)
	DimColor     = color.New(color.Faint)
)

// Icons for different states
const (
	IconInfo     = "ℹ"
	IconSuccess  = "✓"
	IconError    = "✗"
	IconWarning  = "⚠"
	IconMusic    = "♫"
	IconDownload = "↓"
	IconSearch   = "🔍"
	IconFolder   = "📁"
)

// DisplayManager handles all UI output
type DisplayManager struct {
	mu              sync.Mutex
	activeDownloads map[string]*SimpleProgress
}

// NewDisplayManager creates a new display manager
func NewDisplayManager() *DisplayManager {
	return &DisplayManager{
		activeDownloads: make(map[string]*SimpleProgress),
	}
}

// PrintHeader prints a styled header
func (dm *DisplayManager) PrintHeader(text string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	fmt.Println()
	HeaderColor.Printf("═══ %s %s ═══\n", IconMusic, text)
	fmt.Println()
}

// PrintInfo prints an info message
func (dm *DisplayManager) PrintInfo(format string, args ...interface{}) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	InfoColor.Printf("%s ", IconInfo)
	fmt.Printf(format+"\n", args...)
}

// PrintSuccess prints a success message
func (dm *DisplayManager) PrintSuccess(format string, args ...interface{}) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	SuccessColor.Printf("%s ", IconSuccess)
	fmt.Printf(format+"\n", args...)
}

// PrintError prints an error message
func (dm *DisplayManager) PrintError(format string, args ...interface{}) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	ErrorColor.Printf("%s ", IconError)
	fmt.Printf(format+"\n", args...)
}

// PrintWarning prints a warning message
func (dm *DisplayManager) PrintWarning(format string, args ...interface{}) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	WarningColor.Printf("%s ", IconWarning)
	fmt.Printf(format+"\n", args...)
}

// PrintSearching prints a searching message
func (dm *DisplayManager) PrintSearching(service string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	InfoColor.Printf("%s Searching %s... ", IconSearch, service)
}

// PrintSearchResult prints the result of a search
func (dm *DisplayManager) PrintSearchResult(success bool) {
	if success {
		SuccessColor.Printf("%s\n", IconSuccess)
	} else {
		ErrorColor.Printf("%s\n", IconError)
	}
}

// PrintTrackInfo prints formatted track information
func (dm *DisplayManager) PrintTrackInfo(title, artist, album string, quality int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	fmt.Println()
	BoldColor.Println("Track Details:")
	fmt.Printf("  %s Title:   %s\n", IconMusic, title)
	fmt.Printf("  %s Artist:  %s\n", IconMusic, artist)
	fmt.Printf("  %s Album:   %s\n", IconMusic, album)
	fmt.Printf("  %s Quality: ", IconMusic)
	
	switch quality {
	case 9:
		SuccessColor.Printf("FLAC (Lossless)\n")
	case 3:
		InfoColor.Printf("MP3 320kbps\n")
	case 1:
		WarningColor.Printf("MP3 128kbps\n")
	default:
		fmt.Printf("Quality %d\n", quality)
	}
	fmt.Println()
}

// PrintAlbumInfo prints formatted album information
func (dm *DisplayManager) PrintAlbumInfo(title, artist string, trackCount int, quality int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	fmt.Println()
	BoldColor.Println("Album Details:")
	fmt.Printf("  %s Title:   %s\n", IconMusic, title)
	fmt.Printf("  %s Artist:  %s\n", IconMusic, artist)
	fmt.Printf("  %s Tracks:  %d\n", IconMusic, trackCount)
	fmt.Printf("  %s Quality: ", IconMusic)
	
	switch quality {
	case 9:
		SuccessColor.Printf("FLAC (Lossless)\n")
	case 3:
		InfoColor.Printf("MP3 320kbps\n")
	case 1:
		WarningColor.Printf("MP3 128kbps\n")
	default:
		fmt.Printf("Quality %d\n", quality)
	}
	fmt.Println()
}

// PrintPlaylistInfo prints formatted playlist information
func (dm *DisplayManager) PrintPlaylistInfo(title string, owner string, trackCount int, quality int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	fmt.Println()
	BoldColor.Println("Playlist Details:")
	fmt.Printf("  %s Title:   %s\n", IconMusic, title)
	if owner != "" {
		fmt.Printf("  %s Owner:   %s\n", IconMusic, owner)
	}
	fmt.Printf("  %s Tracks:  %d\n", IconMusic, trackCount)
	fmt.Printf("  %s Quality: ", IconMusic)
	
	switch quality {
	case 9:
		SuccessColor.Printf("FLAC (Lossless)\n")
	case 3:
		InfoColor.Printf("MP3 320kbps\n")
	case 1:
		WarningColor.Printf("MP3 128kbps\n")
	default:
		fmt.Printf("Quality %d\n", quality)
	}
	fmt.Println()
}

// StartProgress creates and starts a new progress bar
func (dm *DisplayManager) StartProgress(id string, total int64, description string) *SimpleProgress {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	// Truncate description if too long
	maxLen := 40
	if len(description) > maxLen {
		description = description[:maxLen-3] + "..."
	}
	
	progress := NewSimpleProgress(description, total)
	dm.activeDownloads[id] = progress
	return progress
}

// UpdateProgress updates an existing progress bar
func (dm *DisplayManager) UpdateProgress(id string, current int64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	if progress, exists := dm.activeDownloads[id]; exists {
		progress.Update(current)
	}
}

// FinishProgress marks a progress bar as complete
func (dm *DisplayManager) FinishProgress(id string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	if progress, exists := dm.activeDownloads[id]; exists {
		progress.Finish()
		delete(dm.activeDownloads, id)
	}
}


// PrintDownloadSummary prints a summary of downloads
func (dm *DisplayManager) PrintDownloadSummary(succeeded, failed, total int) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	BoldColor.Println("Download Summary")
	fmt.Println(strings.Repeat("─", 50))
	
	if succeeded > 0 {
		SuccessColor.Printf("%s Succeeded: %d\n", IconSuccess, succeeded)
	}
	if failed > 0 {
		ErrorColor.Printf("%s Failed:    %d\n", IconError, failed)
	}
	fmt.Printf("  Total:     %d\n", total)
	
	if failed == 0 {
		fmt.Println()
		SuccessColor.Printf("%s All downloads completed successfully!\n", IconSuccess)
	} else if succeeded == 0 {
		fmt.Println()
		ErrorColor.Printf("%s All downloads failed.\n", IconError)
	} else {
		fmt.Println()
		WarningColor.Printf("%s Some downloads failed. Check the errors above.\n", IconWarning)
	}
	fmt.Println(strings.Repeat("─", 50))
}

// PrintFileExists prints a message when a file already exists
func (dm *DisplayManager) PrintFileExists(filename string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	DimColor.Printf("%s File already exists: %s\n", IconSuccess, filename)
}

// PrintSavePath prints where a file was saved
func (dm *DisplayManager) PrintSavePath(path string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	fmt.Printf("%s Saved to: ", IconFolder)
	InfoColor.Println(path)
}