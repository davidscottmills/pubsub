# PubSub

PubSub is a Go Package for in-memory publish/subscribe.

[![test and build](https://github.com/davidscottmills/pubsub/workflows/test%20and%20build/badge.svg)](https://github.com/davidscottmills/pubsub/actions?query=workflow%3A%22test+and+build%22)
[![Coverage Status](https://coveralls.io/repos/github/davidscottmills/pubsub/badge.svg)](https://coveralls.io/github/davidscottmills/pubsub)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidscottmills/pubsub)](https://goreportcard.com/report/github.com/davidscottmills/pubsub)
[![Documentation](https://godoc.org/github.com/davidscottmills/pubsub?status.svg)](http://godoc.org/github.com/davidscottmills/pubsub)
[![GitHub issues](https://img.shields.io/github/issues/davidscottmills/pubsub.svg)](https://github.com/davidscottmills/pubsub/issues)
[![license](https://img.shields.io/github/license/davidscottmills/pubsub.svg?maxAge=2592000)](https://github.com/davidscottmills/pubsub/LICENSE.md)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/davidscottmills/pubsub.svg)](https://github.com/davidscottmills/pubsub)

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

    // Subscribe to a subject
    s, err := ps.Subscribe("subject.name", handerFunc)
    defer s.Unsubscribe()

    // Publish
    err := ps.Publish("subject.name", "Hello, world!")

    someStruct := struct{}{}

    // Publish any type you'd like
    ps.Publish("subject.name", someStruct)

    runAndBlock()
}
```

## Subject Routing

Subject naming matches that of NATS.

### Subject Names

- All ascii alphanumeric characters except spaces/tabs and separators which are "." and ">" are allowed. Subject names can be optionally token-delimited using the dot character (.)
- Subjects are case sensative

### Subject Wildcards
- The asterisk character (*) matches a single token at any level of the subject.
- The greater than symbol (>), also known as the full wildcard, matches one or more tokens at the tail of a subject, and must be the last token. The wildcarded subject foo.> will match foo.bar or foo.bar.baz.1, but not foo. 
- Wildcards must be a separate token (foo.*.baz or foo.> are syntactically valid; foo*.bar, f*o.b*r and foo> are not)

## TODO

- Implement and test close or drain on PubSub
- Implement and test errors
