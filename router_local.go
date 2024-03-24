package jrpc

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
)

func NewRouterModule(router *Router, node node.Node) *RouterModule {
	return &RouterModule{
		Router: router,
		node:   node,
	}
}

type RouterModule struct {
	*Router
	node node.Node
}

func (r *RouterModule) Setup(router *Router) *RouterModule {
	r.Router = router
	return r
}

func (r *RouterModule) Run(ctx context.Context) (err error) {
	for _, route := range r.routes {
		rr := *r
		rr.port = route
		go func(r RouterModule) {
			if err := r.registerRoute(ctx, r.port); err != nil {
				panic(err)
			}
		}(rr)
	}
	return
}

func (r *RouterModule) registerRoute(ctx context.Context, route string) (err error) {
	if err = r.node.LocalRouter().AddRoute(route, r); err != nil {
		return
	}
	go func() {
		<-ctx.Done()
		_ = r.node.LocalRouter().RemoveRoute(route)
	}()
	return
}

func (r *RouterModule) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, _ net.Hints) (net.SecureWriteCloser, error) {
	q := query.Query()
	return net.Accept(query, caller, func(conn net.SecureConn) {
		rpc := *NewFlow(conn)
		r.Route(ctx, q, rpc)
	})
}
