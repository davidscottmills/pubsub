package pubsub

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_PubSub_Subscribe_Invalid_Num_Concurrent_Go_Routines_Returns_Error(t *testing.T) {
	mh := func(m *Msg) {}
	ps := NewPubSub()

	ncgrs := []int{0, -1}
	for _, ncgr := range ncgrs {
		_, err := ps.Subscribe("test.test", mh, ncgr)
		require.Equal(t, ErrSubscriptionBoundingSettingsError, err)
	}
}
func Test_PubSub_Subscribe_Unsubscribe(t *testing.T) {
	mh := func(m *Msg) {}
	ps := NewPubSub()

	s1, _ := ps.Subscribe("test.test", mh)
	require.Same(t, ps, s1.ps)
	require.Equal(t, 1, len(ps.subscriptions))

	s2, _ := ps.Subscribe("test.test", mh)
	require.Same(t, ps, s2.ps)
	require.Equal(t, 2, len(ps.subscriptions))
	require.NotEqual(t, s1.sid, s2.sid)

	s1.Unsubscribe()
	require.Equal(t, 1, len(ps.subscriptions))

	s2.Unsubscribe()
	require.Equal(t, 0, len(ps.subscriptions))
}

func Test_PubSub_Listen(t *testing.T) {
	mh3called := false
	mhch1, mhch2 := make(chan bool), make(chan bool)
	mh1 := func(m *Msg) {
		mhch1 <- true
	}
	mh2 := func(m *Msg) {
		mhch2 <- true
	}
	mh3 := func(m *Msg) { mh3called = true }

	subject := "test.test"
	ps := NewPubSub()

	s1, _ := ps.Subscribe(subject, mh1)
	defer s1.Unsubscribe()
	s2, _ := ps.Subscribe(subject, mh2)
	defer s2.Unsubscribe()
	s3, _ := ps.Subscribe("not.test", mh3)
	defer s3.Unsubscribe()
	ps.Publish(subject, "Hello, world!")

	for i := 0; i < 2; i++ {
		select {
		case <-mhch1:
			continue
		case <-mhch2:
			continue
		}
	}

	require.False(t, mh3called)
}

func Test_PubSub_Listen_Unsubscribe_Publish(t *testing.T) {
	mh2called := false
	mhch1 := make(chan bool)
	mh1 := func(m *Msg) {
		mhch1 <- true
	}
	mh2 := func(m *Msg) { mh2called = true }

	subject := "test.test"
	ps := NewPubSub()
	s1, _ := ps.Subscribe(subject, mh1)
	defer s1.Unsubscribe()
	s2, _ := ps.Subscribe(subject, mh2)
	s2.Unsubscribe()
	ps.Publish(subject, "Hello, world!")

	<-mhch1

	require.False(t, mh2called)
}

func Test_PubSub_Listen_Publish_Unsubscribe(t *testing.T) {
	mhch1, mhch2 := make(chan bool), make(chan bool)
	mh1 := func(m *Msg) { mhch1 <- true }
	mh2 := func(m *Msg) { mhch2 <- true }

	subject := "test.test"
	ps := NewPubSub()
	s1, _ := ps.Subscribe(subject, mh1)
	defer s1.Unsubscribe()
	s2, _ := ps.Subscribe(subject, mh2)
	// If we publish before unsubscribe is called,
	// we should expect that the subscriber will receive the message.
	ps.Publish(subject, "Hello, world!")
	go s2.Unsubscribe()

	for i := 0; i < 2; i++ {
		select {
		case <-mhch1:
		case <-mhch2:
		}
	}
}

func Test_PubSub_Multiple_Messages(t *testing.T) {
	mhch1 := make(chan bool)
	mh1 := func(m *Msg) {
		mhch1 <- true
		time.Sleep(2 * time.Second)
	}

	subject := "test.test"
	ps := NewPubSub()
	s1, _ := ps.Subscribe(subject, mh1)
	defer s1.Unsubscribe()
	for i := 0; i < 100; i++ {
		ps.Publish(subject, "Hello, world!")
	}

	for i := 0; i < 100; i++ {
		<-mhch1
	}
}

