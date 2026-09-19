package network

import (
	"errors"
	"net"
	"testing"
	"time"

	"org.mutantcat.chickreomte/code/network/encoding/gzip"
	"org.mutantcat.chickreomte/code/network/encoding/proto"
)

func TestCompressedMessage(t *testing.T) {
	cp, err := gzip.New()
	if err != nil {
		t.Fatal(err)
	}
	c := &Conn{codec: proto.New(), compressor: cp}
	data, err := c.serialize(&Msg{From: "sender"})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := c.unserialize(data)
	if err != nil {
		t.Fatal(err)
	}
	if msg.GetFrom() != "sender" {
		t.Fatal(msg)
	}
}

func TestWriteAfterClose(t *testing.T) {
	a, b := net.Pipe()
	defer b.Close()
	c := NewConn(a)
	c.Close()
	if err := c.WriteMessage(&Msg{}, time.Second); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("expected closed error, got %v", err)
	}
}
