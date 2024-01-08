package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"gopkg.in/yaml.v3"
)

var conf struct {
	Start   string
	Refresh time.Duration
	Gym     struct {
		URL      string
		Origin   string
		Location string
		Activity string
	}
	Discord struct {
		Username string
		Webhook  string
	}
}

const dateFmt = "2006-01-02"

var currentDate time.Time

func theThing() {
	date := currentDate.Format(dateFmt)
	niceDate := currentDate.Format("Monday 02 January")

	log.Printf("Checking %s", niceDate)

	sessions, err := getSessions(date)
	if err != nil {
		if errors.Is(err, errRedirect) {
			log.Print("Not available yet")
		} else {
			log.Printf("Error: %v", err)
		}

		return
	}

	var msg string
	msg += fmt.Sprintf("Sessions available for %s", niceDate)
	for _, session := range sessions {
		msg += "\n"
		msg += fmt.Sprintf("\t%d spaces at %s", session.Spaces, session.StartsAt.Hour)
	}

	log.Print(msg)
	if len(sessions) == 0 {
		return
	}

	// Send notification.

	message := Message{
		Username: &conf.Discord.Username,
		Content:  &msg,
	}

	err = sendMessage(message, conf.Discord.Webhook)
	if err != nil {
		log.Print(err)
	}

	// Advance to next day.
	currentDate = currentDate.Add(time.Hour * 24)
}

func main() {
	// Load config.
	yamlFile, err := os.ReadFile("conf.yaml")
	if err != nil {
		log.Fatalf("load condfig: %v", err)
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Fatalf("yaml.Unmarshal: %v", err)
	}

	currentDate, err = time.Parse(dateFmt, conf.Start)
	if err != nil {
		log.Fatalf("time.Parse: %v", err)
	}

	// Create context which is cancelled by CTRL+C.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case <-c:
		case <-ctx.Done():
		}
		cancel()
		signal.Stop(c)
	}()

	// Main program loop.
	for {
		theThing()

		select {
		case <-time.After(conf.Refresh):
		case <-ctx.Done():
			log.Printf("Done")
			os.Exit(0)
		}
	}

}
