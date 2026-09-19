package vnc

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lwch/logging"
	"github.com/lwch/runtime"
	"org.mutantcat.chickreomte/code/client/conn"
	"org.mutantcat.chickreomte/code/network"
)

// New new vnc
func (v *VNC) New(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	if link := v.GetLink(); link != nil {
		link.Close(true)
	}
	q := r.FormValue("quality")
	s := r.FormValue("show_cursor")
	quality, err := strconv.ParseUint(q, 10, 32)
	if err != nil {
		quality = 50
	}
	showCursor, err := strconv.ParseBool(s)
	if err != nil {
		showCursor = false
	}
	id, err := runtime.UUID(16, "0123456789abcdef")
	if err != nil {
		logging.Error("failed to generate link_id for vnc: %s, err=%v",
			v.Name, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	link := v.NewLink(id, v.cfg.Target, nil, conn).(*Link)
	ready := false
	defer func() {
		if !ready {
			link.Close(true)
		}
	}()
	conn.SendConnectVnc(id, v.cfg, quality, showCursor)
	ch := conn.ChanRead(id)
	timeout := time.After(v.readTimeout)
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
			logging.Error("create vnc %s by rule %s failed, timtout", link.id, link.parent.Name)
			http.Error(w, "timeout", http.StatusBadGateway)
			return
		}
		if msg.GetXType() != network.Msg_connect_rep {
			conn.Requeue(id, msg)
			time.Sleep(v.readTimeout / 10)
			continue
		}
		rep := msg.GetCrep()
		if !rep.GetOk() {
			logging.Error("create vnc %s by rule %s failed, err=%s",
				link.id, link.parent.Name, rep.GetMsg())
			http.Error(w, rep.GetMsg(), http.StatusBadGateway)
			return
		}
		break
	}
	logging.Info("new vnc: name=%s, id=%s", v.Name, id)
	ready = true
	fmt.Fprint(w, id)
}
