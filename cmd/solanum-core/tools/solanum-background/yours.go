package background

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Ahmadlolbit/solanumos/internal/atomic"
	"github.com/Ahmadlolbit/solanumos/internal/bus"
	"github.com/Ahmadlolbit/solanumos/internal/env"
	"github.com/Ahmadlolbit/solanumos/internal/fsutil"
	"github.com/Ahmadlolbit/solanumos/internal/look"
	"github.com/Ahmadlolbit/solanumos/internal/wallpaper"
)

var (
	pictureKinds = []string{"image/jpeg", "image/png", "image/webp"}
	yourCommands = map[string]func([]string) int{"add": addPictures, "remove": removePicture, "keep": keepDrawing}
	choose       = bus.ChoosePictures
)

func themeName() string {
	name, _ := fsutil.Text(filepath.Join(state(), "theme.name"))
	return name
}

func yourFolder(every bool) string {
	if every {
		return look.Everywhere()
	}
	return look.OwnBackgrounds(themeName())
}

func addOne(folder, file string) (string, error) {
	kind := wallpaper.PictureKind(file)
	if kind == "" {
		return "", errors.New("it isn't a picture")
	}
	data, err := os.ReadFile(file)
	if err == nil {
		err = os.MkdirAll(folder, 0o777)
	}
	var path *fs.PathError
	if errors.As(err, &path) {
		return "", path.Err
	}
	if err != nil {
		return "", err
	}
	kept, err := atomic.Fresh(folder, strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)), kind, 0o644)
	if err == nil {
		err = atomic.Write(kept, data, 0o644)
	}
	return kept, err
}

func chosenFiles(args []string) ([]string, int) {
	if len(args) > 0 {
		return args, 0
	}
	chosen, err := choose(pictureKinds)
	switch {
	case errors.Is(err, bus.ErrNothingChosen):
		return nil, 0
	case err != nil:
		fmt.Fprintf(os.Stderr, "Couldn't open the file chooser: %v\n", err)
		return nil, 1
	}
	return chosen, 0
}

func addPictures(args []string) int {
	every := len(args) > 0 && args[0] == "--every"
	if every {
		args = args[1:]
	}
	files, code := chosenFiles(args)
	added := 0
	for _, file := range files {
		path, err := addOne(yourFolder(every), file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Left out %s: %v.\n", file, err)
			code = 1
			continue
		}
		fmt.Println(path)
		added++
	}
	if added > 0 {
		return 0
	}
	return code
}

func removePicture(args []string) int {
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	file, err := filepath.Abs(args[0])
	if err == nil && !slices.Contains([]string{yourFolder(false), yourFolder(true)}, filepath.Dir(file)) {
		err = fmt.Errorf("%s isn't one of your pictures; only those in %s can be removed here", file, filepath.Join(env.AppConfig(), "backgrounds"))
	}
	shown := err == nil && current() == file
	if err == nil {
		err = fsutil.Trash(file, filepath.Join(env.DataHome(), "Trash"))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("Moved %s to the Trash\n", file)
	if !shown {
		return 0
	}
	if len(backgrounds()) == 0 {
		return drawAnother()
	}
	return moveOn()
}

func keepDrawing(args []string) int {
	drawing := current()
	if len(args) != 0 || !wallpaper.IsDrawing(drawing) {
		fmt.Fprintln(os.Stderr, "the wallpaper isn't a drawing; draw one first: solanum-background draw")
		return 1
	}
	path, err := addOne(yourFolder(false), drawing)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(path)
	return apply(path)
}
