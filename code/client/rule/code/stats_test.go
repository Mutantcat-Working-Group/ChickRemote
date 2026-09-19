package code

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestConcurrentStatistics(t *testing.T) {
	ws := &Workspace{}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ws.recordSent(2)
				ws.recordReceived(3)
				ws.GetBytes()
				ws.GetPackets()
			}
		}()
	}
	wg.Wait()
	if recv, send := ws.GetBytes(); recv != 2400 || send != 1600 {
		t.Fatalf("bytes = %d, %d", recv, send)
	}
}

func TestRenderEscapesName(t *testing.T) {
	c := &Code{Name: "</title><script>alert(1)</script>"}
	w := httptest.NewRecorder()
	c.Render(nil, w, httptest.NewRequest("GET", "/", nil))
	if strings.Contains(w.Body.String(), "<script>alert(1)</script>") {
		t.Fatal("rule name rendered as executable HTML")
	}
	if !strings.Contains(w.Body.String(), "&lt;/title&gt;") {
		t.Fatal("escaped rule name missing")
	}
}
