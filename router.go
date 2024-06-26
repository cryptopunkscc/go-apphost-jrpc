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
	env           []any
	port          string
	query         string
	args          string
	rpc           *Flow
	registerRoute func(ctx context.Context, route string) error
}

var ErrMalformedRequest = errors.New("malformed request")
var ErrUnauthorized = errors.New("unauthorized")

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

func (r *Router) Command(cmd string) *Router { return r.shift(cmd, true) }
func (r *Router) Query(query string) *Router { return r.shift(query, false) }

func (r *Router) shift(query string, force bool) *Router {
	rr := *r
	rr.Conn(rr.rpc)
	rr.query = strings.TrimPrefix(query, r.port)
	rr.query = strings.TrimPrefix(rr.query, ".")
	rr.registry, rr.args = r.registry.Unfold(rr.query)
	rr.port = strings.TrimSuffix(rr.query, rr.args)

	if rr.args == rr.query && rr.query != "" && force {
		// nothing was unfolded query cannot be handled
		rr.registry = NewRegistry[*Caller]()
		return &rr
	}
	if rr.port == "" {
		rr.port = r.port
	}
	if rr.args == "\n" {
		rr.args = ""
	} else {
		rr.args, _ = strings.CutPrefix(rr.args, "?")
	}
	return &rr
}

func (r *Router) Authorize(ctx context.Context, query any) bool {
	res, _ := r.Command("!").With(ctx, query).Call()
	return len(res) == 0 || res[0] != false
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

		case rr.args != "":
			// caller not found and there are unhandled data in rpc buffer
			if !rr.respond(ctx, ErrMalformedRequest) {
				return
			}
		}

		r.rpc.Clear()
		if !scanner.Scan() {
			return
		}
		rr = *r.Query(scanner.Text())

		//authorize if registry changed
		if rr.registry.value != r.registry.value && !rr.Authorize(ctx, query) {
			if !rr.respond(ctx, ErrUnauthorized) {
				return
			}
		}
	}
}

func (r *Router) Conn(conn io.ReadWriteCloser) *Router {
	r.rpc = NewFlow(conn)
	if r.logger != nil {
		r.rpc.Logger(r.logger)
	}
	return r
}

func (r *Router) Call() (result []any, err error) {
	r.loadArgs()
	if r.registry.IsEmpty() {
		return nil, fmt.Errorf("route not found for query %s%s ", r.port, r.args)
	}
	result, err = r.registry.Get().With(r.env...).Call(r.rpc)
	return
}

func (r *Router) loadArgs() {
	if r.rpc != nil && r.args != "" {
		if !strings.HasSuffix(r.args, "\n") {
			r.args += "\n"
		}
		r.rpc.Append([]byte(r.args))
	}
	r.args = ""
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
