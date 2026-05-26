package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d-fi/GoFi/api"
	"github.com/d-fi/GoFi/converter"
	"github.com/d-fi/GoFi/internal/dfi"
	"github.com/d-fi/GoFi/request"
	"github.com/d-fi/GoFi/types"
)

type Options struct {
	Addr       string
	ConfigPath string
}

type Server struct {
	cfgPath string
	mux     *http.ServeMux

	mu      sync.Mutex
	cfg     dfi.Config
	session sessionState
	jobs    map[int64]*downloadJob
	nextID  int64
}

type sessionState struct {
	Ready    bool   `json:"ready"`
	UserName string `json:"userName,omitempty"`
	Error    string `json:"error,omitempty"`
}

type downloadJob struct {
	ID          int64     `json:"id"`
	Source      string    `json:"source"`
	Quality     string    `json:"quality"`
	SaveToDir   string    `json:"saveToDir"`
	Status      string    `json:"status"`
	TotalTracks int       `json:"totalTracks"`
	DoneTracks  int       `json:"doneTracks"`
	Progress    float64   `json:"progress"`
	Current     string    `json:"current,omitempty"`
	Error       string    `json:"error,omitempty"`
	Files       []string  `json:"files,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	cancel      context.CancelFunc
	trackPct    map[int]float64
}

type configResponse struct {
	Config dfi.Config   `json:"config"`
	HasARL bool         `json:"hasArl"`
	Source sessionState `json:"session"`
}

type previewRequest struct {
	Query string `json:"query"`
}

type searchOptionsRequest struct {
	Type  string `json:"type"`
	Query string `json:"query"`
}

type searchOptionsResponse struct {
	Type    string         `json:"type"`
	Options []searchOption `json:"options"`
}

type searchOption struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

type previewResponse struct {
	LinkType string         `json:"linkType"`
	Tracks   []trackPreview `json:"tracks"`
}

type trackPreview struct {
	Index    int    `json:"index"`
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int    `json:"duration"`
}

type startRequest struct {
	Query     string `json:"query"`
	Quality   string `json:"quality"`
	SaveToDir string `json:"saveToDir"`
	CoverSize int    `json:"coverSize"`
	Tracks    []int  `json:"tracks"`
}

type jobResponse struct {
	Job *downloadJob `json:"job"`
}

func Run(ctx context.Context, opts Options) error {
	srv := NewServer(opts)
	server := &http.Server{
		Addr:              opts.Addr,
		Handler:           srv,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errs := make(chan error, 1)
	go srv.autoConnect()
	go func() {
		log.Printf("d-fi web listening on http://%s", opts.Addr)
		errs <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errs:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func NewServer(opts Options) *Server {
	if opts.ConfigPath == "" {
		opts.ConfigPath = "d-fi.config.json"
	}
	s := &Server{
		cfgPath: opts.ConfigPath,
		cfg:     dfi.LoadConfig(opts.ConfigPath),
		mux:     http.NewServeMux(),
		jobs:    map[int64]*downloadJob{},
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /api/config", s.handleConfig)
	s.mux.HandleFunc("PUT /api/config", s.handleUpdateConfig)
	s.mux.HandleFunc("POST /api/search-options", s.handleSearchOptions)
	s.mux.HandleFunc("POST /api/preview", s.handlePreview)
	s.mux.HandleFunc("POST /api/downloads", s.handleStartDownload)
	s.mux.HandleFunc("GET /api/jobs", s.handleJobs)
	s.mux.HandleFunc("POST /api/jobs/{id}/cancel", s.handleCancelJob)
	s.mux.HandleFunc("DELETE /api/jobs/{id}", s.handleDeleteJob)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	cfg := s.currentConfig()
	cfg.Cookies.ARL = ""
	writeJSON(w, http.StatusOK, configResponse{
		Config: cfg,
		HasARL: s.resolveARL() != "",
		Source: s.currentSession(),
	})
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg dfi.Config
	if err := readJSON(r, &cfg); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	s.mu.Lock()
	newARL := strings.TrimSpace(cfg.Cookies.ARL)
	currentARL := s.cfg.Cookies.ARL
	if cfg.Cookies.ARL == "" {
		cfg.Cookies.ARL = currentARL
	}
	s.cfg.Concurrency = max(1, cfg.Concurrency)
	s.cfg.SaveLayout = cfg.SaveLayout
	s.cfg.Playlist = cfg.Playlist
	s.cfg.TrackNumber = cfg.TrackNumber
	s.cfg.FallbackTrack = cfg.FallbackTrack
	s.cfg.FallbackQuality = cfg.FallbackQuality
	s.cfg.CoverSize = cfg.CoverSize
	s.cfg.Cookies = cfg.Cookies
	cfgToSave := s.cfg
	s.mu.Unlock()

	if err := cfgToSave.Save(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if newARL != "" {
		if _, err := s.connectWithARL(newARL); err != nil {
			writeError(w, http.StatusUnauthorized, err)
			return
		}
	}
	s.handleConfig(w, r)
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	if err := s.ensureSession(); err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	var req previewRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	res, err := s.resolveInput(strings.TrimSpace(req.Query))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, previewResponse{
		LinkType: res.linkType,
		Tracks:   previewTracks(res.tracks),
	})
}

func (s *Server) handleSearchOptions(w http.ResponseWriter, r *http.Request) {
	if err := s.ensureSession(); err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}
	var req searchOptionsRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	options, err := searchOptions(strings.TrimSpace(req.Type), strings.TrimSpace(req.Query))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, searchOptionsResponse{
		Type:    req.Type,
		Options: options,
	})
}

func (s *Server) handleStartDownload(w http.ResponseWriter, r *http.Request) {
	if err := s.ensureSession(); err != nil {
		writeError(w, http.StatusUnauthorized, err)
		return
	}

	var req startRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing URL or search"))
		return
	}

	cfg := s.currentConfig()
	_, label, err := parseQuality(req.Quality)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if req.CoverSize > 0 {
		switch label {
		case "128":
			cfg.CoverSize.MP3_128 = req.CoverSize
		case "flac":
			cfg.CoverSize.FLAC = req.CoverSize
		default:
			cfg.CoverSize.MP3_320 = req.CoverSize
		}
	}

	res, err := s.resolveInput(req.Query)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	tracks := selectTracks(res.tracks, req.Tracks)
	if len(tracks) == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("no tracks selected"))
		return
	}
	pathTemplate := cfg.Layout(res.linkType)
	if req.SaveToDir != "" {
		pathTemplate = filepath.Join(req.SaveToDir, "{SNG_TITLE}")
	}

	ctx, cancel := context.WithCancel(context.Background())
	job := &downloadJob{
		ID:          atomic.AddInt64(&s.nextID, 1),
		Source:      req.Query,
		Quality:     label,
		SaveToDir:   pathTemplate,
		Status:      "queued",
		TotalTracks: len(tracks),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		cancel:      cancel,
		trackPct:    map[int]float64{},
	}
	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()

	go s.runDownloadJob(ctx, job.ID, res.linkType, res.linkInfo, tracks, pathTemplate, label, cfg, concurrency)
	writeJSON(w, http.StatusAccepted, jobResponse{Job: s.snapshotJob(job.ID)})
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	jobs := make([]*downloadJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, cloneJob(job))
	}
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
}

func (s *Server) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	job := s.snapshotJob(id)
	if job == nil {
		writeError(w, http.StatusNotFound, fmt.Errorf("job not found"))
		return
	}
	s.mu.Lock()
	if live := s.jobs[id]; live != nil && live.cancel != nil {
		live.cancel()
	}
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, jobResponse{Job: s.snapshotJob(id)})
}

func (s *Server) handleDeleteJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	s.mu.Lock()
	if job := s.jobs[id]; job != nil && job.cancel != nil {
		job.cancel()
	}
	delete(s.jobs, id)
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) runDownloadJob(ctx context.Context, jobID int64, linkType string, info any, tracks []types.TrackType, pathTemplate, quality string, cfg dfi.Config, concurrency int) {
	s.updateJob(jobID, func(job *downloadJob) {
		job.Status = "running"
	})

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var failed atomic.Int64

	for i, track := range tracks {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(index int, track types.TrackType) {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				return
			}
			s.updateJob(jobID, func(job *downloadJob) {
				job.Current = track.SNG_TITLE + " - " + track.ART_NAME
			})
			path, err := dfi.DownloadTrack(ctx, dfi.DownloadTrackOptions{
				Track:           track,
				Quality:         quality,
				Info:            info,
				CoverSizes:      cfg.CoverSize,
				Path:            pathTemplate,
				TotalTracks:     len(tracks),
				TrackNumber:     cfg.TrackNumber,
				FallbackTrack:   cfg.FallbackTrack,
				FallbackQuality: cfg.FallbackQuality,
				Hooks: dfi.DownloadTrackHooks{
					Status: func(message string) {
						s.updateJob(jobID, func(job *downloadJob) {
							job.Current = message
						})
					},
					Progress: func(track types.TrackType, transferred, total int64) {
						if total <= 0 {
							return
						}
						progress := float64(transferred) / float64(total) * 100
						s.updateJob(jobID, func(job *downloadJob) {
							job.trackPct[index] = progress
							job.Progress = jobProgress(job)
						})
					},
				},
			})
			s.updateJob(jobID, func(job *downloadJob) {
				delete(job.trackPct, index)
				if err != nil {
					failed.Add(1)
					job.Error = err.Error()
				} else {
					job.DoneTracks++
					job.Progress = jobProgress(job)
					if path != "" {
						job.Files = append(job.Files, path)
					}
				}
			})
		}(i, track)
	}

	wg.Wait()
	playlistPath := ""
	if ctx.Err() == nil && failed.Load() == 0 && linkType == "playlist" && os.Getenv("SIMULATE") == "" {
		if job := s.snapshotJob(jobID); job != nil && len(job.Files) > 1 {
			var err error
			playlistPath, err = dfi.WritePlaylistFile(info, job.Files, cfg.Playlist.ResolveFullPath)
			if err != nil {
				s.updateJob(jobID, func(job *downloadJob) {
					job.Status = "error"
					job.Error = err.Error()
				})
				return
			}
		}
	}
	s.updateJob(jobID, func(job *downloadJob) {
		if ctx.Err() != nil {
			job.Status = "canceled"
			job.Error = ctx.Err().Error()
			return
		}
		if failed.Load() > 0 {
			job.Status = "error"
			return
		}
		job.Status = "done"
		job.Progress = 100
		job.Current = ""
		if playlistPath != "" {
			job.Files = append(job.Files, playlistPath)
		}
	})
}

type resolvedInput struct {
	linkType string
	linkInfo any
	tracks   []types.TrackType
}

func (s *Server) resolveInput(query string) (resolvedInput, error) {
	if query == "" {
		return resolvedInput{}, fmt.Errorf("missing URL or search")
	}
	if dfi.LooksLikeURL(query) {
		data, err := converter.ParseInfo(query)
		if err != nil {
			return resolvedInput{}, err
		}
		tracks := data.Tracks
		if data.LinkType == "playlist" {
			tracks = dfi.DedupePlaylistTracks(tracks)
		}
		return resolvedInput{linkType: data.LinkType, linkInfo: data.LinkInfo, tracks: tracks}, nil
	}

	switch {
	case strings.HasPrefix(query, "artist:"):
		search, err := api.SearchMusic(strings.TrimPrefix(query, "artist:"), 1, "ARTIST")
		if err != nil {
			return resolvedInput{}, err
		}
		if len(search.ARTIST.Data) == 0 {
			return resolvedInput{}, fmt.Errorf("no artist found")
		}
		return s.resolveInput("https://deezer.com/us/artist/" + search.ARTIST.Data[0].ART_ID)
	case strings.HasPrefix(query, "album:"):
		search, err := api.SearchMusic(strings.TrimPrefix(query, "album:"), 1, "ALBUM")
		if err != nil {
			return resolvedInput{}, err
		}
		if len(search.ALBUM.Data) == 0 {
			return resolvedInput{}, fmt.Errorf("no album found")
		}
		return s.resolveInput("https://deezer.com/us/album/" + search.ALBUM.Data[0].ALB_ID)
	case strings.HasPrefix(query, "playlist:"):
		search, err := api.SearchMusic(strings.TrimPrefix(query, "playlist:"), 1, "PLAYLIST")
		if err != nil {
			return resolvedInput{}, err
		}
		if len(search.PLAYLIST.Data) == 0 {
			return resolvedInput{}, fmt.Errorf("no playlist found")
		}
		return s.resolveInput("https://deezer.com/us/playlist/" + search.PLAYLIST.Data[0].PlaylistID)
	default:
		search, err := api.SearchMusic(query, 15, "TRACK")
		if err != nil {
			return resolvedInput{}, err
		}
		tracks := append([]types.TrackType(nil), search.TRACK.Data...)
		return resolvedInput{linkType: "track", tracks: tracks}, nil
	}
}

func (s *Server) currentConfig() dfi.Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg
}

func (s *Server) currentSession() sessionState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.session
}

func (s *Server) setSession(state sessionState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.session = state
}

func (s *Server) autoConnect() {
	arl := s.resolveARL()
	if arl == "" {
		return
	}
	if _, err := s.connectWithARL(arl); err != nil {
		log.Printf("d-fi web auto-connect failed: %v", err)
	}
}

func (s *Server) ensureSession() error {
	if s.currentSession().Ready {
		return nil
	}
	arl := s.resolveARL()
	if arl == "" {
		return fmt.Errorf("missing Deezer ARL")
	}
	_, err := s.connectWithARL(arl)
	return err
}

func (s *Server) connectWithARL(arl string) (sessionState, error) {
	arl = strings.TrimSpace(arl)
	if arl == "" {
		err := fmt.Errorf("missing Deezer ARL")
		state := sessionState{Error: err.Error()}
		s.setSession(state)
		return state, err
	}
	if _, err := request.InitDeezerAPI(arl); err != nil {
		state := sessionState{Error: err.Error()}
		s.setSession(state)
		return state, err
	}
	user, err := api.GetUser()
	if err != nil {
		state := sessionState{Error: err.Error()}
		s.setSession(state)
		return state, err
	}
	state := sessionState{Ready: true, UserName: user.BlogName}
	s.setSession(state)
	return state, nil
}

func (s *Server) resolveARL() string {
	if arl := strings.TrimSpace(os.Getenv("DEEZER_ARL")); arl != "" {
		return arl
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.TrimSpace(s.cfg.Cookies.ARL)
}

func (s *Server) updateJob(id int64, update func(*downloadJob)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job := s.jobs[id]
	if job == nil {
		return
	}
	update(job)
	job.UpdatedAt = time.Now()
}

func (s *Server) snapshotJob(id int64) *downloadJob {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneJob(s.jobs[id])
}

func cloneJob(job *downloadJob) *downloadJob {
	if job == nil {
		return nil
	}
	out := *job
	out.cancel = nil
	out.trackPct = nil
	out.Files = append([]string(nil), job.Files...)
	return &out
}

func jobProgress(job *downloadJob) float64 {
	if job.TotalTracks <= 0 {
		return 0
	}
	sum := float64(job.DoneTracks) * 100
	for _, progress := range job.trackPct {
		sum += progress
	}
	return sum / float64(job.TotalTracks)
}

func previewTracks(tracks []types.TrackType) []trackPreview {
	out := make([]trackPreview, 0, len(tracks))
	for i, track := range tracks {
		out = append(out, trackPreview{
			Index:    i,
			ID:       track.SNG_ID,
			Title:    track.SNG_TITLE,
			Artist:   track.ART_NAME,
			Album:    track.ALB_TITLE,
			Duration: dfi.AsInt(track.DURATION),
		})
	}
	return out
}

func searchOptions(searchType, query string) ([]searchOption, error) {
	if query == "" {
		return nil, fmt.Errorf("missing search text")
	}
	switch searchType {
	case "artist":
		search, err := api.SearchMusic(query, 15, "ARTIST")
		if err != nil {
			return nil, err
		}
		options := make([]searchOption, 0, len(search.ARTIST.Data))
		for _, item := range search.ARTIST.Data {
			options = append(options, searchOption{
				Title:       item.ART_NAME,
				Description: fmt.Sprintf("%d fans", item.NB_FAN),
				URL:         "https://deezer.com/us/artist/" + item.ART_ID,
			})
		}
		return options, nil
	case "album":
		search, err := api.SearchMusic(query, 15, "ALBUM")
		if err != nil {
			return nil, err
		}
		options := make([]searchOption, 0, len(search.ALBUM.Data))
		for _, item := range search.ALBUM.Data {
			options = append(options, searchOption{
				Title:       item.ALB_TITLE,
				Description: fmt.Sprintf("by %s, %s tracks", item.ART_NAME, item.NUMBER_TRACK),
				URL:         "https://deezer.com/us/album/" + item.ALB_ID,
			})
		}
		return options, nil
	case "playlist":
		search, err := api.SearchMusic(query, 15, "PLAYLIST")
		if err != nil {
			return nil, err
		}
		options := make([]searchOption, 0, len(search.PLAYLIST.Data))
		for _, item := range search.PLAYLIST.Data {
			options = append(options, searchOption{
				Title:       item.Title,
				Description: fmt.Sprintf("by %s, %d tracks", item.ParentUsername, item.NbSong),
				URL:         "https://deezer.com/us/playlist/" + item.PlaylistID,
			})
		}
		return options, nil
	default:
		return nil, fmt.Errorf("unsupported search type: %s", searchType)
	}
}

func selectTracks(tracks []types.TrackType, indexes []int) []types.TrackType {
	if len(indexes) == 0 {
		return tracks
	}
	out := make([]types.TrackType, 0, len(indexes))
	seen := map[int]bool{}
	for _, index := range indexes {
		if index < 0 || index >= len(tracks) || seen[index] {
			continue
		}
		seen[index] = true
		out = append(out, tracks[index])
	}
	return out
}

func parseQuality(value string) (int, string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "", "320", "3", "mp3_320", "320kbps", "128", "1", "mp3_128", "128kbps", "flac", "9":
		quality, _, label := dfi.ParseQuality(value)
		return quality, label, nil
	default:
		return 0, "", fmt.Errorf("invalid quality: %s", value)
	}
}

func parseID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid job id")
	}
	return id, nil
}

func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
