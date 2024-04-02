simplified flow
```mermaid
flowchart
    App([App]) --> Query
    Module([Module]) --> Query
    Query --> Authorize -. only on first query .-> Rejected([Rejected])
    Authorize --> Execute --> Respond --> Return([EOF])
    Respond --> Await --> Return2([EOF])
    Await -.optional.-> Query2[Query]
    Query2 --> Execute
    Query2 -- if query changed --> Authorize
    
```

exact flow
```mermaid
flowchart
    NewModule([NewModule]) --> Router
    NewApp([NewApp]) --> Router -->
    Router.Interface --> Router.Routes --> Router.Logger --> 
    Router.Run --> 
    Module.registerRoute --> Module.RouteQuery
    Router.Run -->
    App.registerRoute --> App.routeQuery
    Module.RouteQuery --> Router.Query
    App.routeQuery --> Router.Query
    Router.Query --> Router.Authorize
    Conn -. optional args.-> Router.Call
    Router.Authorize --> Rejected(["Rejected"])
    Router.Authorize --> Router.Handle -->
    Router.Call[Router.Call] --> Router.respond
    Router.Handle --> Router.respond -- closed for write --> Return(["EOF"])
    Router.respond --> Scanner.Scan -- closed for read --> Return2(["EOF"])
    Conn2[Conn] -. optional next command .-> Scanner.Scan
    Scanner.Scan --> Router.Query2[Router.Query] --> Router.Handle
    Router.Query2 -- command changed --> Router.Authorize2[Router.Authorize] --> Router.Handle
```
