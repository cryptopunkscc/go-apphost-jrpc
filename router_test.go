package jrpc

import (
	"context"
	"testing"
)

func TestRouter(t *testing.T) {
	r := NewRouter()
	f := func(ctx context.Context, conn Conn, arg testRouterStruct) testRouterStruct {
		return arg
	}
	r.Func("test", f)
	r.Route(context.Background(), `test{"i":1,"s":"a"}`, *NewFlow(nil))
}

type testRouterStruct struct {
	I int    `json:"i" pos:"1"`
	S string `json:"s" pos:"2"`
}
