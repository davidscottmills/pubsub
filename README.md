# PubSub

PubSub is a Go Package for in-memory publish/subscribe

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

    // Subscribe to a subject
    s := ps.Subscribe("subject.name", func(m *pubsub.Msg) {
        fmt.Println(m.Data)
    })
    defer s.Unsubscribe()

    // Publish
    ps.Publish("subject.name", "Hello, world!")

    runAndBlock()
}
```
## Tenets
- No race conditions
- Good performance

## TODO
- Implement and test close or drain on PubSub
- Implement and test errors