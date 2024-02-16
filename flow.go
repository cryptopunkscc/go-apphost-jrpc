package rpc

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"log"
)

type Flow struct{ Serializer }

func NewFlow(conn io.ReadWriteCloser) Conn {
	s := Flow{
		Serializer{
			ReadWriteCloser: conn,
			enc:             json.NewEncoder(conn),
			dec:             json.NewDecoder(conn),
		},
	}
	switch conn2 := conn.(type) {
	case *astral.Conn:
		s.remoteID = conn2.RemoteIdentity()
	case net.SecureConn:
		s.remoteID = conn2.RemoteIdentity()
	}
	return &s
}

func QueryFlow(identity id.Identity, service string) (s Conn, err error) {
	query, err := astral.Query(identity, service)
	if err != nil {
		return
	}
	return NewFlow(query), nil
}

func (conn *Flow) WithLogger(logger *log.Logger) Conn {
	connLogger := NewConnLogger(conn.ReadWriteCloser, logger)
	conn.enc = json.NewEncoder(connLogger)
	conn.dec = json.NewDecoder(connLogger)
	return conn
}

func (conn *Flow) Copy() Conn {
	return conn
}

func (conn *Flow) Flush() {
	// no-op
}
