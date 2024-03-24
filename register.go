package jrpc

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
	"log"
)

type Server[T any] struct {
	Accept  func(query *astral.QueryData) (io.ReadWriteCloser, error)
	Handler func(ctx context.Context, conn Conn) T
}

func NewServer[T any](handler func(ctx context.Context, conn Conn) T) *Server[T] {
	return &Server[T]{Handler: handler, Accept: acceptAll}
}

func (s Server[T]) Middleware(accept func(query *astral.QueryData) (io.ReadWriteCloser, error)) Server[T] {
	s.Accept = accept
	return s
}

func (s Server[T]) Logger(logger *log.Logger) Server[T] {
	accept := s.Accept
	s.Accept = func(query *astral.QueryData) (conn io.ReadWriteCloser, err error) {
		if conn, err = accept(query); err == nil {
			conn = NewConnLogger(conn, logger)
		}
		return
	}
	return s
}

func (s Server[T]) Run(ctx context.Context) (err error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
	}
	srvName := fmt.Sprintf("%v*", s.Handler(ctx, nil))
	listener, err := astral.Register(srvName)
	if err != nil {
		return
	}
	defer listener.Close()
	if s.Accept == nil {
		s.Accept = acceptAll
	}
	queryCh := listener.QueryCh()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-queryCh:
			if data == nil {
				return
			}
			go handleQuery(ctx, data, s.Accept,
				func(ctx context.Context, conn Conn) any {
					return s.Handler(ctx, conn)
				},
			)
		}
	}
}

func acceptAll(q *astral.QueryData) (io.ReadWriteCloser, error) {
	return q.Accept()
}

func handleQuery(
	ctx context.Context,
	data *astral.QueryData,
	accept func(query *astral.QueryData) (io.ReadWriteCloser, error),
	service func(ctx context.Context, rpc Conn) any,
) {
	// accept conn
	conn, err := accept(data)
	if err != nil {
		return
	}
	defer conn.Close()
	log.Println("handling query:", data.Query())

	Handle(ctx, data.Query(), conn, service)
}
