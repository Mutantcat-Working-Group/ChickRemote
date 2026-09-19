package vnc

import (
	"encoding/json"

	"github.com/lwch/logging"
)

func (link *Link) mouseEvent(data []byte) {
	var payload struct {
		Payload struct {
			Button string `json:"button"`
			Status string `json:"status"`
			X      int    `json:"x"`
			Y      int    `json:"y"`
		} `json:"payload"`
	}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		logging.Error("unmarshal: %v", err)
		return
	}
	link.remote.SendVNCMouse(link.target, link.id,
		payload.Payload.Button, payload.Payload.Status, payload.Payload.X, payload.Payload.Y)
}

func (link *Link) keyboardEvent(data []byte) {
	var payload struct {
		Payload struct {
			Status string `json:"status"`
			Key    string `json:"key"`
		} `json:"payload"`
	}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		logging.Error("unmarshal: %v", err)
		return
	}
	link.remote.SendVNCKeyboard(link.target, link.id,
		payload.Payload.Status, payload.Payload.Key)
}

func (link *Link) cadEvent() {
	link.remote.SendVNCCADEvent(link.target, link.id)
}

func (link *Link) scrollEvent(data []byte) {
	var payload struct {
		Payload struct {
			X int32 `json:"x"`
			Y int32 `json:"y"`
		} `json:"payload"`
	}
	err := json.Unmarshal(data, &payload)
	if err != nil {
		logging.Error("unmarshal: %v", err)
		return
	}
	link.remote.SendVNCScroll(link.target, link.id,
		payload.Payload.X, payload.Payload.Y)
}
