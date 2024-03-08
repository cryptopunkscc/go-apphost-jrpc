package jrpc

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
	"log"
)

type Request struct {
	Serializer
	service   string
	newLogger func(closer io.ReadWriteCloser) io.ReadWriteCloser
}

func (r *Request) WithLogger(logger *log.Logger) Conn {
	r.newLogger = func(conn io.ReadWriteCloser) io.ReadWriteCloser {
		return NewConnLogger(conn, logger)
	}
	return r
}

func NewRequest(
	identity id.Identity,
	service string,
) Conn {
	return &Request{
		Serializer: Serializer{
			remoteID: identity,
		},
		service: service,
	}
}

func (r *Request) Copy() Conn {
	c := *r
	return &c
}

func (r *Request) Flush() {
	if r.ReadWriteCloser != nil {
		_ = r.ReadWriteCloser.Close()
	}
}

func (r *Request) Encode(value any) (err error) {
	u := r.service

	if r.ReadWriteCloser, err = astral.Query(r.remoteID, u); err != nil {
		return
	}
	logger := r.ReadWriteCloser
	if r.newLogger != nil {
		logger = r.newLogger(logger)
	}
	r.enc = json.NewEncoder(logger)
	r.dec = json.NewDecoder(logger)

	err = r.Serializer.Encode(value)

	return
}
