package dfi

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var terminalStatus = newStatusWriter()

type statusWriter struct {
	mu          sync.Mutex
	interactive bool
	active      bool
	width       int
}

func newStatusWriter() *statusWriter {
	return &statusWriter{
		interactive: isTerminal(os.Stdout),
		width:       0,
	}
}

func isTerminal(file *os.File) bool {
	stat, err := file.Stat()
	if err != nil {
		return false
	}
	return stat.Mode()&os.ModeCharDevice != 0
}

func (w *statusWriter) Update(message string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.interactive {
		return
	}

	message = singleLine(message)
	w.width = len(message)
	w.active = true
	fmt.Fprint(os.Stdout, "\r\033[2K"+message)
}

func (w *statusWriter) Done(message string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.interactive && w.active {
		fmt.Fprint(os.Stdout, "\r\033[2K")
		w.active = false
		w.width = 0
	}
	if message != "" {
		fmt.Fprintln(os.Stdout, message)
	}
}

func (w *statusWriter) Println(message string) {
	w.Done(message)
}

func singleLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
