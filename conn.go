package rpc

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	net2 "github.com/cryptopunkscc/astrald/net"
	"io"
)

type Conn struct {
	io.ReadWriteCloser
	enc      *json.Encoder
	dec      *json.Decoder
	remoteID id.Identity
}

func (conn *Conn) RemoteIdentity() id.Identity {
	return conn.remoteID
}

func NewConn(conn io.ReadWriteCloser) (c *Conn) {
	c = &Conn{
		ReadWriteCloser: conn,
		enc:             json.NewEncoder(conn),
		dec:             json.NewDecoder(conn),
	}
	switch conn2 := conn.(type) {
	case *astral.Conn:
		c.remoteID = conn2.RemoteIdentity()
	case net2.SecureConn:
		c.remoteID = conn2.RemoteIdentity()
	}
	return
}

func (conn *Conn) Encode(value any) (err error) {
	r := value
	switch v := value.(type) {
	case error:
		r = Failure{v.Error()}
	}
	return conn.enc.Encode(r)
}

func (conn *Conn) Decode(value any) (err error) {
	// decode raw value
	r := raw{}
	if err = conn.dec.Decode(&r); err != nil {
		return
	}

	// try decode as failure
	f := Failure{}
	if err = json.Unmarshal(r.bytes, &f); err == nil && f.Error != "" {
		return errors.New(f.Error)
	}

	// decode value
	return json.Unmarshal(r.bytes, value)
}

func Call(conn Conn, name string, args ...any) error {
	payload := []any{name}
	if args != nil && len(args) > 0 {
		payload = append(payload, args...)
	}
	return conn.Encode(payload)
}

func Await(conn Conn) (err error) {
	var r interface{}
	err = conn.Decode(&r)
	return
}

func Command(conn Conn, method string, args ...any) (err error) {
	if err = Call(conn, method, args...); err != nil {
		return
	}
	if err = Await(conn); errors.Is(err, io.EOF) {
		err = nil
	}
	return
}

func Query[R any](conn Conn, method string, args ...any) (r R, err error) {
	if err = Call(conn, method, args...); err == nil {
		r, err = Decode[R](conn)
	}
	if err != nil && err.Error() == "EOF" {
		err = nil
	}
	return
}

func Decode[R any](conn Conn) (r R, err error) {
	err = conn.Decode(&r)
	return
}

func Subscribe[R any](conn Conn, method string, args ...any) (c <-chan R, err error) {
	if err = Call(conn, method, args...); err != nil {
		return
	}
	cc := make(chan R)
	go func() {
		defer close(cc)
		var r R
		for {
			if err = conn.Decode(&r); err != nil {
				return
			}
			cc <- r
		}
	}()
	c = cc
	return
}
