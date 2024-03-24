package jrpc

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	"log"
)

type Flow struct{ Serializer }

func NewFlow(conn io.ReadWriteCloser) *Flow {
	sc := NewByteScannerReadWriteCloser(conn)
	s := Flow{
		Serializer{
			ReadWriteCloser: sc,
			ByteScanner:     sc,
			enc:             json.NewEncoder(sc),
			dec:             json.NewDecoder(sc),
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
	conn.logger = NewConnLogger(conn.ReadWriteCloser, logger)
	conn.enc = json.NewEncoder(conn.logger)
	conn.dec = json.NewDecoder(conn.logger)
	return conn
}

func (conn *Flow) Call(method string, value any) (err error) {
	query := []byte(method)
	if value != nil {
		var bytes []byte
		if bytes, err = json.Marshal(value); err != nil {
			return
		}
		query = append(query, bytes...)
	}
	query = append(query, []byte("\n")...)
	_, err = conn.logger.Write(query)
	return
}

func (conn *Flow) Copy() Conn {
	return conn
}

func (conn *Flow) Flush() {
	// no-op
}
