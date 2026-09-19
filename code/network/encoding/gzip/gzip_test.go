package gzip

import (
	"bytes"
	stdgzip "compress/gzip"
	"io"
	"testing"
)

func TestRoundTripAllLevels(t *testing.T) {
	for level := 0; level <= stdgzip.BestCompression; level++ {
		c, err := New(level)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			var buf bytes.Buffer
			w, err := c.Compress(&buf)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = w.Write([]byte("round trip")); err != nil {
				t.Fatal(err)
			}
			if err = w.Close(); err != nil {
				t.Fatal(err)
			}
			r, err := c.Decompress(&buf)
			if err != nil {
				t.Fatal(err)
			}
			got, err := io.ReadAll(r)
			if err != nil {
				t.Fatal(err)
			}
			if err = r.Close(); err != nil {
				t.Fatal(err)
			}
			if string(got) != "round trip" {
				t.Fatalf("level %d: %q", level, got)
			}
		}
	}
}

func TestInvalidInput(t *testing.T) {
	c, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.Decompress(bytes.NewBufferString("invalid")); err == nil {
		t.Fatal("expected invalid gzip error")
	}
}
