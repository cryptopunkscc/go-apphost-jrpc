package jrpc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"reflect"
	"strings"
	"unicode"
)

type Router struct {
	port     string
	registry *Registry[*Caller]
	routes   []string
}

func NewRouter(routes ...string) *Router {
	return &Router{
		registry: NewRegistry[*Caller](),
		routes:   routes,
	}
}

func (r *Router) Func(name string, function any) *Router {
	r.registry.Add(name, NewCaller().Func(function))
	return r
}

func (r *Router) Interface(srv any) *Router {
	t := reflect.TypeOf(srv)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if !m.IsExported() {
			continue
		}
		f := m.Func.Interface()
		name := []rune(m.Name)
		name[0] = unicode.ToLower(name[0])
		r.registry.Add(string(name), NewCaller(srv).Func(f))
	}
	return r
}

func (r *Router) Setup() *Router {
	r.Func("api", r.respondApi)
	return r
}

func (r *Router) Route(ctx context.Context, query string, rpc Flow) {
	query = strings.TrimPrefix(query, strings.TrimSuffix(r.port, "*"))
	scanner := bufio.NewScanner(rpc)

	for ctx.Err() == nil {

		if query == "" {
			if !scanner.Scan() {
				return
			}
			query = scanner.Text()
		}

		args, caller := r.registry.Unfold(query)
		if caller == nil {
			respond(ctx, rpc, nil, fmt.Errorf("route not found: %s", query))
			return
		}

		caller = caller.With(ctx, rpc)

		var argsReader ByteScannerReader = rpc
		if args != "" {
			argsReader = NewByteScannerReader(strings.NewReader(args))
			args = ""
		}

		result, err := caller.Call(argsReader)

		if !respond(ctx, rpc, result, err) {
			return
		}

		query = ""
	}
}

func (r *Router) respondApi() (arr []string) {
	m := make(map[string]*Caller)
	r.registry.All(nil, m)
	for s := range m {
		arr = append(arr, s)
	}
	return
}

func respond(ctx context.Context, rpc Flow, result []any, err error) (b bool) {
	// trim end nil values
	for n := len(result) - 1; n > 0 && result[n] == nil; n-- {
		result = result[0:n]
	}

	if err == nil && len(result) > 0 {
		last := result[len(result)-1]

		if err, _ = last.(error); err == nil {
			v := reflect.ValueOf(last)
			switch v.Kind() {

			// channel response
			case reflect.Chan:
				sel := []reflect.SelectCase{{
					Dir:  reflect.SelectRecv,
					Chan: v,
				}}
				for {
					select {
					case <-ctx.Done():
						return
					default:
						_, recv, ok := reflect.Select(sel)
						if !ok {
							return
						}
						r := recv.Interface()
						if err = rpc.Encode(r); err != nil {
							return
						}
					}
				}

			// default response
			default:
				if len(result) == 1 {
					err = rpc.Encode(last)
				} else {
					err = rpc.Encode(result)
				}
			}
		}
	} else {
		_ = rpc.Encode(struct{}{})
	}

	b = !errors.Is(err, io.EOF)
	if err != nil && b {
		// error response
		_ = rpc.Encode(err)
	}
	return
}

// =================================================================================================

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

// =================================================================================================

func (r *Router) Serve(ctx context.Context, node node.Node) error {
	err := NewRouterModule(r, node).Run(ctx)
	return err
}

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
