package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/d-fi/GoFi/internal/dfi"
)

func main() {
	if err := dfi.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		pauseOnWindowsError()
		os.Exit(1)
	}
}

func pauseOnWindowsError() {
	if !shouldPauseOnError(runtime.GOOS, os.Getenv) {
		return
	}
	fmt.Fprint(os.Stderr, "Press Enter to exit...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func shouldPauseOnError(goos string, getenv func(string) string) bool {
	if goos != "windows" {
		return false
	}
	value := strings.TrimSpace(strings.ToLower(getenv("DFI_NO_PAUSE")))
	return value != "1" && value != "true" && value != "yes"
}
