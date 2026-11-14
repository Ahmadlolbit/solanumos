package run

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

const holdVariable = "SOLANUM_HOLD"

var (
	holding sync.Mutex
	held    = strings.Fields(strings.ReplaceAll(os.Getenv(holdVariable), ",", " "))
	waiting []string
)

func Hold(names ...string) {
	holding.Lock()
	defer holding.Unlock()
	held = append(held, names...)
	os.Setenv(holdVariable, strings.Join(held, ","))
}

func Release() []string {
	holding.Lock()
	defer holding.Unlock()
	called := waiting
	held, waiting = nil, nil
	os.Unsetenv(holdVariable)
	return called
}

func kept(output io.Writer, name string) bool {
	holding.Lock()
	defer holding.Unlock()
	if !slices.Contains(held, filepath.Base(name)) {
		return false
	}
	if !slices.Contains(waiting, name) {
		waiting = append(waiting, name)
	}
	if output != nil {
		fmt.Fprintf(output, "%s waits for the end of the run\n", name)
	}
	return true
}
