package main

import (
	"context"
	"errors"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"time"
)

type apiService struct {
	conn rpc.Conn
}

func NewApiService(_ context.Context, conn rpc.Conn) Api {
	return &apiService{conn: conn}
}

func (a apiService) String() string {
	return "testApi"
}

func (a apiService) Method(_ bool, _ int, _ string) {
	// no-op
}

func (a apiService) Method1(b bool) (err error) {
	if b {
		err = errors.New("example error")
	}
	return
}

func (a apiService) Method2(arg *Arg) (Arg, error) {
	return Arg{S: "some text", I: 1, Arg: arg}, nil
}

func (a apiService) Method2S() (string, error) {
	return a.String(), nil
}

func (a apiService) Method2B() (bool, error) {
	return true, nil
}

func (a apiService) MethodC() (out <-chan Arg, err error) {
	c := make(chan Arg)
	out = c
	go func() {
		for i := 0; i < 5; i++ {
			c <- Arg{I: 100 + i}
			time.Sleep(1 * time.Millisecond)
		}
		close(c)
	}()
	return
}
