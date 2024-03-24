package jrpc

import (
	"io"
)

type ByteScannerReadWriteCloser interface {
	io.ByteScanner
	io.ReadWriteCloser
}
type ByteScannerReader interface {
	io.ByteScanner
	io.Reader
}

type byteScannerReadWriteCloser struct {
	io.WriteCloser
	ByteScannerReader
}

func NewByteScannerReadWriteCloser(rwc io.ReadWriteCloser) ByteScannerReadWriteCloser {
	return &byteScannerReadWriteCloser{WriteCloser: rwc, ByteScannerReader: NewByteScannerReader(rwc)}
}

type byteScannerReader struct {
	io.Reader
	offset int
	end    int
	buff   []byte
}

func NewByteScannerReader(reader io.Reader) ByteScannerReader {
	return &byteScannerReader{Reader: reader}
}

func (r *byteScannerReader) ReadByte() (b byte, err error) {
	if r.offset == len(r.buff) {
		r.buff = append(r.buff, make([]byte, 512)...)
		l := 0
		if l, err = r.Reader.Read(r.buff[r.offset:len(r.buff)]); err != nil {
			return
		}
		r.end = r.offset + l
	}

	b = r.buff[r.offset]
	r.offset++
	return
}

func (r *byteScannerReader) UnreadByte() error {
	if r.offset > 0 {
		r.offset--
	}
	return nil
}

func (r *byteScannerReader) Read(p []byte) (n int, err error) {
	if r.offset == r.end {
		return r.Reader.Read(p)
	}
	for n = 0; n < cap(p) && r.offset < r.end; n++ {
		p[n] = r.buff[r.offset]
		r.offset++
	}
	if n == cap(p) {
		return
	}
	l, err := r.Reader.Read(p[r.offset:])
	n += l
	if n > r.end {
		r.offset = 0
		r.end = 0
	}
	return
}

var _ io.ReadWriteCloser = &byteScannerReadWriteCloser{}
var _ io.ByteScanner = &byteScannerReader{}
