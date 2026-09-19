package vnc

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"sync"
	"time"

	"github.com/lwch/logging"
	"org.mutantcat.chickreomte/code/client/conn"
	"org.mutantcat.chickreomte/code/client/rule/vnc/process"
	"org.mutantcat.chickreomte/code/network"
	"org.mutantcat.chickreomte/code/utils"
)

const (
	zoneWidth  = 64
	zoneHeight = 64
)

// Link vnc link
type Link struct {
	closeOnce  sync.Once
	settingsMu sync.Mutex
	parent     *VNC
	id         string // link id
	target     string // target id
	remote     *conn.Conn
	// vnc
	ps         *process.Process
	quality    uint32
	showCursor bool
	img        *image.RGBA
	// runtime
	sendBytes  uint64
	recvBytes  uint64
	sendPacket uint64
	recvPacket uint64
	idx        int
	reDraw     bool
}

// GetID get link id
func (link *Link) GetID() string {
	return link.id
}

// GetBytes get send and recv bytes
func (link *Link) GetBytes() (uint64, uint64) {
	return link.recvBytes, link.sendBytes
}

// GetPackets get send and recv packets
func (link *Link) GetPackets() (uint64, uint64) {
	return link.recvPacket, link.sendPacket
}

// SetQuality transfer quality
func (link *Link) SetQuality(q uint32) {
	link.settingsMu.Lock()
	defer link.settingsMu.Unlock()
	if q > 100 {
		q = 100
	}
	link.quality = q
	link.reDraw = true
}

// SetCursor set show cursor
func (link *Link) SetCursor(b bool) {
	link.settingsMu.Lock()
	defer link.settingsMu.Unlock()
	link.showCursor = b
	link.reDraw = true
}

// Fork fork worker process
func (link *Link) Fork(confDir string) error {
	p, err := process.CreateWorker(link.parent.Name, confDir, link.showCursor)
	if err != nil {
		return err
	}
	link.ps = p
	return nil
}

// Forward forward data
func (link *Link) Forward() {
	if link.ps == nil {
		link.Close(true)
		return
	}
	go link.remoteRead()
	go link.localRead()
}

func (link *Link) remoteRead() {
	defer link.Close(true)
	ch := link.remote.ChanRead(link.id)
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		switch msg.GetXType() {
		case network.Msg_vnc_ctrl:
			ctrl := msg.GetVctrl()
			link.SetQuality(ctrl.GetQuality())
			link.SetCursor(ctrl.GetCursor())
			link.ps.SetCursor(ctrl.GetCursor())
		case network.Msg_vnc_mouse:
			link.ps.MouseEvent(msg.GetVmouse())
		case network.Msg_vnc_keyboard:
			link.ps.KeyboardEvent(msg.GetVkbd())
		case network.Msg_vnc_cad:
			link.ps.CADEvent()
		case network.Msg_vnc_scroll:
			link.ps.ScrollEvent(msg.GetVscroll())
		case network.Msg_vnc_clipboard:
			if msg.GetVclipboard().GetSet() {
				link.ps.SetClipboard(msg.GetVclipboard())
			} else {
				data := link.ps.GetClipboard()
				link.remote.SendVNCClipboardData(link.target, link.id, true, data)
			}
		}
	}
}

func (link *Link) localRead() {
	// TODO: exit by context
	defer utils.Recover("capture")
	defer link.Close(true)
	img, err := link.ps.Capture(3 * time.Second)
	if err != nil {
		logging.Error("capture: %v", err)
		return
	}
	link.sendAll(img)
	link.img = img
	size := img.Rect
	fps := link.parent.cfg.Fps
	if fps == 0 {
		fps = 10
	}
	if fps > 50 {
		fps = 50
	}
	sleep := time.Second / time.Duration(fps)
	for {
		time.Sleep(sleep)
		img, err = link.ps.Capture(0)
		if err != nil {
			logging.Error("capture: %v", err)
			return
		}
		link.settingsMu.Lock()
		redraw := link.reDraw
		link.reDraw = false
		link.settingsMu.Unlock()
		if img.Rect.Dx() != size.Dx() ||
			img.Rect.Dy() != size.Dy() ||
			redraw ||
			link.idx%10000 == 0 {
			link.sendAll(img)
		} else {
			link.sendDiff(img)
		}
		link.img = img
		size = img.Rect
		link.idx++
	}
}

// Close close link
func (link *Link) Close(send bool) {
	link.closeOnce.Do(func() {
		if link.ps != nil {
			link.ps.Close()
		}
		if send {
			link.remote.SendDisconnect(link.target, link.id)
		}
		link.parent.remove(link.id)
		link.remote.ChanClose(link.id)
	})
}

func cut(src *image.RGBA, rect image.Rectangle) *image.RGBA {
	size := rect.Size()
	ret := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	draw.Draw(ret, ret.Bounds(), src, rect.Min, draw.Src)
	return ret
}

func (link *Link) sendAll(img *image.RGBA) {
	link.settingsMu.Lock()
	quality := link.quality
	link.settingsMu.Unlock()
	size := img.Bounds()
	screen := image.Rect(0, 0, img.Rect.Dx(), img.Rect.Dy())
	var buf bytes.Buffer
	for y := 0; y < size.Max.Y; y += zoneHeight {
		for x := 0; x < size.Max.X; x += zoneWidth {
			width := size.Max.X - x
			height := size.Max.Y - y
			if width > zoneWidth {
				width = zoneWidth
			}
			if height > zoneHeight {
				height = zoneHeight
			}
			rect := image.Rect(x, y, x+width, y+height)
			next := cut(img, rect)
			if quality == 100 {
				link.remote.SendVNCImage(link.target, link.id,
					screen, rect, network.VncImage_raw, next.Pix)
				continue
			}
			buf.Reset()
			err := jpeg.Encode(&buf, next, &jpeg.Options{Quality: int(quality)})
			if err == nil {
				link.remote.SendVNCImage(link.target, link.id,
					screen, rect, network.VncImage_jpeg, buf.Bytes())
			} else {
				link.remote.SendVNCImage(link.target, link.id,
					screen, rect, network.VncImage_raw, next.Pix)
			}
		}
	}
}

func (link *Link) sendDiff(img *image.RGBA) {
	link.settingsMu.Lock()
	quality := link.quality
	link.settingsMu.Unlock()
	blocks := calcDiff(link.img, img)
	screen := image.Rect(0, 0, img.Rect.Dx(), img.Rect.Dy())
	var buf bytes.Buffer
	for _, block := range blocks {
		next := cut(img, block)
		if quality == 100 {
			link.remote.SendVNCImage(link.target, link.id,
				screen, block, network.VncImage_raw, next.Pix)
			continue
		}
		buf.Reset()
		err := jpeg.Encode(&buf, next, &jpeg.Options{Quality: int(quality)})
		if err == nil {
			link.remote.SendVNCImage(link.target, link.id,
				screen, block, network.VncImage_jpeg, buf.Bytes())
		} else {
			link.remote.SendVNCImage(link.target, link.id,
				screen, block, network.VncImage_raw, next.Pix)
		}
	}
}
