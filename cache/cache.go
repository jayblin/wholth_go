package cache

import "sync"

var G_cache = make(map[string]map[string]any)
var G_cache_mutex sync.Mutex

func Set(group string, key string, value any) {
	G_cache_mutex.Lock()
	defer G_cache_mutex.Unlock()

	if nil == G_cache[group] {
		G_cache[group] = make(map[string]any)
	}

	G_cache[group][key] = value
}

// todo check if value or ref is returned
func Get(group string, key string) (any, bool) {
	G_cache_mutex.Lock()
	defer G_cache_mutex.Unlock()

	grp, ok := G_cache[group]

	if !ok  {
		return nil, ok
	}

	val, ok := grp[key]

	return val, ok
}

func Delete(group string, key string) {
	G_cache_mutex.Lock()
	defer G_cache_mutex.Unlock()

	grp, ok := G_cache[group]

	if !ok  {
		return
	}

	delete(grp, key)
}
