package rpc

import (
	"errors"
	"io"
	"log"
)

type Conn interface {
	io.ReadWriteCloser
	WithLogger(logger *log.Logger) Conn
	Copy() Conn
	Encode(value any) (err error)
	Decode(value any) (err error)
	Flush()
}

func Call(conn Conn, name string, args ...any) error {
	payload := []any{name}
	if args != nil && len(args) > 0 {
		payload = append(payload, args...)
	}
	return conn.Encode(payload)
}

func Decode[R any](conn Conn) (r R, err error) {
	err = conn.Decode(&r)
	return
}

func Await(conn Conn) (err error) {
	var r interface{}
	err = conn.Decode(&r)
	return
}

func Command(conn Conn, method string, args ...any) (err error) {
	c := conn.Copy()
	defer conn.Flush()
	if err = Call(c, method, args...); err != nil {
		return
	}
	if err = Await(c); errors.Is(err, io.EOF) {
		err = nil
	}
	return
}

func Query[R any](conn Conn, method string, args ...any) (r R, err error) {
	conn = conn.Copy()
	defer conn.Flush()
	if err = Call(conn, method, args...); err == nil {
		r, err = Decode[R](conn)
	}
	if err != nil && err.Error() == "EOF" {
		err = nil
	}
	return
}

func Subscribe[R any](conn Conn, method string, args ...any) (c <-chan R, err error) {
	conn = conn.Copy()
	if err = Call(conn, method, args...); err != nil {
		return
	}
	cc := make(chan R)
	go func() {
		defer close(cc)
		defer conn.Flush()
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
