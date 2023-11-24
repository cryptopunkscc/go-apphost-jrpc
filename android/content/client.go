package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"io"
)

type Client struct {
	id.Identity
	rpc.Conn
}

func (c *Client) Connect() (err error) {
	conn, err := astral.Query(c.Identity, Port)
	if err == nil {
		c.Conn = *rpc.NewConn(context.Background(), conn)
	}
	return
}

func (c *Client) Info(uri string) (files Info, err error) {
	return rpc.Query[Info](c.Conn, "info", uri)
}

func (c *Client) Reader(uri string, offset int64) (reader io.ReadCloser, err error) {
	if err = rpc.Call(c.Conn, "reader", uri, offset); err != nil {
		return
	}
	reader = &c.Conn
	return
}
