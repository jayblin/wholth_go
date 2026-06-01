package cache

import (
	"errors"
	"sync"
)

var G_cache = make(map[string]map[string]any)
var G_cache_mutex sync.Mutex

// @deprecated
func Set(group string, key string, value any) {
	G_cache_mutex.Lock()
	defer G_cache_mutex.Unlock()

	if nil == G_cache[group] {
		G_cache[group] = make(map[string]any)
	}

	G_cache[group][key] = value
}

// @deprecated
func Get(group string, key string) (any, bool) {
	G_cache_mutex.Lock()
	defer G_cache_mutex.Unlock()

	grp, ok := G_cache[group]

	if !ok {
		return nil, ok
	}

	val, ok := grp[key]

	return val, ok
}

// @deprecated
func Has(group string, key string) bool {
	G_cache_mutex.Lock()
	defer G_cache_mutex.Unlock()

	grp, ok := G_cache[group]

	if !ok {
		return ok
	}

	_, ok = grp[key]

	return ok
}

// @deprecated
func Delete(group string, key string) {
	G_cache_mutex.Lock()
	defer G_cache_mutex.Unlock()

	grp, ok := G_cache[group]

	if !ok {
		return
	}

	delete(grp, key)
}

var G_cache_v2 = make(map[string]any)
var G_cache_v2_mutex sync.Mutex
var G_tags = make(map[string][]string)

func SetV2(key string, value any, tags ...string) {
	G_cache_v2_mutex.Lock()
	defer G_cache_v2_mutex.Unlock()

	G_cache_v2[key] = value

	for _, tag := range tags {
		G_tags[tag] = append(G_tags[tag], key)
	}
}

func GetV2[T any](key string) (*T, error) {
	G_cache_v2_mutex.Lock()
	defer G_cache_v2_mutex.Unlock()

	res, ok := G_cache_v2[key]

	if !ok {
		return nil, nil
	}

	resT, ok := res.(T)

	if !ok {
		return nil, errors.New("CACHE_ENTRY_CAST_ERROR")
	}

	return &resT, nil
}

func DeleteV2(key string) {
	G_cache_v2_mutex.Lock()
	defer G_cache_v2_mutex.Unlock()

	_, ok := G_cache_v2[key]

	if ok {
		delete(G_cache_v2, key)
	}
}

func DeleteByTag(tag string) {
	for _, keys := range G_tags {
		for _, key := range keys {
			DeleteV2(key)
		}
	}
}
