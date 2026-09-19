package vnc

import (
	"image"
	"testing"
)

func TestDiffDoesNotReadOutsideOddWidthRect(t *testing.T) {
	a := image.NewRGBA(image.Rect(0, 0, 3, 2))
	b := image.NewRGBA(a.Rect)
	b.Pix[b.PixOffset(0, 1)] = 1
	if isDiff4(a, b, image.Rect(0, 0, 3, 1)) {
		t.Fatal("comparison read a pixel from the following row")
	}
}

func TestDiffSubimageStride(t *testing.T) {
	a := image.NewRGBA(image.Rect(0, 0, 8, 4))
	b := image.NewRGBA(a.Rect)
	rect := image.Rect(2, 1, 5, 3)
	b.Pix[b.PixOffset(4, 2)] = 1
	blocks := calcDiff(a.SubImage(rect).(*image.RGBA), b.SubImage(rect).(*image.RGBA))
	if len(blocks) != 1 || blocks[0] != rect {
		t.Fatalf("blocks = %v", blocks)
	}
}
