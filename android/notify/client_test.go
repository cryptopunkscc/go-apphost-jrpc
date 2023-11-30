package notify

import (
	"context"
	"testing"
	"time"
)

func TestClient_All(t *testing.T) {
	cancelServer := TestServer(t, true)
	defer cancelServer()

	t.Run("Create", func(t *testing.T) {
		c, cancel := ConnectTestClient(t)
		defer cancel()
		expected := Channel{
			Id:         "test",
			Name:       "test",
			Importance: 9001,
		}
		actual := c.Create(expected)
		Verify(t, actual, expected)
	})

	t.Run("Notify", func(t *testing.T) {
		c, cancel := ConnectTestClient(t)
		defer cancel()
		expected := Notification{
			Id:        1,
			ChannelId: "id",
		}
		actual := c.Notify(expected)
		Verify(t, actual, expected)
	})
}

func TestSelect(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second * 1)
		cancelFunc()
	}()
	select {
	case <-ctx.Done():
		return
	default:
		time.Sleep(time.Second * 5)
	}
}
