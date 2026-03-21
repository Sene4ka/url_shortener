package generator

import (
	"testing"
	"time"
)

func TestGenerateId_LengthAndChars(t *testing.T) {
	g, err := NewIdGenerator(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), 10)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		id, err := g.GenerateId()
		if err != nil {
			t.Fatal(err)
		}
		if len(id) != 10 {
			t.Errorf("expected length 10, got %d: %s", len(id), id)
		}
		for _, c := range id {
			if !isAllowedChar(c) {
				t.Errorf("disallowed character %c in %s", c, id)
			}
		}
	}
}

func TestGenerateId_Uniqueness(t *testing.T) {
	g, err := NewIdGenerator(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), 10)
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id, err := g.GenerateId()
		if err != nil {
			t.Fatal(err)
		}
		if seen[id] {
			t.Errorf("duplicate id generated: %s", id)
		}
		seen[id] = true
	}
}

func isAllowedChar(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z')
}
