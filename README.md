# Apphost JSON RPC

Basic extension to [lib astral](https://github.com/cryptopunkscc/astrald/tree/master/lib/astral). Provides a convenient wrapper for custom JSON RPC protocol.


## Usage

Implementing and running a service:

```go
package main

import rpc "github.com/cryptopunkscc/go-apphost-jrpc"

type service struct{}

func newService(_ *rpc.Conn) *service {
	return &service{}
}

func (s service) String() string {
	return "simple_calc"
}

func (s service) Sum(a int, b int) (c int, err error) {
	c = a + b
	return
}

func main() {
	err := rpc.Server[service]{Handler: newService}.Run()
	if err != nil {
		panic(err)   
	}
}
```

Calling method on the service:

```go
package main

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	rpc "github.com/cryptopunkscc/go-apphost-jrpc"
)

func main() {
	conn, _ := astral.Query(id.Identity{}, "simple_calc")
	r, _ := rpc.Query[int](conn, "sum", 2, 2)
	println(r)
}
```

See more comprehensive [example](./example).


## Protocol 

The client can request data from Service by sending a method represented by a JSON array, where the first element is a method name and the rest are positional arguments.

example method:

```json
["methodName", 1, true, "string arg", {"name": "object arg"}]
```

The service can respond by sending:
* `null` if there is nothing to send. 
* One error object.
* N amount of JSON objects.

example error:
```json
{"error": "some error message"}
```

The client can request a list of API methods provided by service by sending reserved method:

```json
["api"]
```

In response the service should send a JSON array containing method names:

```json
["api", "method1", "method2", "methodN"]
```
