package main

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/go-apphost-jrpc"
	"io"
	"log"
	"time"
)

func main() {

	// register service
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := rpc.Server[Api]{
			Accept: func(query *astral.QueryData) (conn io.ReadWriteCloser, err error) {
				conn, err = query.Accept()
				conn = rpc.NewConnLogger(conn, log.New(log.Writer(), "service ", 0))
				return
			},
			Handler: func(ctx context.Context, conn rpc.Conn) Api {
				return NewApiService()
			},
		}.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Millisecond * 100)

	var err error
	rpcConn, err := rpc.QueryFlow(id.Identity{}, apiService{}.String())
	//rpcConn := rpc.NewRequest(id.Identity{}, apiService{}.String())
	rpcConn.WithLogger(log.New(log.Writer(), " client ", 0))
	if err != nil {
		panic(err)
	}

	// case
	if _, err = rpc.Query[[]string](rpcConn, "api"); err != nil {
		panic(err)
	}

	// case
	if _, err = rpc.Query[string](rpcConn, "string"); err != nil {
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
