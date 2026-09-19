package conn

import (
	"context"
	"sync"
	"testing"
	"time"

	"org.mutantcat.chickreomte/code/client/global"
	"org.mutantcat.chickreomte/code/network"
)

func testConn() *Conn {
	return &Conn{
		cfg:          &global.Configure{WriteTimeout: time.Millisecond},
		ctx:          context.Background(),
		read:         make(map[string]chan *network.Msg),
		drop:         make(map[string]time.Time),
		unknownRead:  make(chan *network.Msg, 1),
		onDisconnect: make(chan string, 1),
	}
}

func TestDisconnectClosesLink(t *testing.T) {
	c := testConn()
	c.AddLink("link")
	ch := c.ChanRead("link")
	c.handleLinkedMessage(&network.Msg{XType: network.Msg_disconnect, LinkId: "link"})
	select {
	case <-ch:
	default:
		t.Fatal("disconnect did not unblock link reader")
	}
}

func TestUnknownReadIsClosedForMissingLink(t *testing.T) {
	c := testConn()
	select {
	case <-c.ChanRead("missing"):
	default:
		t.Fatal("missing link would block indefinitely")
	}
}

func TestConcurrentDispatchClose(t *testing.T) {
	for i := 0; i < 100; i++ {
		c := testConn()
		c.AddLink("link")
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); c.Requeue("link", &network.Msg{}) }()
		go func() { defer wg.Done(); c.ChanClose("link") }()
		wg.Wait()
	}
}
