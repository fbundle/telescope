package editor

import (
	"telescope/feature"
	"time"
)

func insertToSlice[T any](l []T, i int, v T) []T {
	var zero T
	l = append(l, zero)
	copy(l[i+1:], l[i:])
	l[i] = v
	return l
}

func deleteFromSlice[T any](l []T, i int) []T {
	copy(l[i:], l[i+1:])
	return l[:len(l)-1]
}
func concatSlices[T any](ls ...[]T) []T {
	l := make([]T, 0)
	for _, s := range ls {
		l = append(l, s...)
	}
	return l
}

func newLoader(totalSize int) *loader {
	return &loader{
		totalSize:            totalSize,
		loadedSize:           0,
		lastRenderPercentage: -1,
		lastRenderTime:       time.Time{},
	}
}

type loader struct {
	totalSize            int
	loadedSize           int
	lastRenderPercentage int
	lastRenderTime       time.Time
}

func (l *loader) add(amount int) bool {
	l.loadedSize += amount
	percentage := int(100 * float64(l.loadedSize) / float64(l.totalSize))
	t := time.Now()
	if percentage > l.lastRenderPercentage || t.Sub(l.lastRenderTime) >= feature.LOADING_PROGRESS_INTERVAL_MS*time.Millisecond {
		l.lastRenderPercentage = percentage
		l.lastRenderTime = t
		return true
	}
	return false
}
