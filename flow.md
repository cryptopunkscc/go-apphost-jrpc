simplified flow
```mermaid
flowchart
    App([App]) --> Query
    Module([Module]) --> Query
    Query --> Authorize -.only on first query.-> Rejected([Rejected])
    Authorize --> Execute --> 
    Respond --> Return([EOF])
    Respond -->Await --> Return2([EOF])
    Await --> Execute
    Await --if query changed--> Authorize
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
    Router.Authorize --> Rejected(["Rejected"])
    Router.Authorize --> Router.Handle -->
    Router.Call[Router.Call] --> Router.respond
    Router.Handle --> Router.respond --> Return(["EOF"])
    Router.respond --> Scanner.Scan --> Return2(["EOF"])
    Scanner.Scan --> Router.Query2[Router.Query] --> Router.Handle
    Router.Query2 --> Router.Authorize2[Router.Authorize] --> Router.Handle
```
