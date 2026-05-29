package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/d-fi/GoFi/internal/dfi"
	"github.com/d-fi/GoFi/internal/web"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "web" {
		if err := runWeb(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			pauseOnWindowsError()
			os.Exit(1)
		}
		return
	}
	if err := dfi.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		pauseOnWindowsError()
		os.Exit(1)
	}
}

func runWeb(args []string) error {
	fs := flag.NewFlagSet("d-fi web", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:8080", "HTTP listen address")
	config := fs.String("config", "d-fi.config.json", "config file path")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	printWebBanner(*addr, *config)
	return web.Run(context.Background(), web.Options{
		Addr:       *addr,
		ConfigPath: *config,
	})
}

func printWebBanner(addr, config string) {
	fmt.Println("             ♥ d-fi web - " + dfi.Version + " ♥")
	fmt.Println(" ──────────────────────────────────────────────")
	fmt.Println(" │ url      http://" + addr)
	fmt.Println(" │ config   " + config)
	fmt.Println(" │ telegram https://t.me/dFiCommunity")
	fmt.Println(" ──────────────────────────────────────────────")
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
