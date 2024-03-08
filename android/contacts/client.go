package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/go-apphost-jrpc"
)

type Client struct {
	jrpc.Conn
}

func (c Client) Connect(identity id.Identity, port string) (client Client, err error) {
	client.Conn, err = jrpc.QueryFlow(identity, port)
	return
}

func (c Client) Contacts() (<-chan []Contact, error) {
	return jrpc.Subscribe[[]Contact](c, "contacts")
}
