package shell

import (
	"sync"
	"testing"
)

func TestConcurrentStatistics(t *testing.T) {
	link := &Link{}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				link.recordSent(2)
				link.recordReceived(3)
				link.GetBytes()
				link.GetPackets()
			}
		}()
	}
	wg.Wait()
	if recv, send := link.GetBytes(); recv != 2400 || send != 1600 {
		t.Fatalf("bytes = %d, %d", recv, send)
	}
}
