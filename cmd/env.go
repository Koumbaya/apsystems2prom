package apsystems2prom

import (
	"log"
	"os"
	"strconv"
	"time"
)

type environment struct {
	port        string
	tick        time.Duration
	sleepNight  bool
	sunUpHour   int
	sunDownHour int
	username    string
	systemId    string
	ecuId       string
}

func getEnv() environment {
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("Using default port 8080")
		port = "8080"
	}

	tick, err := time.ParseDuration(os.Getenv("TICK"))
	if tick == time.Duration(0) || err != nil {
		log.Println("Using default TICK value (15m)")
		tick = 15 * time.Minute
	}

	sleepNight, err := strconv.ParseBool(os.Getenv("SLEEP_NIGHT"))
	if err != nil {
		log.Println("Using default SLEEP_NIGHT value (false)")
	}

	sunup, err := strconv.Atoi(os.Getenv("SUNUP_HOUR"))
	if err != nil {
		log.Println("Using default SUNUP_HOUR value (6)")
		sunup = 6
	}

	sundown, err := strconv.Atoi(os.Getenv("SUNDOWN_HOUR"))
	if err != nil {
		log.Println("Using default SUNDOWN_HOUR value (21)")
		sundown = 21
	}

	username := os.Getenv("USERNAME")
	if username == "" {
		log.Fatal("USERNAME is required")
	}

	systemId := os.Getenv("SYSTEM_ID")
	if systemId == "" {
		log.Fatal("SYSTEM_ID is required")
	}

	ecuId := os.Getenv("ECU_ID")
	if ecuId == "" {
		log.Fatal("ECU_ID is required")
	}

	return environment{
		port:        port,
		tick:        tick,
		sleepNight:  sleepNight,
		sunUpHour:   sunup,
		sunDownHour: sundown,
		username:    username,
		systemId:    systemId,
		ecuId:       ecuId,
	}
}
