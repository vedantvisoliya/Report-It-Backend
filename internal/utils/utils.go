package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"

	"golang.org/x/image/draw"
)

const (
	MaxImageSizeBytes = 5 * 1024 * 1024 // 2 MB
	MaxImageWidth     = 1200            // cap width at 1200px, height scales proportionally
)

var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
}

func DetectFileContentType(file multipart.File) (string, error) {
	buf := make([]byte, 512)

	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for type detection (%w)", err)
	}

	contentType := http.DetectContentType(buf[:n])

	// Reset reader to the beginning so subsequent reads (e.g. image.Decode) work.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file reader (%w)", err)
	}

	return contentType, nil
}

func ValidateImageType(fileHeader *multipart.FileHeader) (multipart.File, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", fmt.Errorf("failed to open file (%w)", err)
	}

	contentType, err := DetectFileContentType(file)
	if err != nil {
		file.Close()
		return nil, "", err
	}

	if !AllowedImageTypes[contentType] {
		file.Close()
		return nil, "", fmt.Errorf("unsupported file type: %s (only JPEG images are accepted)", contentType)
	}

	return file, contentType, nil
}

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
