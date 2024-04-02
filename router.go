package jrpc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
	"log"
	"reflect"
	"strings"
	"unicode"
)

type Router struct {
	logger        *log.Logger
	registry      *Registry[*Caller]
	routes        []string
	port          string
	env           []any
	args          string
	rpc           *Flow
	registerRoute func(ctx context.Context, route string) error
}

func NewRouter(port string) *Router {
	return &Router{
		port:     port,
		registry: NewRegistry[*Caller](),
	}
}

func (r *Router) Routes(routes ...string) *Router {
	r.routes = append(r.routes, routes...)
	return r
}

func (r *Router) Logger(logger *log.Logger) *Router {
	r.logger = logger
	return r
}

func (r *Router) With(env ...any) *Router {
	rr := *r
	rr.env = append(r.env, env...)
	return &rr
}

func (r *Router) Caller(caller *Caller) *Router {
	r.registry.Add(caller.name, caller)
	return r
}

func (r *Router) Func(name string, function any) *Router {
	return r.Caller(NewCaller(name).Func(function))
}

func (r *Router) Interface(srv any) *Router {
	t := reflect.TypeOf(srv)
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if !m.IsExported() {
			continue
		}
		f := m.Func.Interface()
		runes := []rune(m.Name)
		runes[0] = unicode.ToLower(runes[0])
		name := string(runes)
		if strings.HasSuffix(name, "Auth") {
			name = name[:len(name)-4] + "!"
		}
		r.Caller(NewCaller(name).With(srv).Func(f))
	}
	return r
}

func (r *Router) Run(ctx context.Context) (err error) {
	r.registerApi()
	if len(r.routes) == 0 {
		go func(r Router, route string) {
			if err = r.registerRoute(ctx, route); err != nil {
				panic(err)
			}
		}(*r, r.port)
	}
	for _, cmd := range r.routes {
		rr := *r
		f := "%s.%s"
		if cmd == "*" {
			f = "%s%s"
		}
		go func(r Router, route string) {
			if err = r.registerRoute(ctx, route); err != nil {
				panic(err)
			}
		}(rr, fmt.Sprintf(f, r.port, cmd))
	}
	return
}

func (r *Router) registerApi() *Router {
	var arr []string
	for s := range r.registry.All() {
		if strings.HasSuffix(s, "!") {
			continue
		}
		arr = append(arr, s)
	}
	r.Func("api", func() []string { return arr })
	return r
}

func (r *Router) Query(query string) *Router {
	rr := *r
	q := strings.TrimPrefix(query, r.port)
	q = strings.TrimPrefix(q, ".")
	rr.registry, rr.args = r.registry.Unfold(q)
	rr.port = strings.TrimSuffix(q, rr.args)
	if rr.port == "" {
		rr.port = r.port
	}
	if rr.args == q && q != "" {
		// nothing was unfolded query cannot be handled
		rr.registry = NewRegistry[*Caller]()
	}
	if rr.args == "\n" {
		rr.args = ""
	}
	if rr.rpc == nil {
		rr.rpc = NewFlow(nil)
		rr.rpc.ByteScannerReader = NewByteScannerReader(strings.NewReader(""))
	}
	if rr.rpc != nil {
		rr.rpc.Append([]byte(rr.args))
	}
	return &rr
}

func (r *Router) Authorize(ctx context.Context, query any) bool {
	res, _ := r.Query("!").With(ctx, query).Call()
	return len(res) > 0 && res[0] == false
}

func (r *Router) Handle(ctx context.Context, query any, remoteId id.Identity, conn io.ReadWriteCloser) (err error) {
	r.Conn(conn)
	rr := *r
	scanner := bufio.NewScanner(conn)
	var result []any
	for {
		switch {
		case !rr.registry.IsEmpty():
			// caller found
			result, err = rr.With(ctx, query, remoteId, rr.rpc).Call()
			if !rr.respond(ctx, err, result...) {
				return
			}

		case !rr.rpc.IsEmpty():
			// caller not found and there are unhandled data in rpc buffer
			if !rr.respond(ctx, MalformedRequest) {
				return
			}
		}
		r.rpc.Clear()
		if !scanner.Scan() {
			return
		}
		rr = *r.Query(scanner.Text() + "\n")
	}
}

var MalformedRequest = errors.New("malformed request")

func (r *Router) Conn(conn io.ReadWriteCloser) *Router {
	r.rpc = NewFlow(conn)
	if r.logger != nil {
		r.rpc.Logger(r.logger)
	}
	if r.args != "" {
		r.rpc.Append([]byte(r.args))
	}
	return r
}

func (r *Router) Call() (result []any, err error) {
	if r.registry.IsEmpty() {
		return nil, fmt.Errorf("route not found for query %s%s ", r.port, r.args)
	}
	result, err = r.registry.Get().With(r.env...).Call(r.rpc)
	return
}

func (r *Router) respond(ctx context.Context, err error, result ...any) (b bool) {

	// eof / error / empty / arr
	switch {
	case errors.Is(err, io.EOF):
		return false
	case err != nil:
		return r.rpc.Encode(err) == nil
	case len(result) == 0:
		return r.rpc.Encode(EmptyResponse) == nil
	case len(result) > 1:
		return r.rpc.Encode(result) == nil
	}

	res := result[len(result)-1]
	v := reflect.ValueOf(res)

	// single
	if v.Kind() != reflect.Chan {
		return r.rpc.Encode(res) == nil
	}

	// channel
	sel := []reflect.SelectCase{{Dir: reflect.SelectRecv, Chan: v}}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if _, v, b = reflect.Select(sel); !b {
				return
			}
			res = v.Interface()
			if err = r.rpc.Encode(res); err != nil {
				return false
			}
		}
	}
}

var EmptyResponse = struct{}{}
