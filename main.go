package main

import (
	"edgeproxy/cli"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	signalchan := make(chan os.Signal, 1)
	signal.Notify(signalchan, os.Interrupt, syscall.SIGTERM)
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go func() {
		select {
		case <-signalchan:
			fmt.Println()
			log.Info("Program interruption detected... closing...")
			stop()
		}
	}()

	cli.Execute(ctx)
}
