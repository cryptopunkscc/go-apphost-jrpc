package rpc

import (
	"encoding/json"
	"io"
	"log"
)

type ConnLogger struct {
	io.ReadWriteCloser
	*log.Logger
}

func NewConnLogger(conn io.ReadWriteCloser, logger *log.Logger) *ConnLogger {
	return &ConnLogger{
		ReadWriteCloser: conn,
		Logger:          logger,
	}
}

func (cl *ConnLogger) Read(b []byte) (n int, err error) {
	n, err = cl.ReadWriteCloser.Read(b)
	if n > 0 {
		cl.Print("< ", string(b[:n]))
	}
	return
}

func (cl *ConnLogger) Write(b []byte) (n int, err error) {
	n, err = cl.ReadWriteCloser.Write(b)
	if n > 0 {
		cl.Print("> ", string(b[:n]))
	}
	return
}

func (conn *Conn) WithLogger(logger *log.Logger) *Conn {
	connLogger := NewConnLogger(conn.ReadWriteCloser, logger)
	conn.enc = json.NewEncoder(connLogger)
	conn.dec = json.NewDecoder(connLogger)
	return conn
}
