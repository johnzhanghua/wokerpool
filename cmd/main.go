package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	e := make(chan os.Signal)
	signal.Notify(e, os.Interrupt, syscall.SIGTERM)

	_, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	<-e
	os.Exit(1)
}
