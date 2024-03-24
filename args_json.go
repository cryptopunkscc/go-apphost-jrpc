package jrpc

import (
	"encoding/json"
	"io"
	"reflect"
)

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

func (d jsonArgsDecoder) TestScan(scan io.ByteScanner) bool {
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

func (d jsonArgsDecoder) Decode(conn ByteScannerReader, args []any) error {
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
