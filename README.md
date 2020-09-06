# PubSub
PubSub is a Go Package for in-memory publish/subscribe.

[![test and build](https://github.com/davidscottmills/pubsub/workflows/test%20and%20build/badge.svg)](https://github.com/davidscottmills/pubsub/actions?query=workflow%3A%22test+and+build%22)
[![Coverage Status](https://coveralls.io/repos/github/davidscottmills/pubsub/badge.svg)](https://coveralls.io/github/davidscottmills/pubsub)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidscottmills/pubsub)](https://goreportcard.com/report/github.com/davidscottmills/pubsub)
[![Documentation](https://godoc.org/github.com/davidscottmills/pubsub?status.svg)](http://godoc.org/github.com/davidscottmills/pubsub)
[![GitHub issues](https://img.shields.io/github/issues/davidscottmills/pubsub.svg)](https://github.com/davidscottmills/pubsub/issues)
[![license](https://img.shields.io/github/license/davidscottmills/pubsub.svg?maxAge=2592000)](https://github.com/davidscottmills/pubsub/LICENSE.md)

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
