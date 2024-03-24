package notify

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/go-apphost-jrpc"
	"github.com/cryptopunkscc/go-apphost-jrpc/android"
)

type Client struct {
	id.Identity
	jrpc.Conn
	port string
}

func NewClient() ApiClient {
	return &Client{}
}

func (c *Client) Connect() (err error) {
	if c.port == "" {
		c.port = android.NotifyPort
	}
	c.Conn, err = jrpc.QueryFlow(c.Identity, c.port)
	return
}

func (c *Client) Create(channel android.Channel) (err error) {
	return jrpc.Command(c.Conn, "create", channel)
}

func (c *Client) Notify(notification android.Notification) (err error) {
	return jrpc.Command(c.Conn, "notify", notification)
}

func Notifier(c ApiClient) (dispatch Notify) {
	nc := make(chan []android.Notification, 128)
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
