package notify

import (
	"context"
	"fmt"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"github.com/cryptopunkscc/go-apphost-jrpc/android"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

type testClient struct {
	Client
}

func NewTestClient() ApiClient {
	return &Client{port: testPort}
}

func ConnectTestClient(t *testing.T) (ApiClient, func()) {
	c := &Client{port: testPort}
	if err := c.Connect(); err != nil {
		t.Fatal(err)
	}
	c.WithLogger(log.Default())
	return c, func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestServer(t *testing.T, err bool) (cancelFunc context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		err := rpc.Server[testService]{
			Handler: func(ctx context.Context, conn rpc.Conn) (ts testService) {
				ts.err = err
				return
			},
		}.Run(ctx)
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(time.Second * 1)
	return
}

const testPort = "android/notify/jrpc/test"

var _ android.NotifyServiceApi = testService{}

type testService struct {
	err bool
}

func (t testService) String() string {
	return testPort
}

func (t testService) Create(channel *android.Channel) (err error) {
	if t.err {
		err = Response(channel)
	} else {
		log.Println(channel)
	}
	return
}

func (t testService) Notify(notification *android.Notification) (err error) {
	if t.err {
		err = Response(notification)
	} else {
		log.Println(notification)
	}
	return
}

func Response(args ...any) error {
	return fmt.Errorf("%v", args...)
}

func Verify(t *testing.T, err error, args ...any) {
	assert.EqualError(t, err, fmt.Sprint(args...))
}
