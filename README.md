# PubSub

PubSub is a Go Package for in-memory publish/subscribe.
This package is currently in prerelease.

## Installation

```bash
go get github.com/davidscottmills/pubsub
```

## Usage

```go
package main()

import (
    "fmt"
    "github.com/davidscottmills/pubsub"
)

var ps *PubSub

func main() {
    ps := NewPubSub()

    handerFunc := func(m *pubsub.Msg) {
        fmt.Println(m.Data)
    }

    // Unbounded number of go routines
    // Subscribe to a subject
    s, err := ps.Subscribe("subject.name", handerFunc)
    defer s.Unsubscribe()

    // Bound the number of concurrent handler functions by passing number
    // of concurrent go routines that you want to allow.
    // Subscribe to a subject
    s, err := ps.Subscribe("subject.name", handerFunc, 10)
    defer s.Unsubscribe()

    // Publish
    ps.Publish("subject.name", "Hello, world!")

    someStruct := struct{}{}

    // Publish any type you'd like
    ps.Publish("subject.name", someStruct)

    runAndBlock()
}
```

## TODO
- Implement and test close or drain on PubSub
- Implement and test errors