package cache

import (
	"github.com/coocood/freecache"
)

var (
	Cache *freecache.Cache
)

func InitCache() {
	cacheSize := 100 * 1024 * 1024
	Cache = freecache.NewCache(cacheSize)
}
