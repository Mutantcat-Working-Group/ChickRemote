package vnc

import (
	"bytes"
	"image"
)

func calcDiff(src, dst *image.RGBA) []image.Rectangle {
	// 宽度必须为2的倍数
	const width = 64
	const height = 64
	size := dst.Bounds()
	ret := make([]image.Rectangle, 0, (size.Max.X*size.Max.Y)/(width*height))
	for y := size.Min.Y; y < size.Max.Y; y += height {
		for x := size.Min.X; x < size.Max.X; x += width {
			dWidth := size.Max.X - x
			dHeight := size.Max.Y - y
			if dWidth > width {
				dWidth = width
			}
			if dHeight > height {
				dHeight = height
			}
			rect := image.Rect(x, y, x+dWidth, y+dHeight)
			if dWidth%2 == 0 {
				if isDiff8(src, dst, rect) {
					ret = append(ret, rect)
				}
			} else {
				if isDiff4(src, dst, rect) {
					ret = append(ret, rect)
				}
			}
			////////
		}
	}
	return ret
}

func isDiff8(src, dst *image.RGBA, rect image.Rectangle) bool {
	return isDiff4(src, dst, rect)
}

func isDiff4(src, dst *image.RGBA, rect image.Rectangle) bool {
	if !rect.In(src.Rect) || !rect.In(dst.Rect) {
		return true
	}
	width := rect.Dx() * 4
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		si, di := src.PixOffset(rect.Min.X, y), dst.PixOffset(rect.Min.X, y)
		if !bytes.Equal(src.Pix[si:si+width], dst.Pix[di:di+width]) {
			return true
		}
	}
	return false
}
