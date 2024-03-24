package jrpc

import (
	"reflect"
)

type Caller struct {
	env      []any
	decoders []ArgsDecoder
	f        reflect.Value
}

func NewCaller(env ...any) (c *Caller) {
	c = &Caller{env: env}
	c.Decoder(NewJsonArgsDecoder())
	c.Decoder(NewClirArgsDecoder())
	return
}

func (exec *Caller) New(env ...any) *Caller {
	c := *exec
	c.env = append(c.env, env...)
	return &c
}

func (exec *Caller) Func(function any) *Caller {
	if exec.f = reflect.ValueOf(function); exec.f.Kind() != reflect.Func {
		panic("argument must be a function")
	}
	return exec
}

func (exec *Caller) Decoder(decoder ...ArgsDecoder) *Caller {
	exec.decoders = append(exec.decoders, decoder...)
	return exec
}

func (exec *Caller) Decoders(decoder []ArgsDecoder) *Caller {
	exec.decoders = append(exec.decoders, decoder...)
	return exec
}

func (exec *Caller) Call(args string) (result []any, err error) {
	values := make([]reflect.Value, 0)

	for _, a := range exec.env {
		values = append(values, reflect.ValueOf(a))
	}

	if values, err = exec.args(exec.f.Type(), values, args); err != nil {
		return
	}

	values = exec.f.Call(values)

	for _, value := range values {
		result = append(result, value.Interface())
	}
	return
}

func (exec *Caller) args(t reflect.Type, env []reflect.Value, arg string) (values []reflect.Value, err error) {
	//values = env

	for i := 0; i < t.NumIn() && len(env) > 0; i++ {
		for len(env) > 0 && !env[0].Type().AssignableTo(t.In(i)) {
			env = env[1:]
			continue
		}
		if len(env) > 0 {
			values = append(values, env[0])
		}
	}

	var decoded []any

	for i := len(values); i < t.NumIn(); i++ {
		at := t.In(i)
		av := reflect.New(at)
		values = append(values, av.Elem())
		decoded = append(decoded, av.Interface())
	}

	if len(decoded) > 0 {
		bytes := []byte(arg)
		for _, d := range exec.decoders {
			if d.Test(bytes) {
				if err = d.Unmarshal([]byte(arg), decoded); err != nil {
					return
				}
				break
			}
		}
	}

	return
}
