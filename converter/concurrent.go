package converter

import (
	"sync"

	"github.com/d-fi/GoFi/types"
)

const converterConcurrency = 10

func convertTracksConcurrently[T any](items []T, convert func(int, T) (types.TrackType, bool)) []types.TrackType {
	type result struct {
		index int
		track types.TrackType
		ok    bool
	}

	jobs := make(chan int)
	results := make(chan result, len(items))

	workerCount := min(len(items), converterConcurrency)

	var wg sync.WaitGroup
	for range workerCount {
		wg.Go(func() {
			for index := range jobs {
				track, ok := convert(index, items[index])
				results <- result{index: index, track: track, ok: ok}
			}
		})
	}

	for index := range items {
		jobs <- index
	}
	close(jobs)
	wg.Wait()
	close(results)

	ordered := make([]*types.TrackType, len(items))
	for result := range results {
		if !result.ok {
			continue
		}
		track := result.track
		ordered[result.index] = &track
	}

	tracks := make([]types.TrackType, 0, len(items))
	for _, track := range ordered {
		if track != nil {
			tracks = append(tracks, *track)
		}
	}
	return tracks
}

func convertTrackListsConcurrently[T any](items []T, convert func(int, T) []types.TrackType) []types.TrackType {
	type result struct {
		index  int
		tracks []types.TrackType
	}

	jobs := make(chan int)
	results := make(chan result, len(items))

	workerCount := min(len(items), converterConcurrency)

	var wg sync.WaitGroup
	for range workerCount {
		wg.Go(func() {
			for index := range jobs {
				results <- result{index: index, tracks: convert(index, items[index])}
			}
		})
	}

	for index := range items {
		jobs <- index
	}
	close(jobs)
	wg.Wait()
	close(results)

	ordered := make([][]types.TrackType, len(items))
	for result := range results {
		ordered[result.index] = result.tracks
	}

	tracks := make([]types.TrackType, 0, len(items))
	for _, group := range ordered {
		tracks = append(tracks, group...)
	}
	return tracks
}
