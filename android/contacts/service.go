package contacts

import (
	"context"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
)

type service struct {
	*Module
	parent context.Context
	ctx    context.Context
	conn   rpc.Conn
}
