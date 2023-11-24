package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"reflect"
	"strings"
)

type Server[T any] struct {
	Ctx     context.Context
	Accept  func(query *astral.QueryData) (*astral.Conn, error)
	Handler func(conn *Conn) T
}

func (s Server[T]) Run() (err error) {
	srvName := fmt.Sprintf("%v", s.Handler(nil))
	listener, err := astral.Register(srvName)
	if err != nil {
		return
	}
	defer listener.Close()
	if s.Accept == nil {
		s.Accept = acceptAll
	}
	if s.Ctx == nil {
		s.Ctx = context.Background()
	}
	queryCh := listener.QueryCh()
	for {
		select {
		case <-s.Ctx.Done():
			return
		case data := <-queryCh:
			go handleQuery(s.Ctx, data, s.Accept,
				func(conn *Conn) any {
					return s.Handler(conn)
				},
			)
		}
	}
}

func acceptAll(q *astral.QueryData) (*astral.Conn, error) {
	return q.Accept()
}

func handleQuery(
	ctx context.Context,
	data *astral.QueryData,
	accept func(query *astral.QueryData) (*astral.Conn, error),
	service func(*Conn) any,
) {
	// accept conn
	conn, err := accept(data)
	if err != nil {
		return
	}
	defer conn.Close()

	// create service
	rpc := NewConn(ctx, conn)
	defer rpc.Close()
	srv := service(rpc)
	if srv == nil {
		return
	}

	// try handle conn
	switch srv.(type) {
	case error:
		err = fmt.Errorf("cannot invoke service: %v", srv)
	default:
		err = handleConn(rpc, srv)
	}
	if err != nil {
		_ = rpc.Encode(err)
	}
	return
}

func handleConn(rpc *Conn, srv any) (err error) {
	for rpc.Ctx.Err() == nil {

		// decode method
		m := method{}
		if err = rpc.Decode(&m); err != nil {
			return
		}

		// invoke method
		r, e := invoke(srv, m)
		if e != nil {
			r = e
		}

		// handle chan result
		v := reflect.ValueOf(r)
		if r != nil && v.Kind() == reflect.Chan {
			sel := []reflect.SelectCase{{
				Dir:  reflect.SelectRecv,
				Chan: v,
			}}
		loop:
			for {
				select {
				case <-rpc.Ctx.Done():
					break loop
				default:
					_, recv, ok := reflect.Select(sel)
					if !ok {
						break loop
					}
					r = recv.Interface()
					if err = rpc.Encode(r); err != nil {
						break loop
					}
				}
			}
			//v.Close()
			return
		}

		// handle normal results
		_ = rpc.Encode(r)
	}
	return
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
