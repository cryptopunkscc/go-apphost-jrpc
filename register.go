package jrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"io"
	"log"
	"reflect"
	"strings"
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

func Handle(
	ctx context.Context,
	data string,
	conn io.ReadWriteCloser,
	service func(ctx context.Context, rpc Conn) any,
) {
	// create service
	rpc := NewFlow(conn)
	ctx, closeCtx := context.WithCancel(ctx)
	defer func() {
		closeCtx()
		_ = rpc.Close()
	}()
	srv := service(ctx, rpc)
	if srv == nil {
		return
	}

	// try handle conn
	var err error
	switch srv.(type) {
	case error:
		err = fmt.Errorf("cannot invoke service: %v", srv)
	default:
		err = handleConn(ctx, data, rpc, srv)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		_ = rpc.Encode(err)
	}
	return
}

func handleConn(ctx context.Context, query string, rpc Conn, srv any) error {
	for ctx.Err() == nil {
		m := method{}
		if mm, _ := strings.CutPrefix(query, fmt.Sprint(srv)); mm != "" {
			// decode method from query
			if err := json.Unmarshal([]byte(mm), &m); err != nil {
				return err
			}
		} else {
			// decode method from connection
			if err := rpc.Decode(&m); err != nil {
				return err
			}
		}

		// invoke method
		r, err := invoke(srv, m)
		if err != nil {
			r = err
		}

		// handle chan result
		if b := handleChannel(ctx, rpc, r); b {
			return nil
		}

		// handle normal results
		if err = rpc.Encode(r); err != nil {
			return nil
		}
	}
	return nil
}

func handleChannel(ctx context.Context, rpc Conn, r any) (b bool) {
	v := reflect.ValueOf(r)
	if r == nil || v.Kind() != reflect.Chan {
		return
	}

	b = true
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
			if err := rpc.Encode(r); err != nil {
				return
			}
		}
	}
}

func invoke(service any, method method) (a any, err error) {
	st := reflect.TypeOf(service)

	// return list of methods
	if method.name == "api" {
		s := st.NumMethod()
		ms := make([]string, s)
		for i := 0; i < s; i++ {
			m := st.Method(i)
			if m.IsExported() {
				n := m.Name
				ms[i] = strings.ToLower(n[:1]) + n[1:]
			}
		}
		a = ms
		return
	}

	// find method to call
	for i := 0; i < st.NumMethod(); i++ {
		m := st.Method(i)

		if m.IsExported() && strings.EqualFold(m.Name, method.name) {

			if len(method.raw) < m.Func.Type().NumIn()-1 {
				err = fmt.Errorf("%v: %s Call with too few input arguments", service, method.name)
				return
			}

			// decode arguments
			var values = []reflect.Value{reflect.ValueOf(service)}

			for i := 0; i < len(method.raw); i++ {
				at := m.Type.In(i + 1)
				av := reflect.New(at)
				ai := av.Interface()
				if err = json.Unmarshal(method.raw[i], ai); err != nil {
					err = fmt.Errorf("cannot unmarshal param %d: %v", i, err)
					return
				}
				values = append(values, av.Elem())
			}

			// call method
			res := m.Func.Call(values)

			// handle result
			switch len(res) {
			case 1:
				switch e := res[0].Interface().(type) {
				case error:
					err = e
				default:
					a = e
				}
			case 2:
				if r := res[0].Interface(); r != nil {
					a = r
				}
				switch e := res[0].Interface().(type) {
				case error:
					err = e
				}
			}
			return
		}
	}

	// method not found
	err = fmt.Errorf("invalid method %v", method.name)
	return
}
