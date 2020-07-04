package accesscache

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func assert(t *testing.T, expected interface{}, actual interface{}) {
	if actual != expected {
		t.Errorf("Value was incorrect, got: %v, want: %v", actual, expected)
	}
}

// Test Appending to last viewed
func TestAppendLastViewed(t *testing.T) {
	m := NewAccessCache(500)
	m.appendLastViewed("a")
	assert(t, m.lastviewed[0], "a")
	m.appendLastViewed("b")
	assert(t, m.lastviewed[0], "a")
	m.appendLastViewed("a")
	assert(t, m.lastviewed[1], "a")
}

// Test size allocation and growth in bytes
func TestSizeAllocation(t *testing.T) {
	m := NewAccessCache(24)
	assert(t, uint64(24), m.maxsize)
	m.Set("a", 1024)
	assert(t, m.GetCacheSize(), uint64(8))
	m.Set("b", 1024)
	assert(t, m.GetCacheSize(), uint64(16))
	m.Set("c", 1024)
	assert(t, m.GetCacheSize(), uint64(24))
	m.Set("d", 1024)
	assert(t, m.GetCacheSize(), uint64(24))
}

// Test last viewed order
func TestLastViewedOrder(t *testing.T) {
	m := NewAccessCache(500)
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	assert(t, "a", m.lastviewed[0])
	assert(t, "c", m.lastviewed[2])
	assert(t, m.Count(), 3)
}

func TestSizes(t *testing.T) {
	m := NewAccessCache(900)
	rand.Seed(time.Now().UnixNano())
	err := m.Set("a", randSeq(1024))
	assert(t, "Cannot add elements larger than the maximum cache size", err.Error())
}

// test order and shifting
func TestReadWriteCacheAndOrder(t *testing.T) {
	m := NewAccessCache(40)
	m.Set("a", 1)
	assert(t, m.lastviewed[0], "a")
	m.Set("b", 2)
	assert(t, m.lastviewed[1], "b")
	m.Set("c", 3)
	assert(t, m.lastviewed[2], "c")

	// should move the item to the end of the queue
	value, _ := m.Get("a")
	assert(t, m.lastviewed[2], "a")
	assert(t, value, 1)

	m.Set("d", 4)
	assert(t, m.lastviewed[3], "d")
	m.Set("e", 5)
	assert(t, m.lastviewed[4], "e")
	m.Set("f", 6)
	assert(t, m.lastviewed[4], "f")
	m.Set("g", 7)
	assert(t, m.lastviewed[4], "g")
	assert(t, uint64(40), m.GetCacheSize())
}

func populateCache(i int, m *AccessCache, t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	m.Set(strconv.Itoa(i), randSeq(1024*1024*5))
	// if the collection contains 9 already, 9 needs to be lastviewed always
	_, ok := m.Get("9")
	if ok {
		assert(t, m.GetLastViewedKey(), "9")
	}
}

func TestGetItemSizesAndDuration(t *testing.T) {
	m := NewAccessCache(1024 * 1024 * 25)
	assert(t, len(m.GetItemSizes()), 0)
	assert(t, m.GetAverageDurationForGet(), 0.0)
	assert(t, m.GetAverageDurationForSet(), 0.0)
	m.Set("a", 1)
	assert(t, m.GetItemSizes()["a"], uint64(8))
}

// test concurrency and shifting
func TestConcurrency(t *testing.T) {
	m := NewAccessCache(1024 * 1024 * 25)
	m.verbose = true
	for i := 0; i < 10; i++ {
		go populateCache(i, &m, t)

	}
	time.Sleep(time.Duration(20000) * time.Millisecond)
}

// test large amounts of data
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type teststruct struct {
	name string
	age  int
}

func TestSizeOf(t *testing.T) {
	m := NewAccessCache(500)
	m.Set("a", 5)
	a := teststruct{}
	assert(t, int(sizeof("")), 16)
	assert(t, int(sizeof("a")), 16+1)
	assert(t, int(sizeof("abc")), 16+3)
	assert(t, int(sizeof(1024)), 8)
	assert(t, int(sizeof(a)), 16+8)
	assert(t, int(sizeof(m)), 171)
	a.age = 6
	a.name = "test"
	assert(t, int(sizeof(a)), 8+16+4)
}

func TestIsNativeType(t *testing.T) {
	m := NewAccessCache(500)
	v := reflect.ValueOf(m)
	assert(t, isNativeType(5), true)
	assert(t, isNativeType(v.Kind()), false)
}

func TestEmptyCache(t *testing.T) {
	m := NewAccessCache(500)
	m.clearOutdatedItems()
	key := m.GetLastViewedKey()
	assert(t, "", key)
}
