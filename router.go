package jrpc

import (
	"context"
	"errors"
	"io"
)

type Router struct {
	port     string
	registry Registry[*Caller]
}

func (r Router) Route(ctx context.Context, query string, conn io.ReadWriteCloser) {
	args, caller := r.registry.Unfold(query)
	rpc := NewFlow(conn)
	defer conn.Close()
	caller = caller.New(ctx, rpc)

	var result []any
	var err error

	if args != "" {
		// call with query args
		result, err = caller.Call(args)
	} else {

	}

	// trim end nil values
	for n := len(result) - 1; n > 0 && result[n] == nil; n-- {
		result = result[0:n]
	}

	// handle call results
	if err == nil && len(result) > 0 {
		if err, _ = result[len(result)-1].(error); err == nil {
			_ = rpc.Encode(result)
		}
	}

	// encode error if not EOF
	if err != nil && !errors.Is(err, io.EOF) {
		_ = rpc.Encode(err)
	}
}
