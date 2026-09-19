package vnc

import (
	"image"
	"image/color"
	"net/http/httptest"
	"testing"
)

func TestCutPreservesLastPixelAndStride(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 8, 4))
	rect := image.Rect(2, 1, 5, 3)
	src.SetRGBA(4, 2, color.RGBA{1, 2, 3, 255})
	got := cut(src.SubImage(rect).(*image.RGBA), rect)
	if got.RGBAAt(2, 1) != (color.RGBA{1, 2, 3, 255}) {
		t.Fatal("last pixel was truncated")
	}
}

func TestRemoveOldLinkKeepsReplacement(t *testing.T) {
	v := &VNC{link: &Link{id: "new"}}
	v.remove("old")
	if v.GetLink() == nil {
		t.Fatal("removed replacement")
	}
}

func TestWebsocketRejectsWrongLink(t *testing.T) {
	v := &VNC{link: &Link{id: "new"}}
	w := httptest.NewRecorder()
	v.WS(nil, w, httptest.NewRequest("GET", "/ws/old", nil))
	if w.Code != 404 {
		t.Fatalf("status = %d", w.Code)
	}
}
