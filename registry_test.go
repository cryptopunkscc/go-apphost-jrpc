package jrpc

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestRegistry2(t *testing.T) {
	var err error = nil
	arr := []any{"a", "b", err}
	for n := len(arr) - 1; n > 0 && arr[n] == nil; n-- {
		arr = arr[0:n]
	}
	for _, a := range arr {
		if err, ok := a.(error); err == nil {
			log.Println(ok)
			log.Println(a)
		}
	}
}

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
