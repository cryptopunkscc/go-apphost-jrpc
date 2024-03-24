package main

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/go-apphost-jrpc"
	"log"
	"time"
)

func main() {

	// register service
	ctx, cancel := context.WithCancel(context.Background())

	srv := NewApiService()
	rpc := jrpc.NewApp("testApi")
	rpc.Routes(
		"*",
		//"method*",
		//"method1",
	)
	rpc.Interface(srv)
	rpc.Logger(log.New(log.Writer(), "service ", 0))
	//rpc.Func("method", srv.Method)
	//rpc.Func("method1", srv.Method1)
	//rpc.Func("method2", srv.Method2)
	//rpc.Func("method2B", srv.Method2B)
	//rpc.Func("methodC", srv.MethodC
	//rpc.Func("method2S", srv.Method2S)
	if err := rpc.Run(ctx); err != nil {
		panic(err)
	}

	time.Sleep(time.Millisecond * 100)

	var err error
	rpcConn, err := jrpc.QueryFlow(id.Identity{}, "testApi")
	//rpcConn := jrpc.NewRequest(id.Identity{}, "testApi")
	if err != nil {
		panic(err)
	}
	rpcConn.Logger(log.New(log.Writer(), " client ", 0))

	// case
	if _, err = jrpc.Query[[]string](rpcConn, "api"); err != nil {
		panic(err)
	}

	// case
	if _, err = jrpc.Query[string](rpcConn, "string"); err != nil {
		panic(err)
	}

	// prepare client
	client := NewApiClient(rpcConn)

	// case
	client.Method(true, 10, "example")

	// case
	if err = client.Method1(false); err != nil {
		panic(err)
	}

	// case
	if err = client.Method1(true); err != nil && err.Error() != "example error" {
		panic(err)
	}

	// case
	if _, err = client.Method2(nil); err != nil {
		panic(err)
	}

	// case
	if _, err = client.Method2(&Arg{S: "example", I: 1000}); err != nil {
		panic(err)
	}

	// case
	if _, err = client.Method2S(); err != nil {
		panic(err)
	}

	// case
	if _, err = client.Method2B(); err != nil {
		panic(err)
	}

	// case
	c, err := client.MethodC()
	if err != nil {
		panic(err)
	}
	for range c {
	}

	// finish
	cancel()
}
