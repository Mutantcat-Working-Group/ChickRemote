package handler

import (
	"sync"
	"time"

	"org.mutantcat.chickreomte/code/network"
)

type clients struct {
	sync.RWMutex
	parent *Handler
	data   map[string]*client // id => client
}

func newClients(parent *Handler) *clients {
	return &clients{
		parent: parent,
		data:   make(map[string]*client),
	}
}

func (cs *clients) new(id string, conn *network.Conn) *client {
	cli := &client{
		id:      id,
		parent:  cs,
		conn:    conn,
		updated: time.Now(),
		links:   make(map[string]struct{}),
	}
	cs.Lock()
	old := cs.data[id]
	cs.data[id] = cli
	cs.Unlock()
	if old != nil {
		old.close()
	}
	return cli
}

func (cs *clients) lookup(id string) *client {
	cs.RLock()
	defer cs.RUnlock()
	return cs.data[id]
}

func (cs *clients) close(cli *client) {
	cs.Lock()
	if cs.data[cli.id] == cli {
		delete(cs.data, cli.id)
	}
	cs.Unlock()
	cli.close()
}
