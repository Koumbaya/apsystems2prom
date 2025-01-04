package main

import (
	"context"
	"errors"
	"log"
	"time"
)

func startScheduler(ctx context.Context, env environment) {
	scrap := scrapper{
		username: env.username,
		systemId: env.systemId,
		ecuId:    env.ecuId,
		vid:      env.vid,
	}

	lastDataPoint, err := scrap.scrape(ctx)
	if err != nil {
		log.Fatalf("failed to scrape initial data, exiting: %v", err)
	}
	updateMetrics(lastDataPoint)

	log.Printf("scraped initial data at %s, will trigger every %.0f minutes \n", time.Now().UTC().Format("15:04:05"), env.tick.Minutes())

	night := false
	ticker := time.NewTicker(env.tick)
	for {
		select {
		case <-ticker.C:
			currHour := time.Now().UTC().Hour()
			if env.sleepNight && currHour > env.sunDownHour || currHour < env.sunUpHour {
				if night == false {
					night = true
					log.Printf("entering night mode at %s\n", time.Now().UTC().Format("15:04:05"))
				}
				continue
			}

			if night == true {
				night = false
				log.Printf("leaving night mode at %s\n", time.Now().UTC().Format("15:04:05"))
			}

			newDataPoint, err := scrap.scrape(ctx)
			if err != nil {
				if !errors.Is(err, ErrNoData) {
					log.Printf("error scraping data: %v\n", err)
				}
				continue
			}

			if newDataPoint.time != lastDataPoint.time {
				updateMetrics(newDataPoint)
				lastDataPoint = newDataPoint
			}

		case <-ctx.Done():
			return
		}
	}
}
