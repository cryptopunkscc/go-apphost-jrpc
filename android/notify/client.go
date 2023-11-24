package notify

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
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

func (c *Client) Create(channel Channel) (err error) {
	return rpc.Command(c.Conn, "create", channel)
}

func (c *Client) Notify(notification Notification) (err error) {
	return rpc.Command(c.Conn, "notify", notification)
}

func (c *Client) Notifier() (dispatch Notify) {
	nc := make(chan []Notification, 128)
	dispatch = nc
	go func() {
		defer c.Close()
		for notifications := range nc {
			for _, n := range notifications {
				err := c.Notify(n)
				if err != nil {
					return
				}
			}
		}
	}()
	return
}
