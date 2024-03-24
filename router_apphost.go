package jrpc

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

func (r *Router) Register(ctx context.Context, accept ...ApphostMiddleware) {
	for _, route := range r.routes {
		rr := *r
		rr.port = route
		go func(r Router) {
			if err := r.register(ctx, accept); err != nil {
				panic(err)
			}
		}(rr)
	}
	return
}

func (r *Router) register(ctx context.Context, accept []ApphostMiddleware) (err error) {
	listener, err := astral.Register(r.port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	done := ctx.Done()
	queries := listener.QueryCh()
	ack := AcceptAllMiddleware
	if len(accept) > 0 {
		ack = accept[0]
	}
	for {
		select {
		case <-done:
			return
		case q := <-queries:
			go func(q *astral.QueryData) {
				var conn *Flow
				if conn, err = ack(q, nil); err != nil {
					return
				}
				r.Route(ctx, q.Query(), *conn)
				_ = conn.Close()
			}(q)
		}
	}
}

func AcceptAllMiddleware(data *astral.QueryData, _ *Flow) (f *Flow, err error) {
	conn, err := data.Accept()
	if err != nil {
		return
	}
	f = NewFlow(conn)
	return
}

type ApphostMiddleware func(data *astral.QueryData, f *Flow) (*Flow, error)
