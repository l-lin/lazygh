package tui

import (
	"container/list"
	"sync"
)

type syntaxRangeCacheKey struct {
	hash   uint64
	length int
}

type syntaxRangeCacheEntry struct {
	key    syntaxRangeCacheKey
	text   string
	ranges []styledRuneRange
}

type syntaxRangeCache struct {
	maxEntries int
	mu         sync.Mutex
	order      list.List
	entries    map[syntaxRangeCacheKey][]*list.Element
}

func newSyntaxRangeCache(maxEntries int) *syntaxRangeCache {
	if maxEntries < 1 {
		maxEntries = 1
	}
	return &syntaxRangeCache{maxEntries: maxEntries, entries: make(map[syntaxRangeCacheKey][]*list.Element, maxEntries)}
}

func (cache *syntaxRangeCache) Get(text string) ([]styledRuneRange, bool) {
	if cache == nil || cache.maxEntries < 1 {
		return nil, false
	}

	key := syntaxRangeCacheKeyForText(text)
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for _, element := range cache.entries[key] {
		entry := element.Value.(*syntaxRangeCacheEntry)
		if entry.text != text {
			continue
		}
		cache.order.MoveToFront(element)
		return cloneStyledRuneRanges(entry.ranges), true
	}
	return nil, false
}

func (cache *syntaxRangeCache) Put(text string, ranges []styledRuneRange) {
	if cache == nil || cache.maxEntries < 1 {
		return
	}

	key := syntaxRangeCacheKeyForText(text)
	clonedRanges := cloneStyledRuneRanges(ranges)

	cache.mu.Lock()
	defer cache.mu.Unlock()

	for _, element := range cache.entries[key] {
		entry := element.Value.(*syntaxRangeCacheEntry)
		if entry.text != text {
			continue
		}
		entry.ranges = clonedRanges
		cache.order.MoveToFront(element)
		return
	}

	entry := &syntaxRangeCacheEntry{key: key, text: text, ranges: clonedRanges}
	element := cache.order.PushFront(entry)
	cache.entries[key] = append(cache.entries[key], element)
	for cache.order.Len() > cache.maxEntries {
		cache.removeOldest()
	}
}

func (cache *syntaxRangeCache) removeOldest() {
	element := cache.order.Back()
	if element == nil {
		return
	}
	cache.order.Remove(element)

	entry := element.Value.(*syntaxRangeCacheEntry)
	bucket := cache.entries[entry.key]
	for index, bucketElement := range bucket {
		if bucketElement != element {
			continue
		}
		bucket = append(bucket[:index], bucket[index+1:]...)
		break
	}
	if len(bucket) == 0 {
		delete(cache.entries, entry.key)
		return
	}
	cache.entries[entry.key] = bucket
}

func syntaxRangeCacheKeyForText(text string) syntaxRangeCacheKey {
	const (
		fnvOffset64 = 14695981039346656037
		fnvPrime64  = 1099511628211
	)

	hash := uint64(fnvOffset64)
	for index := range len(text) {
		hash ^= uint64(text[index])
		hash *= fnvPrime64
	}
	return syntaxRangeCacheKey{hash: hash, length: len(text)}
}

func cloneStyledRuneRanges(ranges []styledRuneRange) []styledRuneRange {
	if len(ranges) == 0 {
		return nil
	}
	return append([]styledRuneRange(nil), ranges...)
}
