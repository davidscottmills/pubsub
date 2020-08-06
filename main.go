package main

import (
	"sync"
)

type PubSub struct {
	mu            sync.RWMutex
	subscriptions map[int]*Subscription
	mc            chan *Msg
	ssid          int
}

type Subscription struct {
	mu      sync.Mutex
	sid     int
	subject string
	ps      *PubSub
	mh      MsgHandler
	mch     chan *Msg
}

type Msg struct {
	subject string
	data    interface{}
}

type MsgHandler func(m *Msg)

func main() {

}

func NewPubSub() *PubSub {
	subs := make(map[int]*Subscription)
	mc := make(chan *Msg)
	ps := PubSub{mu: sync.RWMutex{}, subscriptions: subs, mc: mc, ssid: 0}
	go ps.listen()
	return &ps
}

func (ps *PubSub) Subscribe(subject string, mh MsgHandler) *Subscription {
	ps.mu.Lock()
	s := NewSubscription(ps.ssid, subject, ps, mh)
	ps.subscriptions[ps.ssid] = s
	ps.ssid++
	go ps.Listen(s)
	ps.mu.Unlock()
	return s
}

func (s *Subscription) Unsubscribe() {
	s.ps.mu.Lock()
	delete(s.ps.subscriptions, s.sid)
	s.ps.mu.Unlock()
}

func (ps *PubSub) Publish(subject string, data interface{}) {
	ps.mu.Lock()
	msg := NewMessage(subject, data)
	ps.mc <- msg
	ps.mu.Unlock()
}

func (ps *PubSub) listen() {
	for {
		msg := <-ps.mc
		for _, v := range ps.subscriptions {
			if v.subject == msg.subject {
				v.mch <- msg
			}
		}
	}
}

func (ps *PubSub) Listen(sub *Subscription) {
	for {
		ch := sub.mch
		msg := <-ch
		sub.mu.Lock()
		sub.mh(msg)
		sub.mu.Unlock()
	}
}

func NewMessage(subject string, data interface{}) *Msg {
	return &Msg{subject: subject, data: data}
}

func NewSubscription(sid int, subject string, ps *PubSub, mh MsgHandler) *Subscription {
	return &Subscription{mu: sync.Mutex{}, sid: sid, subject: subject, ps: ps, mh: mh, mch: make(chan *Msg)}
}
