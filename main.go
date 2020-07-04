package main

import (
	"log"

	"github.com/moethu/go-access-cache/accesscache"
)

func main() {
	m := accesscache.NewAccessCache(60)

	// Adding first item, cache size will be 28
	m.Set("key-a", "first string")
	log.Println("Highest Prio Item:", m.GetLastViewedKey(), "Cache Size:", m.GetCacheSize())

	// Adding second item, cache size will be 57
	m.Set("key-b", "second string")
	log.Println("Highest Prio Item:", m.GetLastViewedKey(), "Cache Size:", m.GetCacheSize())

	// Requesting first item, it get the highest prio, cache size remains 57
	m.Get("key-a")
	log.Println("Highest Prio Item:", m.GetLastViewedKey(), "Cache Size:", m.GetCacheSize())

	// Adding third item, lowest prio item gets rotated out because cache size > 60
	// third item get highest prio, cache size is at 56
	m.Set("key-c", "third string")
	log.Println("Highest Prio Item:", m.GetLastViewedKey(), "Cache Size:", m.GetCacheSize())

	// Requesting first item, wich again gets the highest prio, cache size is at 56
	// containing only the first and the second item
	m.Get("key-a")
	log.Println("Highest Prio Item:", m.GetLastViewedKey(), "Cache Size:", m.GetCacheSize())
}
