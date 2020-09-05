package pubsub

import (
	"errors"
	"sync"
)

type PubSub struct {
	mu            *sync.RWMutex
	subscriptions map[int]*Subscription
	ssid          int
}

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

type Msg struct {
	Subject string
	Data    interface{}
}

type MsgHandler func(m *Msg)

var (
	ErrSubscriptionBoundingSettingsError = errors.New("pubsub: the input for number of concurrent go routines must be an int that is greater than 0")
)

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

func (s *Subscription) Unsubscribe() {
	s.ps.mu.Lock()
	s.uch <- true
	delete(s.ps.subscriptions, s.sid)
	s.ps.mu.Unlock()
}

func (ps *PubSub) Publish(subject string, data interface{}) {
	ps.mu.RLock()
	msg := newMessage(subject, data)
	for _, s := range ps.subscriptions {
		if s.subject == msg.Subject {
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
	return &Msg{Subject: subject, Data: data}
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
