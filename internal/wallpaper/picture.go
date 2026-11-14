package wallpaper

import (
	"bufio"
	"image"
	"io"
	"os"
)

func PictureKind(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	head := make([]byte, 12)
	if count, _ := io.ReadFull(file, head); count == len(head) && string(head[:4]) == "RIFF" && string(head[8:]) == "WEBP" {
		return ".webp"
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return ""
	}
	_, format, err := image.DecodeConfig(bufio.NewReader(file))
	switch {
	case err != nil:
		return ""
	case format == "jpeg":
		return ".jpg"
	}
	return "." + format
}

func IsPicture(path string) bool {
	return PictureKind(path) != ""
}
