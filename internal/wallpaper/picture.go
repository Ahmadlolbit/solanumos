package wallpaper

import (
	"bufio"
	"image"
	"io"
	"os"
)

func IsPicture(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	head := make([]byte, 12)
	if count, _ := io.ReadFull(file, head); count == len(head) && string(head[:4]) == "RIFF" && string(head[8:]) == "WEBP" {
		return true
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return false
	}
	_, _, err = image.DecodeConfig(bufio.NewReader(file))
	return err == nil
}
