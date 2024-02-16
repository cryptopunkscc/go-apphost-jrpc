package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
)

type Client struct {
	rpc.Conn
}

func (c Client) Connect(identity id.Identity, port string) (client Client, err error) {
	client.Conn, err = rpc.QueryFlow(identity, port)
	return
}

func (c Client) Contacts() (<-chan []Contact, error) {
	return rpc.Subscribe[[]Contact](c, "contacts")
}
