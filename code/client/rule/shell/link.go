package shell

import (
	"io"
	"os"
	"sync"

	"github.com/lwch/logging"
	"golang.org/x/text/encoding/simplifiedchinese"
	"google.golang.org/protobuf/proto"
	"org.mutantcat.chickreomte/code/client/conn"
	"org.mutantcat.chickreomte/code/network"
	"org.mutantcat.chickreomte/code/utils"
)

// Link shell link
type Link struct {
	closeOnce sync.Once
	parent    *Shell
	id        string // link id
	target    string // target id
	remote    *conn.Conn
	// in remote
	pid    int
	stdin  io.WriteCloser
	stdout io.ReadCloser
	// runtime
	statsMu    sync.RWMutex
	sendBytes  uint64
	recvBytes  uint64
	sendPacket uint64
	recvPacket uint64
}

// GetID get link id
func (link *Link) GetID() string {
	return link.id
}

// GetBytes get send and recv bytes
func (link *Link) GetBytes() (uint64, uint64) {
	link.statsMu.RLock()
	defer link.statsMu.RUnlock()
	return link.recvBytes, link.sendBytes
}

// GetPackets get send and recv packets
func (link *Link) GetPackets() (uint64, uint64) {
	link.statsMu.RLock()
	defer link.statsMu.RUnlock()
	return link.recvPacket, link.sendPacket
}

func (link *Link) recordSent(n uint64) {
	link.statsMu.Lock()
	defer link.statsMu.Unlock()
	link.sendBytes += n
	link.sendPacket++
}

func (link *Link) recordReceived(n uint64) {
	link.statsMu.Lock()
	defer link.statsMu.Unlock()
	link.recvBytes += n
	link.recvPacket++
}

// Close close link
func (link *Link) Close(send bool) {
	link.closeOnce.Do(func() {
		link.onClose()
		if link.pid > 0 {
			p, err := os.FindProcess(link.pid)
			if err == nil {
				p.Kill()
			}
		}
		if send {
			link.remote.SendDisconnect(link.target, link.id)
		}
		link.parent.remove(link.id)
		link.remote.ChanClose(link.id)
	})
}

// Forward forward data
func (link *Link) Forward() {
	go link.remoteRead()
	go link.localRead()
}

func (link *Link) remoteRead() {
	defer utils.Recover("remoteRead")
	defer link.Close(true)
	ch := link.remote.ChanRead(link.id)
	for {
		msg := <-ch
		if msg == nil {
			return
		}
		data, _ := proto.Marshal(msg)
		link.recordReceived(uint64(len(data)))
		switch msg.GetXType() {
		case network.Msg_shell_resize:
			size := msg.GetSresize()
			link.resize(size.GetRows(), size.GetCols())
		case network.Msg_shell_data:
			_, err := link.stdin.Write(msg.GetSdata().GetData())
			if err != nil {
				logging.Error("write data on shell %s link %s failed, err=%v",
					link.parent.Name, link.id, err)
				return
			}
		}
	}
}

func (link *Link) localRead() {
	defer utils.Recover("localRead")
	defer link.Close(true)
	buf := make([]byte, 16*1024)
	for {
		n, err := link.stdout.Read(buf)
		if err != nil {
			logging.Error("read data on shell %s link %s failed, err=%v",
				link.parent.Name, link.id, err)
			return
		}
		if n == 0 {
			continue
		}
		var data []byte
		switch {
		case isUtf8(buf[:n]):
			data = buf[:n]
		case isGBK(buf[:n]):
			data, err = simplifiedchinese.GBK.NewDecoder().Bytes(buf[:n])
			if err != nil {
				logging.Error("transform gbk to utf8 failed: %v", err)
				continue
			}
		}
		logging.Debug("link %s on shell %s read from local %d bytes",
			link.id, link.parent.Name, n)
		send := link.remote.SendShellData(link.target, link.id, data)
		link.recordSent(send)
	}
}

// SendData send data
func (link *Link) SendData(data []byte) {
	send := link.remote.SendShellData(link.target, link.id, data)
	link.recordSent(send)
}

// SendResize send resize message
func (link *Link) SendResize(rows, cols uint32) {
	link.remote.SendShellResize(link.target, link.id, rows, cols)
}
