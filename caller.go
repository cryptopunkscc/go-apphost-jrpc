package jrpc

import (
	"reflect"
)

type Caller struct {
	name    string
	env     []any
	decoder argsDecoders
	f       reflect.Value
	args    []reflect.Value
}

func NewCaller(name string) (c *Caller) {
	c = &Caller{name: name}
	c.Decoder(NewJsonArgsDecoder(), NewClirArgsDecoder())
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

func (exec *Caller) Call(args ByteScannerReader) (out []any, err error) {
	values, err := exec.call(args)
	if err != nil {
		return
	}
	out = formatOut(values)
	return
}

func (exec *Caller) call(args ByteScannerReader) (out []reflect.Value, err error) {
	values, err := exec.decodeIn(args)
	if err != nil {
		return
	}
	values = exec.f.Call(values)
	err = handleError(values)
	if err != nil {
		return
	}
	out, err = exec.runNested(values, args)
	return
}

func (exec *Caller) decodeIn(args ByteScannerReader) (values []reflect.Value, err error) {
	var initial []reflect.Value
	for _, a := range exec.env {
		initial = append(initial, reflect.ValueOf(a))
	}

	t := exec.f.Type()

	for i := 0; i < t.NumIn() && len(initial) > 0; i++ {
		for len(initial) > 0 && !initial[0].Type().AssignableTo(t.In(i)) {
			initial = initial[1:]
			continue
		}
		if len(initial) > 0 {
			values = append(values, initial[0])
			initial = initial[1:]
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

func handleError(values []reflect.Value) (err error) {
	if len(values) == 0 {
		return
	}
	last := values[len(values)-1]
	if last.Type().Implements(errorInterface) {
		if i := last.Interface(); i != nil {
			err, _ = i.(error)
		}
	}
	return
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func (exec *Caller) runNested(values []reflect.Value, args ByteScannerReader) (r []reflect.Value, err error) {
	for _, value := range values {
		if value.Kind() == reflect.Func {
			e := *(&exec)
			e.f = value
			var rr []reflect.Value
			if rr, err = e.call(args); err != nil {
				return
			}
			r = append(r, rr...)
			continue
		}
		r = append(r, value)
	}
	return
}

func formatOut(values []reflect.Value) (result []any) {
	// filter out error for values
	for _, value := range values {
		result = append(result, value.Interface())
	}

	// trim nil values
	for n := len(result) - 1; n > 0 && result[n] == nil; n-- {
		result = result[0:n]
	}
	return
}
