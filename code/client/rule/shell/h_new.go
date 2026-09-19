package shell

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
	"org.mutantcat.chickreomte/code/client/conn"
	"org.mutantcat.chickreomte/code/network"
)

// New new shell
func (shell *Shell) New(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	id, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		logging.Error("failed to generate link_id for shell: %s, err=%v",
			shell.Name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	link := shell.NewLink(id, shell.cfg.Target, nil, conn).(*Link)
	ready := false
	defer func() {
		if !ready {
			link.Close(true)
		}
	}()
	conn.SendConnectReq(id, shell.cfg)
	ch := conn.ChanRead(id)
	timeout := time.After(shell.readTimeout)
	var repMsg *network.Msg
	for {
		var msg *network.Msg
		select {
		case msg = <-ch:
			if msg == nil {
				http.Error(w, "connection closed", http.StatusBadGateway)
				return
			}
		case <-r.Context().Done():
			return
		case <-timeout:
			logging.Error("create shell %s by rule %s failed, timtout", link.id, link.parent.Name)
			http.Error(w, "timeout", http.StatusBadGateway)
			return
		}
		if msg.GetXType() != network.Msg_connect_rep {
			conn.Requeue(id, msg)
			time.Sleep(shell.readTimeout / 10)
			continue
		}
		rep := msg.GetCrep()
		if !rep.GetOk() {
			logging.Error("create shell %s by rule %s failed, err=%s",
				link.id, link.parent.Name, rep.GetMsg())
			http.Error(w, rep.GetMsg(), http.StatusBadGateway)
			return
		}
		repMsg = msg
		break
	}
	logging.Info("create link %s for shell rule [%s] from %s to %s",
		link.GetID(), shell.cfg.Name,
		repMsg.GetTo(), repMsg.GetFrom())
	ready = true
	fmt.Fprint(w, id)
}
