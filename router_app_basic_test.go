package jrpc

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"github.com/stretchr/testify/assert"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestApp_Run_basic(t *testing.T) {
	port := "test_app"
	app := NewApp(port)
	app.Routes("*")
	app.Func("func0", function0)
	app.Func("func0!", function0)
	app.Func("func1", function1)
	app.Func("func2", function2)
	app.Func("func3", function3)
	app.Func("func4", function4)
	app.Func("func5", function5)
	app.Func("func6", function6)
	app.Func("func7", function7)
	app.Func("func8", function8)
	app.Func("func9", function9)
	app.Func("func10", function10)
	tests := []struct {
		name     string
		expected any
		conn     int
	}{
		{name: "asd", expected: proto.ErrRejected, conn: 1},
		{name: "asd", expected: ErrMalformedRequest, conn: 2},
		{name: "func0", expected: proto.ErrRejected, conn: 1},
		{name: "func0", expected: ErrUnauthorized, conn: 2},
		{name: "func1", expected: map[string]any{}},
		{name: "func2[1]", expected: float64(1)},
		{name: "func2 1", expected: float64(1)},
		{name: "func3", expected: err3},
		{name: "func4[true]", expected: true},
		{name: "func4 true", expected: true},
		{name: "func4[false]", expected: err4},
		{name: "func4 false", expected: err4},
		{name: `func5[true, 1, "a"]`, expected: []any{true, float64(1), "a"}},
		{name: "func5 true 1 a", expected: []any{true, float64(1), "a"}},
		{name: `func6{"i":1}`, expected: &structI{1}},
		{name: `func6[{"i":1}]`, expected: structI{1}},
		{name: `func6 -i 1`, expected: structI{1}},
		{name: `func6 -i 1`, expected: &structI{1}},
		{name: "func7[]", expected: (*structI)(nil)},
		{name: `func7{"i":1}`, expected: &structI{1}},
		{name: "func7 ", expected: (*structI)(nil)},
		//{name: "func7 -i 1", expected: &structI{1}}, //FIXME
		{name: `func8[{"i":1},{"b":true}]`, expected: struct1{structI{1}, structB{true}}},
		{name: `func8 -i 1 -b true`, expected: struct1{structI{1}, structB{true}}},
		{name: `func9[{"i":1},{"b":true}]`, expected: struct1{structI{1}, structB{true}}},
		{name: `func9[{"i":1}]`, expected: struct1{structI{1}, structB{false}}},
		//{name: `func9 -i 1 -b true`, expected: struct1{structI{1}, structB{true}}}, //FIXME
		{name: `func10{"i":{"i":1},"b":{"b":true}}`, expected: struct3{structI{1}, structB{true}}},
		{name: `func10{"i":{"i":1}}`, expected: struct3{structI{1}, structB{false}}},
	}
	clients := []func() Conn{
		func() (c Conn) {
			c = NewRequest(id.Anyone, port)
			c.Logger(log.New(log.Writer(), "", 0))
			return
		},
		func() (c Conn) {
			c, err := QueryFlow(id.Anyone, port)
			if err != nil {
				t.Fatal(err)
			}
			c.Logger(log.New(log.Writer(), "", 0))
			return
		},
	}

	if err := app.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Millisecond)
	for _, tt := range tests {
		for i, c := range clients {
			if tt.conn > 0 && tt.conn-1 != i {
				continue
			}

			client := c()
			t.Run(tt.name, func(t *testing.T) {
				if err := Call(client, tt.name); err != nil {
					assert.Equal(t, tt.expected, err)
					return
				}

				v := reflect.New(reflect.TypeOf(tt.expected))
				if err := client.Decode(v.Interface()); err != nil {
					assert.Equal(t, tt.expected, err)
				} else {
					assert.Equal(t, tt.expected, v.Elem().Interface())
				}
			})
		}
	}
}

func function0() bool { return false }

func function1() {
	log.Println("function1")
}

func function2(i int) int {
	return i
}

var err3 = errors.New("test error 3")

func function3() error {
	return err3
}

var err4 = errors.New("test error 4")

func function4(b bool) (bool, error) {
	if b {
		return true, nil
	}
	return false, err4
}

func function5(b bool, i int, s string) (bool, int, string) {
	return b, i, s
}

type structI struct {
	I int `json:"i" name:"i"`
}

type structB struct {
	B bool `json:"b" name:"b"`
}

type struct1 struct {
	structI
	structB
}

type struct2 struct {
	*structI
	*structB
}

type struct3 struct {
	StructI structI `json:"I"`
	StructB structB `json:"B"`
}

func function6(i structI) structI {
	return i
}

func function7(i *structI) *structI {
	return i
}

func function8(i structI, b structB) struct1 {
	return struct1{i, b}
}

func function9(i *structI, b *structB) *struct2 {
	return &struct2{i, b}
}

func function10(s struct3) struct3 {
	return s
}
