package pubsub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_PubSub_Subscribe_Unsubscribe(t *testing.T) {
	mh := func(m *Msg) {}
	ps := NewPubSub()

	s1 := ps.Subscribe("test.test", mh)
	require.Same(t, ps, s1.ps)
	require.Equal(t, 1, len(ps.subscriptions))

	s2 := ps.Subscribe("test.test", mh)
	require.Same(t, ps, s2.ps)
	require.Equal(t, 2, len(ps.subscriptions))
	require.NotEqual(t, s1.sid, s2.sid)

	s1.Unsubscribe()
	require.Equal(t, 1, len(ps.subscriptions))

	s2.Unsubscribe()
	require.Equal(t, 0, len(ps.subscriptions))
}

func Test_PubSub_Listen(t *testing.T) {
	mh1called, mh2called, mh3called := false, false, false
	mh1 := func(m *Msg) { mh1called = true }
	mh2 := func(m *Msg) { mh2called = true }
	mh3 := func(m *Msg) { mh3called = true }

	subject := "test.test"
	ps := NewPubSub()

	ps.Subscribe(subject, mh1)
	ps.Subscribe(subject, mh2)
	ps.Subscribe("not.test", mh3)
	ps.Publish(subject, "Hello, world!")

	time.Sleep(2 * time.Second)

	require.True(t, mh1called)
	require.True(t, mh2called)
	require.False(t, mh3called)
}

func Test_PubSub_Listen_Unsubscribe(t *testing.T) {
	mh1called, mh2called := false, false
	mh1 := func(m *Msg) { mh1called = true }
	mh2 := func(m *Msg) { mh2called = false }

	subject := "test.test"
	ps := NewPubSub()
	ps.Subscribe(subject, mh1)
	s2 := ps.Subscribe(subject, mh2)
	s2.Unsubscribe()
	ps.Publish(subject, "Hello, world!")

	time.Sleep(2 * time.Second)

	require.True(t, mh1called)
	require.False(t, mh2called)
}
