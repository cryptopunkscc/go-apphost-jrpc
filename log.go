package rpc

import (
	"encoding/json"
	"log"
	"net"
)

type ConnLogger struct {
	net.Conn
	*log.Logger
}

func NewConnLogger(conn net.Conn, logger *log.Logger) *ConnLogger {
	return &ConnLogger{
		Conn:   conn,
		Logger: logger,
	}
}

func (cl *ConnLogger) Read(b []byte) (n int, err error) {
	n, err = cl.Conn.Read(b)
	if n > 0 {
		cl.Print("< ", string(b[:n]))
	}
	return
}

func (cl *ConnLogger) Write(b []byte) (n int, err error) {
	n, err = cl.Conn.Write(b)
	if n > 0 {
		cl.Print("> ", string(b[:n]))
	}
	return
}

func (conn *Conn) WithLogger(logger *log.Logger) *Conn {
	connLogger := NewConnLogger(conn.Conn, logger)
	conn.enc = json.NewEncoder(connLogger)
	conn.dec = json.NewDecoder(connLogger)
	return conn
}
