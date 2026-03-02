package github

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("key1", "value1")

	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if v.(string) != "value1" {
		t.Errorf("expected %q, got %q", "value1", v.(string))
	}
}

func TestCache_Miss(t *testing.T) {
	c := NewCache(5 * time.Minute)

	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected cache miss for nonexistent key")
	}
}

func TestCache_TTLExpiry(t *testing.T) {
	c := NewCache(1 * time.Millisecond)
	c.Set("expiring", "data")

	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get("expiring")
	if ok {
		t.Error("expected cache miss after TTL expiry")
	}
}

func TestCache_OverwriteKey(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("k", "first")
	c.Set("k", "second")

	v, ok := c.Get("k")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v.(string) != "second" {
		t.Errorf("expected %q, got %q", "second", v.(string))
	}
}

func TestCache_MultipleKeys(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("a", 1)
	c.Set("b", 2)

	va, _ := c.Get("a")
	vb, _ := c.Get("b")
	if va.(int) != 1 || vb.(int) != 2 {
		t.Errorf("unexpected values: a=%v b=%v", va, vb)
	}
}
