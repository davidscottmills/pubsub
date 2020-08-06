package pubsub

import (
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
}

type Msg struct {
	Subject string
	Data    interface{}
}

type MsgHandler func(m *Msg)

func NewPubSub() *PubSub {
	subs := make(map[int]*Subscription)
	ps := PubSub{mu: &sync.RWMutex{}, subscriptions: subs, ssid: 0}
	return &ps
}

func (ps *PubSub) Subscribe(subject string, mh MsgHandler) *Subscription {
	ps.mu.Lock()
	s := newSubscription(ps.ssid, subject, ps, mh)
	ps.subscriptions[ps.ssid] = s
	ps.ssid++
	go ps.subListen(s)
	ps.mu.Unlock()
	return s
}

func (s *Subscription) Unsubscribe() {
	s.ps.mu.Lock()
	s.uch <- true
	delete(s.ps.subscriptions, s.sid)
	s.ps.mu.Unlock()
}

func (ps *PubSub) Publish(subject string, data interface{}) {
	ps.mu.Lock()
	msg := newMessage(subject, data)
	for _, s := range ps.subscriptions {
		if s.subject == msg.Subject {
			s.mch <- msg
		}
	}
	ps.mu.Unlock()
}

func (ps *PubSub) subListen(s *Subscription) {
L:
	for {
		select {
		case msg := <-s.mch:
			go s.mh(msg)
		case <-s.uch:
			break L
		}
	}
}

func newMessage(subject string, data interface{}) *Msg {
	return &Msg{Subject: subject, Data: data}
}

func newSubscription(sid int, subject string, ps *PubSub, mh MsgHandler) *Subscription {
	return &Subscription{mu: &sync.Mutex{}, sid: sid, subject: subject, ps: ps, mh: mh, mch: make(chan *Msg), uch: make(chan bool)}
}
