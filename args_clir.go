package jrpc

import (
	"github.com/leaanthony/clir"
	"io"
	"strings"
)

type clirArgsDecoder struct{}

func NewClirArgsDecoder() ArgsDecoder { return &clirArgsDecoder{} }

func (d clirArgsDecoder) Decode(conn ByteScannerReader, args []any) (err error) {
	end := []byte(`\n`)
	b := byte(0)
	var s = 0
	var n = 0
	for {
		if b, err = conn.ReadByte(); err != nil {
			return
		}
		n++
		if b != end[s] {
			s = 0
			continue
		}
		s++
		if s == len(end) {
			break
		}
	}
	for i := 0; i < n; i++ {
		_ = conn.UnreadByte()
	}
	bytes := make([]byte, n)
	if _, err = conn.Read(bytes); err != nil {
		return
	}
	bytes = bytes[:len(bytes)-2]
	err = d.Unmarshal(bytes, args)
	return
}

func (d clirArgsDecoder) Begin() []rune {
	return []rune{'$', ' '}
}

func (d clirArgsDecoder) Test(b []byte) bool {
	return b[0] == '$' || b[0] == ' '
}

func (d clirArgsDecoder) TestScan(scan io.ByteScanner) bool {
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
