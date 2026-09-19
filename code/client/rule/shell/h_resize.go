package shell

import (
	"fmt"
	"net/http"
	"strconv"

	"org.mutantcat.chickreomte/code/client/conn"
)

// Resize resize terminal
func (shell *Shell) Resize(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	rows := r.FormValue("rows")
	cols := r.FormValue("cols")

	shell.RLock()
	link := shell.links[id]
	shell.RUnlock()
	if link == nil {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}

	nRows, rowErr := strconv.ParseUint(rows, 10, 16)
	nCols, colErr := strconv.ParseUint(cols, 10, 16)
	if rowErr != nil || colErr != nil || nRows == 0 || nCols == 0 {
		http.Error(w, "invalid terminal size", http.StatusBadRequest)
		return
	}

	link.SendResize(uint32(nRows), uint32(nCols))

	fmt.Fprint(w, "ok")
}
