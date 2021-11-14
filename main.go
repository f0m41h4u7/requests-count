package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	// Interval to count stats for, in seconds
	statsInterval = 60
	// Granularity for saving stats, in seconds
	statsPrecision = 1
)

func main() {
	// Quantity of time intervals
	statsQuants := int(statsInterval / statsPrecision)
	serv := NewServer("127.0.0.1", "1337", "stats.txt", statsQuants)

	done := make(chan os.Signal, 1)
	errs := make(chan error, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		if err := serv.Stop(); err != nil {
			log.Fatal("server stopped with error: %w", err)
			return
		}
	}()
	defer signal.Stop(done)

	go func() {
		log.Println("server started")
		errs <- serv.Run()
	}()

	ticker := time.NewTicker(time.Duration(statsPrecision) * time.Second)
	stop := false
	for !stop {
		select {
		case <-ticker.C:
			serv.UpdateStats()
		case <-done:
			stop = true
		case err := <-errs:
			if err != nil {
				log.Fatal("server exited with error: %w", err)
			}
		}
	}
}
