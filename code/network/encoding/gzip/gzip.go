package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/lwch/runtime"
	"org.mutantcat.chickreomte/code/network/encoding"
)

type writer struct {
	*gzip.Writer
	pool *sync.Pool
}

// Close close write and put writer to pool
func (w *writer) Close() error {
	err := w.Writer.Close()
	w.pool.Put(w)
	return err
}

type reader struct {
	*gzip.Reader
	pool *sync.Pool
}

// Close close reader and put reader to pool
func (r *reader) Close() error {
	err := r.Reader.Close()
	r.pool.Put(r)
	return err
}

type compressor struct {
	level      int32
	poolWriter [gzip.BestCompression + 1]sync.Pool
	poolReader sync.Pool
}

// New create compressor
func New(level ...int) (encoding.Compressor, error) {
	if len(level) > 0 {
		if level[0] < 0 || level[0] > gzip.BestCompression {
			return nil, fmt.Errorf("invalid gzip compress level: %d", level[0])
		}
	} else {
		level = append(level, 6)
	}
	ret := new(compressor)
	ret.level = int32(level[0])
	for i := 0; i <= gzip.BestCompression; i++ {
		i := i
		ret.poolWriter[i].New = func() interface{} {
			w, err := gzip.NewWriterLevel(io.Discard, i)
			runtime.Assert(err)
			return &writer{Writer: w, pool: &ret.poolWriter[i]}
		}
	}
	return ret, nil
}

// Compress gzip compress
func (c *compressor) Compress(w io.Writer) (io.WriteCloser, error) {
	pw := c.poolWriter[atomic.LoadInt32(&c.level)].Get().(*writer)
	pw.Writer.Reset(w)
	return pw, nil
}

// Decompress gzip decompress
func (c *compressor) Decompress(r io.Reader) (io.ReadCloser, error) {
	if cached := c.poolReader.Get(); cached != nil {
		pr := cached.(*reader)
		if err := pr.Reader.Reset(r); err != nil {
			c.poolReader.Put(pr)
			return nil, err
		}
		return pr, nil
	}
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &reader{Reader: gr, pool: &c.poolReader}, nil
}

// SetLevel set compress level
func (c *compressor) SetLevel(level int) error {
	if level < 0 || level > gzip.BestCompression {
		return fmt.Errorf("invalid gzip compress level: %d", level)
	}
	atomic.StoreInt32(&c.level, int32(level))
	return nil
}
