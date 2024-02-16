package contacts

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
)

const ServiceName = "contacts"

func init() {
	if err := modules.RegisterModule(ServiceName, Loader{}); err != nil {
		panic(err)
	}
}

type Loader struct{}

func (Loader) Load(node modules.Node, _ assets.Assets, log *log.Logger) (modules.Module, error) {
	mod := &Module{node: node, log: log}
	return mod, nil
}

type Module struct {
	node modules.Node
	log  *log.Logger
}

func (m *Module) Run(ctx context.Context) error {
	err := m.node.Router().AddRoute(m.node.Identity(), m.node.Identity(), m, 0)
	if err != nil {
		return err
	}
	return nil
}

func (m *Module) RouteQuery(
	ctx context.Context,
	query net.Query,
	caller net.SecureWriteCloser,
	hints net.Hints,
) (net.SecureWriteCloser, error) {
	if hints.Origin != net.OriginLocal {
		return nil, net.ErrRejected
	}
	return net.Accept(query, caller, func(conn net.SecureConn) {
		rpc.Handle(ctx, conn, func(ctx2 context.Context, rpc rpc.Conn) any {
			return &service{
				Module: m,
				parent: ctx,
				ctx:    ctx2,
				conn:   rpc,
			}
		})
	})
}
