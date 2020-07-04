// Go Access Cache keeping most requested elements in memory
// Usage: m := NewInMemoryCache(<maximum byte size>)
// Store data: m.Set(<key>, <value>)
// Read data: <value>, exists = m.Get(<key>)
package accesscache

import (
	"errors"
	"log"
	"reflect"
	"sync"
	"time"
)

// Go Access Cache Structure
type AccessCache struct {
	// cache holding keys and objects
	cache map[string]interface{}
	// array to track most recently viewed items
	lastviewed []string
	// mutex lock
	mux sync.Mutex
	// maximum cachsize in bytes
	maxsize uint64
	// verbosity for logging
	verbose bool
	// average time in ms to get entries from cache
	avgGet float64
	// average time in ms to add entries to cache
	avgSet float64
	// count avg get
	ctrGet int64
	// count avg set
	ctrSet int64
	// item sizes in bytes
	sizes map[string]uint64
}

// NewAccessCache constructs a new cache
// where size is the maximum size in bytes
func NewAccessCache(size uint64) AccessCache {
	if size <= 0 {
		panic("Size in bytes must be greater 0")
	}

	m := AccessCache{
		cache:      make(map[string]interface{}),
		sizes:      make(map[string]uint64),
		lastviewed: []string{},
		maxsize:    size,
		verbose:    false,
		avgGet:     0.0,
		avgSet:     0.0,
		ctrGet:     0,
		ctrSet:     0,
	}
	return m
}

// indexOfLastViewed gets an element index from last viewed slice
func (c *AccessCache) indexOfLastViewed(element string) int {
	for k, v := range c.lastviewed {
		if element == v {
			return k
		}
	}
	return -1
}

// removeLastViewedAtIndex removes an index from last viewed slice
func (c *AccessCache) removeLastViewedAtIndex(index int) {
	c.lastviewed = append(c.lastviewed[:index], c.lastviewed[index+1:]...)
}

// appendLastViewed removes an item from last viewed slice if exists and append item
func (c *AccessCache) appendLastViewed(key string) {
	i := c.indexOfLastViewed(key)
	if i > -1 {
		c.removeLastViewedAtIndex(i)
	}
	c.lastviewed = append(c.lastviewed, key)
}

// clearOutdatedItems clears outdated items from cache by last viewed
func (c *AccessCache) clearOutdatedItems() {
	for c.GetCacheSize() > c.maxsize {

		// if there is nothing left to order return
		if len(c.lastviewed) == 0 {
			return
		}

		// clear cache for oldest item and remove from last viewed slice
		delete(c.cache, c.lastviewed[0])
		delete(c.sizes, c.lastviewed[0])
		c.removeLastViewedAtIndex(0)
	}

	if c.verbose {
		log.Println("Size:", c.GetCacheSize(), "bytes", "Order:", c.lastviewed)
	}
}

// Get gets an item from cache by key
func (c *AccessCache) Get(key string) (interface{}, bool) {
	start := time.Now()

	c.mux.Lock()
	defer c.mux.Unlock()

	value, ok := c.cache[key]

	// if the value exists: update last viewed
	if ok {
		c.appendLastViewed(key)
	}

	stop := time.Now()
	c.avgGet = calcAvg(c.avgGet, c.ctrGet, stop.Sub(start).Seconds())
	c.ctrGet++
	return value, ok
}

// Set adds or updates an item from cache
// keep in mind that the object you are adding
// should be smaller than the maximum memory of the cache
func (c *AccessCache) Set(key string, value interface{}) error {
	start := time.Now()

	size := sizeof(value)
	if size >= c.maxsize {
		return errors.New("Cannot add elements larger than the maximum cache size")
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	// set the cache value and update last viewed
	c.appendLastViewed(key)
	c.cache[key] = value
	c.sizes[key] = size

	// clear outdated items from cache
	c.clearOutdatedItems()

	stop := time.Now()
	c.avgSet = calcAvg(c.avgSet, c.ctrSet, stop.Sub(start).Seconds())
	c.ctrSet++
	return nil
}

// calcAvg calculates average with correct weights over time
func calcAvg(currAvg float64, currCtr int64, newValue float64) float64 {
	val := newValue * 1000
	return (currAvg*float64(currCtr) + val) / (float64(currCtr) + 1)
}

// GetCacheSize gets the current cache size in bytes
func (c *AccessCache) GetCacheSize() uint64 {
	var size uint64
	size = 0
	for k, _ := range c.cache {
		size = size + c.sizes[k]
	}
	return size
}

// GetItemSizes gets cache size in bytes of all items
func (c *AccessCache) GetItemSizes() map[string]uint64 {
	return c.sizes
}

// Count returns the number if cached items
func (c *AccessCache) Count() int {
	return len(c.lastviewed)
}

// GetAverageDurationForGet returns average ms for Get
func (c *AccessCache) GetAverageDurationForGet() float64 {
	return c.avgGet
}

// GetAverageDurationForSet returns average ms for Set
func (c *AccessCache) GetAverageDurationForSet() float64 {
	return c.avgSet
}

// GetLastViewedKey gets the last viewed key from the cache
func (c *AccessCache) GetLastViewedKey() string {
	c.mux.Lock()
	defer c.mux.Unlock()
	i := len(c.lastviewed) - 1
	if i < 0 {
		return ""
	} else {
		return c.lastviewed[i]
	}
}

var (
	sliceSize  = uint64(reflect.TypeOf(reflect.SliceHeader{}).Size())
	stringSize = uint64(reflect.TypeOf(reflect.StringHeader{}).Size())
)

// isNativeType checks if a type is a native golang type
func isNativeType(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	}
	return false
}

// sizeofInternal Gets memory allocation
func sizeofInternal(val reflect.Value, fromStruct bool, depth int) (sz uint64) {
	if depth++; depth > 1000 {
		panic("sizeOf recursed more than 1000 times.")
	}

	typ := val.Type()

	if !fromStruct {
		sz = uint64(typ.Size())
	}

	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			break
		}
		sz += sizeofInternal(val.Elem(), false, depth)

	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			sz += sizeofInternal(val.Field(i), true, depth)
		}

	case reflect.Array:
		if isNativeType(typ.Elem().Kind()) {
			break
		}
		sz = 0
		for i := 0; i < val.Len(); i++ {
			sz += sizeofInternal(val.Index(i), false, depth)
		}
	case reflect.Slice:
		if !fromStruct {
			sz = sliceSize
		}
		el := typ.Elem()
		if isNativeType(el.Kind()) {
			sz += uint64(val.Len()) * uint64(el.Size())
			break
		}
		for i := 0; i < val.Len(); i++ {
			sz += sizeofInternal(val.Index(i), false, depth)
		}
	case reflect.Map:
		if val.IsNil() {
			break
		}
		kel, vel := typ.Key(), typ.Elem()
		if isNativeType(kel.Kind()) && isNativeType(vel.Kind()) {
			sz += uint64(kel.Size()+vel.Size()) * uint64(val.Len())
			break
		}
		keys := val.MapKeys()
		for i := 0; i < len(keys); i++ {
			sz += sizeofInternal(keys[i], false, depth) + sizeofInternal(val.MapIndex(keys[i]), false, depth)
		}
	case reflect.String:
		if !fromStruct {
			sz = stringSize
		}
		sz += uint64(val.Len())
	}
	return
}

// sizeof returns the estimated memory usage of object(s) not just the size of the type.
func sizeof(objs ...interface{}) (sz uint64) {
	for i := range objs {
		sz += sizeofInternal(reflect.ValueOf(objs[i]), false, 0)
	}
	return
}
