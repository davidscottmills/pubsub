// Package pubsub provides a light weight in-memory way of doing subject based publish-subscribe.
package pubsub

import (
	"errors"
	"sync"
)

// A PubSub represents the client that mangages subscribing, publishing and routing of messages to subject handlers.
type PubSub struct {
	mu            *sync.RWMutex
	subscriptions map[int]*Subscription
	ssid          int
}

// A Subscription is a subscription to a specific subject.
type Subscription struct {
	mu      *sync.Mutex
	sid     int
	subject string
	ps      *PubSub
	mh      MsgHandler
	mch     chan *Msg //message channel
	uch     chan bool //unsubscribe channel
	sem     *chan int //semaphore to control number of concurrent go routines
}

// A Msg is a message that is to be handled by subscribers when data is published to a subject.
type Msg struct {
	subject string
	// The data that the message is meant for a handler to use.
	Data interface{}
}

// A MsgHandler is a function that subject subscribers must pass to handle messages.
type MsgHandler func(m *Msg)

var (
	// An error given when bounding settings have been pass incorrectly.
	ErrSubscriptionBoundingSettingsError = errors.New("pubsub: the input for number of concurrent go routines must be an int that is greater than 0")
)

// Instantiate a new PubSub client.
func NewPubSub() *PubSub {
	subs := make(map[int]*Subscription)
	ps := PubSub{mu: &sync.RWMutex{}, subscriptions: subs, ssid: 0}
	return &ps
}

// Subscribe to a subject
// subject - The subject you want to subscribe to
// mh - The message handler
// args - number of allowed concurrent go routines. Default is not to throttle.
func (ps *PubSub) Subscribe(subject string, mh MsgHandler, args ...interface{}) (*Subscription, error) {
	ncgr := 0
	if len(args) > 0 {
		n, ok := args[0].(int)
		if !ok || n <= 0 {
			return nil, ErrSubscriptionBoundingSettingsError
		}
		ncgr = n
	}

	ps.mu.Lock()
	s := newSubscription(ps.ssid, ncgr, subject, ps, mh)
	ps.subscriptions[ps.ssid] = s
	ps.ssid++
	go ps.subListen(s)
	ps.mu.Unlock()
	return s, nil
}

// Unsubscribe from a subject.
func (s *Subscription) Unsubscribe() {
	s.ps.mu.Lock()
	s.uch <- true
	delete(s.ps.subscriptions, s.sid)
	s.ps.mu.Unlock()
}

// Publish data to a subject.
// subject - the subject you want to pass the data to
// data - the data you want to pass
func (ps *PubSub) Publish(subject string, data interface{}) {
	msg := newMessage(subject, data)
	ps.mu.RLock()
	for _, s := range ps.subscriptions {
		if s.subject == msg.subject {
			s.mch <- msg
		}
	}
	ps.mu.RUnlock()
}

func (ps *PubSub) subListen(s *Subscription) {
L:
	for {
		select {
		case msg := <-s.mch:
			s.mh(msg)
		case <-s.uch:
			break L
		}
	}
}

func newMessage(subject string, data interface{}) *Msg {
	return &Msg{subject: subject, Data: data}
}

func newSubscription(sid, ncgr int, subject string, ps *PubSub, mh MsgHandler) *Subscription {
	nmh, sem := newMessageHandlerWrapper(ncgr, mh)

	return &Subscription{mu: &sync.Mutex{}, sid: sid, subject: subject, ps: ps, mh: nmh, mch: make(chan *Msg), uch: make(chan bool), sem: sem}
}

func newMessageHandlerWrapper(ncgr int, mh MsgHandler) (MsgHandler, *chan int) {
	var sem chan int

	// Unbounded concurrency
	nmh := func(m *Msg) {
		go mh(m)
	}

	if ncgr > 0 {
		sem = make(chan int, ncgr)

		// Bounded concurrency
		nmh = func(m *Msg) {
			sem <- 1
			go func() {
				mh(m)
				<-sem
			}()
		}
	}

	return nmh, &sem
}
