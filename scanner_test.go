package jrpc

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewReadWriteCloserScanner(t *testing.T) {
	scanner := &ReadWriteCloserScanner{
		ReadScanner: NewReadScanner(strings.NewReader(`[1,2,3]`)),
	}
	b, err := scanner.ReadByte()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, byte('['), b)

	err = scanner.UnreadByte()
	if err != nil {
		t.Fatal(err)
	}

	//arr := make([]int, 3)
	var arr []int
	err = json.NewDecoder(scanner).Decode(&arr)
	if err != nil {
		t.Fatal(err)
	}
}
