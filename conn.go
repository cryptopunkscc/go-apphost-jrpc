package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
)

type Conn struct {
	astral.Conn
	Ctx      context.Context
	closeCtx context.CancelFunc
	enc      *json.Encoder
	dec      *json.Decoder
}

func (conn *Conn) Close() error {
	conn.closeCtx()
	return conn.Conn.Close()
}

func NewConn(ctx context.Context, conn *astral.Conn) (c *Conn) {
	c = &Conn{
		Conn: *conn,
		enc:  json.NewEncoder(conn),
		dec:  json.NewDecoder(conn),
	}
	c.Ctx, c.closeCtx = context.WithCancel(ctx)
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
