package code

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"org.mutantcat.chickreomte/code/client/conn"
)

// Forward forward code-server requests
func (code *Code) Forward(conn *conn.Conn, w http.ResponseWriter, r *http.Request) {
	srcPath := r.URL.Path
	srcQuery := r.URL.Query()
	name := strings.TrimPrefix(r.URL.Path, "/forward/")
	end := strings.Index(name, "/")
	if end <= 0 || !strings.HasPrefix(r.URL.Path, "/forward/") {
		http.Error(w, "invalid workspace path", http.StatusBadRequest)
		return
	}
	name = name[:end]

	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/forward/"+name)
	if len(r.URL.Path) == 0 {
		r.URL.Path = "/"
	}

	var id string

	const argName = "chickreomte_connection_id"

	if r.URL.Path == "/" && len(r.FormValue(argName)) == 0 {
		var err error
		id, err = code.new(conn)
		if err != nil {
			logging.Error("can not create workspace for [%s]: %v", code.Name, err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		http.SetCookie(w, connectionCookie(r, id))
		srcQuery.Set(argName, id)
		http.Redirect(w, r, srcPath+"?"+srcQuery.Encode(), http.StatusTemporaryRedirect)
		return
	}

	cookie, err := r.Cookie("__CHICKREOMTE_CONNECTION_ID__")
	if err != nil {
		logging.Error("get connection id: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id = cookie.Value

	code.RLock()
	workspace := code.workspace[id]
	code.RUnlock()

	if workspace == nil {
		http.NotFound(w, r)
		return
	}

	if code.isWebsocket(r) {
		code.handleWebsocket(workspace, w, r)
	} else {
		code.handleRequest(conn, workspace, w, r)
	}
}

func connectionCookie(r *http.Request, id string) *http.Cookie {
	return &http.Cookie{
		Name:     "__CHICKREOMTE_CONNECTION_ID__",
		Value:    id,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Direct local HTTP is supported; do not trust arbitrary forwarded headers.
		Secure: r.TLS != nil,
	}
}

func (code *Code) isWebsocket(r *http.Request) bool {
	return websocket.IsWebSocketUpgrade(r)
}
