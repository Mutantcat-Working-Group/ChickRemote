package conn

import (
	"time"

	"org.mutantcat.chickreomte/code/network"
)

// SendKeepalive send keepalive message
func (conn *Conn) SendKeepalive() {
	var msg network.Msg
	msg.To = "server"
	msg.XType = network.Msg_keepalive
	select {
	case conn.write <- &msg:
	case <-time.After(conn.cfg.WriteTimeout):
	}
}
