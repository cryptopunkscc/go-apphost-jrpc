package jrpc

import (
	"errors"
	"io"
)

type ArgsDecoder interface {
	Test([]byte) bool
	TestScan(stream io.ByteScanner) bool
	Begin() []rune
	Unmarshal(bytes []byte, args []any) error
	Decode(conn ByteScannerReader, args []any) error
}

type argsDecoders struct {
	decoders []ArgsDecoder
}

func (a *argsDecoders) Append(decoders []ArgsDecoder) *argsDecoders {
	a.decoders = append(a.decoders, decoders...)
	return a
}

func (a *argsDecoders) TestScan(scan io.ByteScanner) bool {
	for _, decoder := range a.decoders {
		if decoder.TestScan(scan) {
			return true
		}
	}
	return false
}

func (a *argsDecoders) Decode(conn ByteScannerReader, args []any) error {
	for _, decoder := range a.decoders {
		if decoder.TestScan(conn) {
			return decoder.Decode(conn, args)
		}
	}
	return errors.New("unknown format")
}

func (a *argsDecoders) Test(bytes []byte) bool {
	for _, decoder := range a.decoders {
		if decoder.Test(bytes) {
			return true
		}
	}
	return false
}

func (a *argsDecoders) Begin() (r []rune) {
	for _, decoder := range a.decoders {
		r = append(r, decoder.Begin()...)
	}
	return
}

func (a *argsDecoders) Unmarshal(bytes []byte, args []any) error {
	for _, decoder := range a.decoders {
		if decoder.Test(bytes) {
			return decoder.Unmarshal(bytes, args)
		}
	}
	return errors.New("unknown format")
}
