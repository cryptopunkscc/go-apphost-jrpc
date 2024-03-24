package jrpc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry[any]()
	r.Add("aaa", "a")
	r.Add("bbb", "b")

	s, v := r.Unfold("aaa")
	assert.Equal(t, "a", v)
	assert.Equal(t, "", s)

	s, v = r.Unfold("aaabb")
	assert.Equal(t, "a", v)
	assert.Equal(t, "bb", s)

	s, v = r.Unfold("bbb")
	assert.Equal(t, "b", v)
	assert.Equal(t, "", s)

	s, v = r.Unfold("a")
	assert.Equal(t, "a", v)
	assert.Equal(t, "", s)

	s, v = r.Unfold("ccc")
	assert.Equal(t, nil, v)
	assert.Equal(t, "ccc", s)
}

func TestRegistry_All(t *testing.T) {
	r := NewRegistry[any]()
	r.Add("a", "a")
	r.Add("aaa", "aaa")
	r.Add("bbb", "bbb")

	expected := map[string]any{
		"a":   "a",
		"aaa": "aaa",
		"bbb": "bbb",
	}
	actual := make(map[string]any)
	r.All(nil, actual)
	assert.Equal(t, expected, actual)
}
