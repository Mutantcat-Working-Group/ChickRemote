package process

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lwch/logging"
	"google.golang.org/protobuf/proto"
	"org.mutantcat.chickreomte/code/client/rule/vnc/vncnetwork"
	"org.mutantcat.chickreomte/code/utils"
)

const (
	listenBegin = 6155
	listenEnd   = 6955
)

// Process process
type Process struct {
	// Keep the atomic PID aligned on 32-bit platforms.
	pid         int64
	initOnce    sync.Once
	closeOnce   sync.Once
	done        chan struct{}
	captureMu   sync.Mutex
	srv         *http.Server
	chWrite     chan *vncnetwork.VncMsg
	chImage     chan *vncnetwork.ImageData
	chClipboard chan *vncnetwork.ClipboardData
}

func (p *Process) doneChan() <-chan struct{} {
	p.initOnce.Do(func() { p.done = make(chan struct{}) })
	return p.done
}

func (p *Process) send(msg *vncnetwork.VncMsg) bool {
	select {
	case p.chWrite <- msg:
		return true
	case <-p.doneChan():
		return false
	}
}

func (p *Process) listenAndServe() (uint16, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.ws)
	port := uint16(listenBegin)
	for {
		if port > listenEnd {
			return 0, errors.New("no port available")
		}
		p.srv = &http.Server{
			Addr:    fmt.Sprintf("127.0.0.1:%d", port),
			Handler: mux,
		}
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			port++
			continue
		}
		go p.srv.Serve(ln)
		return port, nil
	}
}

var upgrader = websocket.Upgrader{EnableCompression: true}

func (p *Process) ws(w http.ResponseWriter, r *http.Request) {
	logging.Info("child process connected")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	defer p.Close()
	go func() {
		<-p.doneChan()
		conn.Close()
	}()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer p.Close()
		defer utils.Recover("ws read")
		defer wg.Done()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				logging.Error("read message: %v", err)
				return
			}
			var msg vncnetwork.VncMsg
			err = proto.Unmarshal(data, &msg)
			if err != nil {
				continue
			}
			switch msg.GetXType() {
			case vncnetwork.VncMsg_capture_data:
				select {
				case p.chImage <- msg.GetData():
				case <-p.doneChan():
					return
				}
			case vncnetwork.VncMsg_clipboard_event:
				select {
				case p.chClipboard <- msg.GetClipboard():
				case <-p.doneChan():
					return
				}
			default:
			}
		}
	}()
	go func() {
		defer utils.Recover("ws write")
		defer p.Close()
		defer wg.Done()
		for {
			var msg *vncnetwork.VncMsg
			select {
			case msg = <-p.chWrite:
			case <-p.doneChan():
				return
			}
			data, err := proto.Marshal(msg)
			if err != nil {
				continue
			}
			err = conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				logging.Error("write message: %v", err)
				return
			}
		}
	}()
	wg.Wait()
}

func (p *Process) kill() {
	pid := atomic.LoadInt64(&p.pid)
	if pid <= 0 {
		return
	}
	ps, _ := os.FindProcess(int(pid))
	if ps != nil {
		ps.Kill()
	}
}

// Close close process
func (p *Process) Close() {
	p.doneChan()
	p.closeOnce.Do(func() {
		close(p.done)
		if p.srv != nil {
			p.srv.Close()
		}
		p.kill()
	})
}

// Capture capture desktop image
func (p *Process) Capture(timeout time.Duration) (*image.RGBA, error) {
	p.captureMu.Lock()
	defer p.captureMu.Unlock()
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	var msg vncnetwork.VncMsg
	msg.XType = vncnetwork.VncMsg_capture_req
	select {
	case p.chWrite <- &msg:
	case <-timer.C:
		return nil, errors.New("capture timeout")
	case <-p.doneChan():
		return nil, net.ErrClosed
	}
	select {
	case data := <-p.chImage:
		if !data.GetOk() {
			return nil, fmt.Errorf("capture failed: %s", data.GetMsg())
		}
		w, h := uint64(data.GetWidth()), uint64(data.GetHeight())
		if w == 0 || h == 0 || w > 32768 || h > 32768 || w*h*4 != uint64(len(data.GetData())) {
			return nil, errors.New("invalid capture dimensions")
		}
		img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
		copy(img.Pix, data.GetData())
		return img, nil
	case <-timer.C:
		return nil, errors.New("capture timeout")
	case <-p.doneChan():
		return nil, net.ErrClosed
	}
}

func dumpImage(img image.Image) {
	f, err := os.Create(`C:\Users\lwch\Pictures\debug.jpeg`)
	if err != nil {
		logging.Error("debug: %v", err)
		return
	}
	defer f.Close()
	err = jpeg.Encode(f, img, nil)
	if err != nil {
		logging.Error("encode: %v", err)
		return
	}
}
