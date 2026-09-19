package shell

import (
	"net/http/httptest"
	"testing"
)

func TestMissingLinks(t *testing.T) {
	s := &Shell{}
	for _, path := range []string{"/resize?id=missing", "/ws/missing"} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", path, nil)
		if path[1] == 'r' {
			s.Resize(nil, w, r)
		} else {
			s.WS(nil, w, r)
		}
		if w.Code != 404 {
			t.Fatalf("%s: status = %d", path, w.Code)
		}
	}
}

func TestResizeInvalidDimensions(t *testing.T) {
	s := &Shell{links: map[string]*Link{"link": {}}}
	for _, query := range []string{"rows=0&cols=80", "rows=-1&cols=80", "rows=24&cols=65536", "rows=abc&cols=80"} {
		w := httptest.NewRecorder()
		s.Resize(nil, w, httptest.NewRequest("GET", "/resize?id=link&"+query, nil))
		if w.Code != 400 {
			t.Fatalf("%s: status = %d", query, w.Code)
		}
	}
}
