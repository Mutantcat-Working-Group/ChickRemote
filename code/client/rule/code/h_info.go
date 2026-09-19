package code

import (
	"encoding/json"
	"net/http"

	"github.com/lwch/logging"
)

// Info get workspace info
func (code *Code) Info(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	code.RLock()
	workspace := code.workspace[id]
	code.RUnlock()
	if workspace == nil {
		http.NotFound(w, r)
		return
	}
	recvBytes, sendBytes := workspace.GetBytes()
	recvPackets, sendPackets := workspace.GetPackets()
	data, err := json.Marshal(map[string]interface{}{
		"name":        code.Name,
		"send_bytes":  sendBytes,
		"send_packet": sendPackets,
		"recv_bytes":  recvBytes,
		"recv_packet": recvPackets,
	})
	if err != nil {
		logging.Error("marshal: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
