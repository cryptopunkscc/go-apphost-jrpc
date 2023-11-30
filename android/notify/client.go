package notify

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/astral"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
)

type Client struct {
	id.Identity
	rpc.Conn
	port string
}

func (c *Client) Connect() (err error) {
	if c.port == "" {
		c.port = Port
	}
	conn, err := astral.Query(c.Identity, c.port)
	if err == nil {
		c.Conn = *rpc.NewConn(conn)
	}
	return
}

func (c *Client) Create(channel Channel) (err error) {
	return rpc.Command(c.Conn, "create", channel)
}

func (c *Client) Notify(notification Notification) (err error) {
	return rpc.Command(c.Conn, "notify", notification)
}

func Notifier(c ApiClient) (dispatch Notify) {
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
