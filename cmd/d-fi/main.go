package main

import (
	"context"
	"fmt"
	"os"

	"github.com/d-fi/GoFi/internal/dfi"
)

func main() {
	if err := dfi.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
