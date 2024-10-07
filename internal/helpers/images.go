package helpers

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/disintegration/gift"
)

func ResizeImage(path string, maxWidth int) (string, error) {
	img, err := loadImage(path)
	if err != nil {
		return "", fmt.Errorf("loadImage failed: %w", err)
	}

	to := PathInsertSize(path, maxWidth)

	filters := []gift.Filter{}

	if maxWidth != 0 {
		filters = append(filters, gift.Resize(maxWidth, 0, gift.LanczosResampling))
	}

	g := gift.New(filters...)

	dst := image.NewNRGBA(g.Bounds(img.Bounds()))
	g.Draw(dst, img)

	return to, saveImage(PathInsertSize(path, maxWidth), dst)
}

func ImageSize(path string) (int, int, error) {
	img, err := loadImage(path)
	if err != nil {
		return 0, 0, fmt.Errorf("loadImage failed: %w", err)
	}

	return img.Bounds().Dx(), img.Bounds().Dy(), nil
}

func loadImage(filename string) (img image.Image, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return img, fmt.Errorf("os.Open failed: %w", err)
	}
	defer f.Close()

	img, _, err = image.Decode(f)
	if err != nil {
		return img, fmt.Errorf("image.Decode failed: %w", err)
	}

	return img, nil
}

func saveImage(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("os.Create failed: %w", err)
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		return fmt.Errorf("png.Encode failed: %w", err)
	}

	return nil
}
