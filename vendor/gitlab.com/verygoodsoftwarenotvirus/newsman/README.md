# Newsman

## Usage

Newsman is meant to provide websocket and webhook eventing within a REST server.

```go
package main

import(
    "log"
    "net/http"

    "gitlab.com/verygoodsoftwarenotvirus/newsman"
)

type Item struct {
    Name string `json:"name"`
}

func main(){
    mux := http.NewServeMux()

    // supplant with your own authz method
    authFunc := func(req *http.Request) bool { return true }

    nm := newsman.NewNewsman(authFunc)

    mux.HandleFunc("/websocket", nm.ServeWebsockets)
    mux.HandleFunc("/item", func(res http.ResponseWriter, req *http.Request) {

        nm.Report(newsman.Event{
            EventType: "new_item",
            Data: Item{Name: "hello!"},
            Topics: []string{"new"},
        })

    })

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

You can see an example like this by running `make example`.

## Websocket filtering

Newsman's `ServeWebsockets` method is an HTTP handler that establishes and maintains websocket connections. Newsman can read through a request URL to determine what types/events/topics you're interested in. So for instance, if I have a running newsman server, and I make a request to `example.com/websocket?event=create&type=invitation&topic=wedding`, I will only get websocket notifications if the event is of "create" type, if the data structure itself is called an "invitation", and if the topic tags associated with the event included the word "wedding". Alternatively, I can make a request to `example.com/websocket?event=*&type=*&topic=*`, I will get all information that the newsman manages (provided that the authorization function allows you).
