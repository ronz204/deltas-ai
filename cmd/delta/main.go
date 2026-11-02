package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

var version = "dev"

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "delta:", err)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, wd)
	stop()
	os.Exit(code)
}
