# Go Access Cache

[![Go Report Card](https://goreportcard.com/badge/github.com/moethu/go-access-cache)](https://goreportcard.com/report/github.com/moethu/go-access-cache)


A simple in memory cache with a maximum byte size keeping most and recent requested elements in memory while rotating rarely requested elements out.

### Example

```go
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
```

### Usage

instantiate new access cache with a size of 1024 bytes
```go
m := NewAccessCache(1024)
```

add value 42 to cache using a unique key "mykey"
```go
m.Set("mykey", 42)
```

retrieve value from cache by key "mykey"
```go
value, exists = m.Get("mykey")
```
