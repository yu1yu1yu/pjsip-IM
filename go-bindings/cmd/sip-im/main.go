package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"pjsip-im/go-bindings/pjsip"
)

func run() error {
	port := flag.Int("port", 5060, "local loopback UDP port")
	username := flag.String("username", "alice", "local SIP username")
	to := flag.String("to", "", "destination SIP URI (optional)")
	text := flag.String("text", "", "UTF-8 message")
	duration := flag.Duration("duration", 0, "auto exit after this duration; 0 waits for Ctrl+C")
	version := flag.Bool("version", false, "print native PJSIP version")
	flag.Parse()
	if *version {
		fmt.Println(pjsip.Version())
		return nil
	}
	ep, err := pjsip.New(pjsip.Config{Port: *port, Username: *username})
	if err != nil {
		return err
	}
	defer ep.Close()
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(map[string]any{"kind": "ready", "port": *port, "username": *username, "pjsip": pjsip.Version()}); err != nil {
		return err
	}
	if *to != "" {
		if _, err := ep.Send(*to, *text); err != nil {
			return err
		}
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)
	var timeout <-chan time.Time
	if *duration > 0 {
		timer := time.NewTimer(*duration)
		defer timer.Stop()
		timeout = timer.C
	}
	for {
		select {
		case event, ok := <-ep.Events():
			if !ok {
				<-ep.Done()
				return ep.Err()
			}
			if err := encoder.Encode(event); err != nil {
				return err
			}
		case <-signals:
			return nil
		case <-timeout:
			return nil
		}
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
