package jrpc

import (
	"encoding/json"
	"errors"
	"github.com/leaanthony/clir"
	"io"
	"reflect"
	"strings"
)

type ArgsDecoder interface {
	Test([]byte) bool
	TestS(stream io.ByteScanner) bool
	Begin() []rune
	Unmarshal(bytes []byte, args []any) error
	Decode(conn *ReadScanner, args []any) error
}

type argsDecoders struct {
	decoders []ArgsDecoder
}

func (a argsDecoders) TestS(scan io.ByteScanner) bool {
	for _, decoder := range a.decoders {
		if decoder.TestS(scan) {
			return true
		}
	}
	return false
}

func (a argsDecoders) Decode(conn *ReadScanner, args []any) error {
	for _, decoder := range a.decoders {
		if decoder.TestS(conn) {
			return decoder.Decode(conn, args)
		}
	}
	return errors.New("unknown format")
}

func NewArgsDecoders(decoders ...ArgsDecoder) ArgsDecoder {
	return &argsDecoders{decoders}
}

func (a argsDecoders) Test(bytes []byte) bool {
	for _, decoder := range a.decoders {
		if decoder.Test(bytes) {
			return true
		}
	}
	return false
}

func (a argsDecoders) Begin() (r []rune) {
	for _, decoder := range a.decoders {
		r = append(r, decoder.Begin()...)
	}
	return
}

func (a argsDecoders) Unmarshal(bytes []byte, args []any) error {
	for _, decoder := range a.decoders {
		if decoder.Test(bytes) {
			return decoder.Unmarshal(bytes, args)
		}
	}
	return errors.New("unknown format")
}

type jsonArgsDecoder struct {
	begin []rune
	end   []rune
}

func NewJsonArgsDecoder() ArgsDecoder {
	return &jsonArgsDecoder{}
}

func (d jsonArgsDecoder) Begin() []rune {
	return []rune{'[', '{'}
}

func (d jsonArgsDecoder) Test(b []byte) bool {
	return b[0] == '[' || b[0] == '{'
}

func (d jsonArgsDecoder) TestS(scan io.ByteScanner) bool {
	b, err := scan.ReadByte()
	if err != nil {
		return false
	}
	_ = scan.UnreadByte()
	return b == '[' || b == '{'
}

func (d jsonArgsDecoder) Unmarshal(bytes []byte, args []any) error {
	if len(args) == 1 {
		// unmarshal struct payload to as first arg
		if bytes[0] == '{' {
			return json.Unmarshal(bytes, &args[0])
		}

		// unmarshal positional arguments to struct
		x := reflect.ValueOf(args[0]).Elem()
		if x.Kind() == reflect.Struct {
			arr := make([]any, x.NumField())
			for i := 0; i < len(arr); i++ {
				arr[i] = reflect.New(x.Field(i).Type()).Interface()
			}
			if err := json.Unmarshal(bytes, &arr); err != nil {
				return err
			}
			for i := 0; i < len(arr); i++ {
				x.Field(i).Set(reflect.ValueOf(arr[i]).Elem())
			}
			return nil
		}
	}

	// perform default unmarshal
	return json.Unmarshal(bytes, &args)
}

func (d jsonArgsDecoder) Decode(conn *ReadScanner, args []any) error {
	jd := json.NewDecoder(conn)
	if len(args) == 1 {
		// unmarshal struct payload to as first arg
		c, err := conn.ReadByte()
		_ = conn.UnreadByte()
		if err != nil {
			return err
		}
		if c == byte('{') {
			return jd.Decode(&args[0])
		}

		// unmarshal positional arguments to struct
		x := reflect.ValueOf(args[0]).Elem()
		if x.Kind() == reflect.Struct {
			arr := make([]any, x.NumField())
			for i := 0; i < len(arr); i++ {
				arr[i] = reflect.New(x.Field(i).Type()).Interface()
			}
			if err := jd.Decode(&arr); err != nil {
				return err
			}
			for i := 0; i < len(arr); i++ {
				x.Field(i).Set(reflect.ValueOf(arr[i]).Elem())
			}
			return nil
		}
	}

	// perform default unmarshal
	return jd.Decode(&args)
}

type clirArgsDecoder struct{}

func (d clirArgsDecoder) Decode(conn *ReadScanner, args []any) (err error) {
	var b byte
	var n = 0
	for b, err = conn.ReadByte(); b != '\n'; n++ {
		if err != nil {
			return
		}
	}
	for i := 0; i < n; i++ {
		_ = conn.UnreadByte()
	}
	bytes := make([]byte, n)
	if _, err = conn.Read(bytes); err != nil {
		return
	}
	err = d.Unmarshal(bytes, args)
	return
}

func NewClirArgsDecoder() ArgsDecoder { return &clirArgsDecoder{} }

func (d clirArgsDecoder) Begin() []rune {
	return []rune{'$', ' '}
}

func (d clirArgsDecoder) Test(b []byte) bool {
	return b[0] == '$' || b[0] == ' '
}

func (d clirArgsDecoder) TestS(scan io.ByteScanner) bool {
	b, err := scan.ReadByte()
	_ = scan.UnreadByte()
	if err != nil {
		return false
	}
	return b == '$' || b == ' '
}

func (d clirArgsDecoder) Unmarshal(bytes []byte, args []any) error {
	f := strings.Fields(string(bytes[1:]))
	for i, s := range f {
		if len(s) < 2 {
			continue
		}
		v := s[0]
		if v == '"' || v == '\'' {
			s = s[1:]
		}
		v = s[len(s)-1]
		if v == '"' || v == '\'' {
			s = s[:len(s)-1]
		}
		f[i] = s
	}
	c := clir.NewCli("", "", "").Action(func() error { return nil })
	for i := range args {
		c.AddFlags(args[i])
	}
	err := c.Run(f...)
	return err
}
