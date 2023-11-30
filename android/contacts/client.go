package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
)

type Client struct {
	rpc.Conn
}

func (c Client) Connect(identity id.Identity, port string) (client Client, err error) {
	if c.ReadWriteCloser, err = astral.Query(identity, port); err == nil {
		client.Conn = *rpc.NewConn(c.ReadWriteCloser)
	}
	return
}

func (c Client) Contacts() (<-chan []Contact, error) {
	return rpc.Subscribe[[]Contact](c.Conn, "contacts")
}
