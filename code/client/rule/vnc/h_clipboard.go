package vnc

import (
	"fmt"
	"net/http"
	"time"

	"org.mutantcat.chickreomte/code/client/conn"
)

// Clipboard get/set clipboard
func (v *VNC) Clipboard(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		v.getClipboard(conn, w, r)
		return
	}
	v.setClipboard(conn, w, r)
}

func (v *VNC) getClipboard(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	link := v.GetLink()
	if link == nil {
		http.NotFound(w, r)
		return
	}
	conn.SendVNCClipboardData(link.target, link.id, false, "")
	select {
	case data := <-v.chClipboard:
		fmt.Fprint(w, data.GetData())
	case <-r.Context().Done():
	case <-time.After(v.readTimeout):
		http.Error(w, "clipboard timeout", http.StatusGatewayTimeout)
	}
}

func (v *VNC) setClipboard(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	data := r.FormValue("data")
	link := v.GetLink()
	if link == nil {
		http.NotFound(w, r)
		return
	}
	conn.SendVNCClipboardData(link.target, link.id, true, data)
	fmt.Fprint(w, "ok")
}
