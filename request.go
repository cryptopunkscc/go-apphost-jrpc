package jrpc

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
)

type Request struct {
	*Serializer
	service string
}

func NewRequest(
	identity id.Identity,
	service string,
) Conn {
	return &Request{
		Serializer: &Serializer{
			remoteID: identity,
		},
		service: service,
	}
}

func (r *Request) Copy() Conn {
	rr := NewRequest(r.remoteID, r.service)
	if r.logger != nil {
		rr.Logger(r.logger.Logger)
	}
	return rr
}

func (r *Request) Flush() {
	if r.WriteCloser != nil {
		_ = r.WriteCloser.Close()
	}
}

func (r *Request) Call(method string, value any) (err error) {
	// build base query
	query := ""
	switch {
	case r.service == "":
		query = method
	case method == "":
		query = r.service
	default:
		query = r.service + "." + method
	}

	// marshal args
	if value != nil {
		if query != "" {
			query += "?"
		}
		if r.marshal == nil {
			r.setupEncoding()
		}
		args, err := r.marshal(value)
		if err != nil {
			return err
		}
		query += string(args)
	}

	// log query
	if r.logger != nil {
		r.logger.Println("> " + query)
	}

	// query stream
	var conn io.ReadWriteCloser
	if conn, err = astral.Query(r.RemoteIdentity(), query); err != nil {
		return
	}

	// setup
	r.setConn(conn)
	return
}
