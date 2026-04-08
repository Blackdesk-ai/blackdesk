package storage

import (
	"testing"
	"time"
)

func TestMemoryCacheGetSet(t *testing.T) {
	cache := NewMemoryCache()
	cache.Set("a", []byte("hello"), time.Second)

	var out []byte
	if !cache.Get("a", &out) {
		t.Fatal("expected cache hit")
	}
	if string(out) != "hello" {
		t.Fatalf("unexpected value %q", string(out))
	}
}

func TestMemoryCacheExpiry(t *testing.T) {
	cache := NewMemoryCache()
	cache.Set("a", []byte("hello"), 5*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	var out []byte
	if cache.Get("a", &out) {
		t.Fatal("expected cache miss after expiry")
	}
}
