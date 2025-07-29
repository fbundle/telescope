package editor

import (
	"telescope/config"
	"telescope/side_channel"
	"time"
)

func insertToSlice[T any](l []T, i int, v T) []T {
	if i < 0 || i > len(l) {
		side_channel.Panic("invalid index", i, l)
		return nil
	}
	if i == len(l) {
		return append(l, v)
	}
	l = append(l, v)
	copy(l[i+1:], l[i:])
	l[i] = v
	return l
}

func deleteFromSlice[T any](l []T, i int) []T {
	if i < 0 || i >= len(l) {
		side_channel.Panic("invalid index", i, l)
		return nil
	}
	copy(l[i:], l[i+1:])
	return l[:len(l)-1]
}
func concatSlices[T any](ls ...[]T) []T {
	c := make([]T, 0)
	for _, l := range ls {
		c = append(c, l...)
	}
	return c
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
	if percentage > l.lastRenderPercentage || t.Sub(l.lastRenderTime) >= config.Load().LOADING_PROGRESS_INTERVAL {
		l.lastRenderPercentage = percentage
		l.lastRenderTime = t
		return true
	}
	return false
}
