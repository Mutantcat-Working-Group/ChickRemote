package handler

import (
	"net"
	"testing"
	"time"

	"org.mutantcat.chickreomte/code/network"
	"org.mutantcat.chickreomte/code/server/global"
)

func TestOldConnectionDoesNotCloseReplacement(t *testing.T) {
	h := New(&global.Configure{ReadTimeout: time.Millisecond, WriteTimeout: time.Millisecond})
	a, b := net.Pipe()
	defer b.Close()
	old := h.clis.new("same-id", network.NewConn(a))
	c, d := net.Pipe()
	defer d.Close()
	replacement := h.clis.new("same-id", network.NewConn(c))
	defer replacement.conn.Close()
	old.run()
	if h.clis.lookup("same-id") != replacement {
		t.Fatal("old connection removed its replacement")
	}
}