func Test_Bounded_Handler_Setting(t *testing.T) {
	mhch1 := make(chan bool)
	mh1 := func(m *Msg) {
		mhch1 <- true
		time.Sleep(2 * time.Second)
	}

	subject := "test.test"
	ncgr := 10
	ps := NewPubSub()
	s, _ := ps.Subscribe(subject, mh1, 10)
	defer s.Unsubscribe()
	for i := 0; i < 11; i++ {
		ps.Publish(subject, "Hello, world!")
	}

	for i := 0; i < 11; i++ {
		if i == 11 {
			require.Equal(t, ncgr, len(*s.sem))
		}
		<-mhch1

	}
}

func Benchmark_HelloWorld_TenSubscriptions_OneMessagesPerSubscription(b *testing.B) {
	// Setup
	ps := NewPubSub()
	mch := make(chan bool)

	subs := []string{}
	for i := 0; i < 10; i++ {
		sub := "sub." + strconv.Itoa(i)
		mh := func(m *Msg) { mch <- true }
		s, _ := ps.Subscribe(sub, mh)
		defer s.Unsubscribe()
		subs = append(subs, sub)
	}
	nmps := 100

	// Start the test
	b.ResetTimer()
	for i := 0; i < nmps; i++ {
		for _, sub := range subs {
			ps.Publish(sub, "Hello World!")
		}
	}

	// Ensure all go routines complete
	for i := 0; i < nmps*len(subs); i++ {
		<-mch
	}
}

func Benchmark_HelloWorld_TenByTenSubscriptions_OneMessagesPerSubscription(b *testing.B) {
	// Setup
	ps := NewPubSub()
	mch := make(chan bool)

	subs := []string{}
	subsPerSub := 10
	for i := 0; i < 10; i++ {
		sub := "sub." + strconv.Itoa(i)
		subs = append(subs, sub)
		for j := 0; j < subsPerSub; j++ {
			mh := func(m *Msg) { mch <- true }
			s, _ := ps.Subscribe(sub, mh)
			defer s.Unsubscribe()
		}
	}
	nmps := 100

	// Start the test
	b.ResetTimer()
	for i := 0; i < nmps; i++ {
		for _, sub := range subs {
			ps.Publish(sub, "Hello World!")
		}
	}

	// Ensure all go routines complete
	for i := 0; i < nmps*len(subs)*subsPerSub; i++ {
		<-mch
	}
}

func Fib(n int) int {
	if n < 2 {
		return n
	}
	return Fib(n-1) + Fib(n-2)
}

func Benchmark_Fibonacci_TenByTenSubscriptions_OneMessagesPerSubscription(b *testing.B) {
	// Setup
	ps := NewPubSub()
	mch := make(chan bool)

	subs := []string{}
	subsPerSub := 10
	for i := 0; i < 10; i++ {
		sub := "sub." + strconv.Itoa(i)
		subs = append(subs, sub)
		for j := 0; j < subsPerSub; j++ {
			mh := func(m *Msg) {
				// Try m.Data cast to int
				mi, ok := m.Data.(int)
				if !ok {
					panic("m.Data was not an int")
				}
				// Do expensive Fibonacci computation
				Fib(mi)
				mch <- true
			}
			s, _ := ps.Subscribe(sub, mh)
			defer s.Unsubscribe()
		}
	}
	nmps := 100

	// Start the test
	b.ResetTimer()
	for i := 0; i < nmps; i++ {
		for _, sub := range subs {
			// Calculating 20th Fibonacci number should be sufficiently difficult
			ps.Publish(sub, 20)
		}
	}

	// Ensure all go routines complete
	for i := 0; i < nmps*len(subs)*subsPerSub; i++ {
		<-mch
	}
}
