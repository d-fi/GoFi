package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// SimpleProgress provides a simple progress display without external dependencies
type SimpleProgress struct {
	mu          sync.Mutex
	description string
	total       int64
	current     int64
	startTime   time.Time
	lastUpdate  time.Time
}

// NewSimpleProgress creates a new simple progress tracker
func NewSimpleProgress(description string, total int64) *SimpleProgress {
	return &SimpleProgress{
		description: description,
		total:       total,
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
	}
}

// Update updates the progress
func (sp *SimpleProgress) Update(current int64) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	
	sp.current = current
	now := time.Now()
	
	// Only update display every 100ms to avoid flickering
	if now.Sub(sp.lastUpdate) < 100*time.Millisecond && sp.current < sp.total {
		return
	}
	sp.lastUpdate = now
	
	sp.render()
}

// Finish completes the progress
func (sp *SimpleProgress) Finish() {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	
	sp.current = sp.total
	sp.render()
	fmt.Print(" ")
	SuccessColor.Printf("%s\n", IconSuccess)
}

// render displays the progress bar
func (sp *SimpleProgress) render() {
	if sp.total <= 0 {
		// Unknown total, just show a spinner
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		elapsed := time.Since(sp.startTime)
		idx := int(elapsed.Milliseconds()/100) % len(spinner)
		// Clear the entire line first, then print
		fmt.Printf("\r\033[K%s %s %s", IconDownload, sp.description, spinner[idx])
		return
	}
	
	// Calculate progress percentage
	percent := int(float64(sp.current) * 100 / float64(sp.total))
	if percent > 100 {
		percent = 100
	}
	
	// Format sizes
	currentMB := float64(sp.current) / 1024 / 1024
	totalMB := float64(sp.total) / 1024 / 1024
	
	// Calculate speed
	elapsed := time.Since(sp.startTime).Seconds()
	if elapsed < 0.1 {
		elapsed = 0.1
	}
	speedMBps := currentMB / elapsed
	
	// Calculate ETA
	var eta string
	if speedMBps > 0 && sp.current < sp.total {
		remainingMB := totalMB - currentMB
		remainingSec := int(remainingMB / speedMBps)
		if remainingSec > 0 {
			eta = fmt.Sprintf(" - %ds remaining", remainingSec)
		}
	}
	
	// Build progress bar
	barWidth := 20
	filled := int(float64(barWidth) * float64(percent) / 100)
	if filled > barWidth {
		filled = barWidth
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	
	// Clear the entire line first, then print progress
	// \r moves cursor to beginning, \033[K clears from cursor to end of line
	fmt.Printf("\r\033[K%s %s [%s] %d%% (%.1f/%.1f MB, %.1f MB/s)%s", 
		IconDownload, 
		sp.description,
		bar,
		percent,
		currentMB,
		totalMB,
		speedMBps,
		eta)
}

// Clear clears the progress line
func (sp *SimpleProgress) Clear() {
	fmt.Printf("\r\033[K")
}