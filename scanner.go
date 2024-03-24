package jrpc

import (
	"io"
)

type ReadWriteCloserScanner struct {
	io.WriteCloser
	*ReadScanner
}

func NewReadWriteCloserScanner(rwc io.ReadWriteCloser) *ReadWriteCloserScanner {
	return &ReadWriteCloserScanner{WriteCloser: rwc, ReadScanner: NewReadScanner(rwc)}
}

type ReadScanner struct {
	io.Reader
	offset int
	end    int
	buff   []byte
}

func NewReadScanner(reader io.Reader) *ReadScanner {
	return &ReadScanner{Reader: reader}
}

func (r *ReadScanner) ReadByte() (b byte, err error) {
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

func (r *ReadScanner) UnreadByte() error {
	if r.offset > 0 {
		r.offset--
	}
	return nil
}

func (r *ReadScanner) Read(p []byte) (n int, err error) {
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

var _ io.ReadWriteCloser = &ReadWriteCloserScanner{}
var _ io.ByteScanner = &ReadScanner{}
