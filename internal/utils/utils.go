package utils

import (
	"bytes"
	"image"
	"image/jpeg"
	"golang.org/x/image/draw"
)

const (
	MaxImageSizeBytes = 5 * 1024 * 1024 // 2 MB
	MaxImageWidth = 1200 // cap width at 1200px, height scales proportionally
)

func ResizeIfLarge(img image.Image, maxWidth int) image.Image {
	bounds := img.Bounds()
	if bounds.Dx() <= maxWidth {
		return img
	}

	newHeight := bounds.Dy() * maxWidth / bounds.Dx()
	dst := image.NewRGBA(image.Rect(0, 0, maxWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst
}

func CompressJpeg(img image.Image, maxSize int) ([]byte, error) {
	quality := 90
	var buf bytes.Buffer

	for quality >= 10 {
		buf.Reset()
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}

		if buf.Len() <= maxSize {
			return buf.Bytes(), nil
		}
		quality -= 10
	}
	return buf.Bytes(), nil
}
