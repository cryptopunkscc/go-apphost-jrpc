package jrpc

import (
	"reflect"
)

type Caller struct {
	env     []any
	decoder argsDecoders
	f       reflect.Value
}

func NewCaller(env ...any) (c *Caller) {
	c = &Caller{env: env}
	c.Decoder(NewJsonArgsDecoder())
	c.Decoder(NewClirArgsDecoder())
	return
}

func (exec *Caller) With(env ...any) *Caller {
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

func (exec *Caller) Decoder(decoders ...ArgsDecoder) *Caller {
	return exec.Decoders(decoders)
}

func (exec *Caller) Decoders(decoders []ArgsDecoder) *Caller {
	exec.decoder.Append(decoders)
	return exec
}

func (exec *Caller) Call(args ByteScannerReader) (result []any, err error) {
	values := make([]reflect.Value, 0)

	for _, a := range exec.env {
		values = append(values, reflect.ValueOf(a))
	}

	if values, err = exec.decodeIn(values, args); err != nil {
		return
	}

	values = exec.f.Call(values)

	for _, value := range values {
		result = append(result, value.Interface())
	}
	return
}

func (exec *Caller) decodeIn(initial []reflect.Value, args ByteScannerReader) (values []reflect.Value, err error) {
	t := exec.f.Type()

	for i := 0; i < t.NumIn() && len(initial) > 0; i++ {
		for len(initial) > 0 && !initial[0].Type().AssignableTo(t.In(i)) {
			initial = initial[1:]
			continue
		}
		if len(initial) > 0 {
			values = append(values, initial[0])
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
		if err = exec.decoder.Decode(args, decoded); err != nil {
			return
		}
	}

	return
}
