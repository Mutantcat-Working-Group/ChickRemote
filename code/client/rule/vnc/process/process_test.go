package process

import (
	"sync"
	"testing"
	"time"

	"org.mutantcat.chickreomte/code/client/rule/vnc/vncnetwork"
)

func TestCaptureTimeoutIncludesSend(t *testing.T) {
	p := &Process{chWrite: make(chan *vncnetwork.VncMsg), chImage: make(chan *vncnetwork.ImageData)}
	done := make(chan error, 1)
	go func() { _, err := p.Capture(10 * time.Millisecond); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("missing timeout error")
		}
	case <-time.After(time.Second):
		t.Fatal("capture blocked while waiting for a worker")
	}
}

func TestConcurrentCloseUnblocksCapture(t *testing.T) {
	p := &Process{chWrite: make(chan *vncnetwork.VncMsg), chImage: make(chan *vncnetwork.ImageData)}
	done := make(chan error, 1)
	go func() { _, err := p.Capture(time.Second); done <- err }()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); p.Close() }()
	}
	wg.Wait()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected closed error")
		}
	case <-time.After(time.Second):
		t.Fatal("capture did not stop")
	}
}

func TestCaptureRejectsFailedFrame(t *testing.T) {
	p := &Process{chWrite: make(chan *vncnetwork.VncMsg, 1), chImage: make(chan *vncnetwork.ImageData, 1)}
	p.chImage <- &vncnetwork.ImageData{Ok: false, Msg: "capture failed"}
	if _, err := p.Capture(time.Second); err == nil {
		t.Fatal("accepted failed capture")
	}
}
