package code

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForwardMalformedPath(t *testing.T) {
	for _, path := range []string{"/forward/name", "/forward/", "/forward//"} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			(&Code{}).Forward(nil, w, httptest.NewRequest("GET", path, nil))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d", w.Code)
			}
		})
	}
}

func TestWebsocketHeaderTokens(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Connection", "keep-alive, upgrade")
	r.Header.Set("Upgrade", "websocket")
	if !(&Code{}).isWebsocket(r) {
		t.Fatal("valid upgrade not recognized")
	}
	r.Header.Del("Upgrade")
	if (&Code{}).isWebsocket(r) {
		t.Fatal("accepted missing websocket header")
	}
}
