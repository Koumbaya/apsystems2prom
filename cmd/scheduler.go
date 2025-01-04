package apsystems2prom

import (
	"context"
	"log"
	"time"
)

func start(ctx context.Context, env environment) {
	scrap := scrapper{
		username: env.username,
		systemId: env.systemId,
		ecuId:    env.ecuId,
	}

	lastDataPoint, err := scrap.scrape(ctx)
	if err != nil {
		log.Fatalf("failed to scrape initial data, exiting: %v", err)
	}
	update(lastDataPoint)

	log.Printf("scraped initial data, will trigger every %.0f minutes \n", env.tick.Minutes())

	ticker := time.NewTicker(env.tick)
	for {
		select {
		case <-ticker.C:
			currHour := time.Now().Hour()
			if env.sleepNight && currHour > env.sunDownHour || currHour < env.sunUpHour {
				continue
			}

			newDataPoint, err := scrap.scrape(ctx)
			if err != nil {
				log.Printf("error scraping data: %v\n", err)
				continue
			}

			if newDataPoint.time != lastDataPoint.time {
				update(newDataPoint)
				lastDataPoint = newDataPoint
			}

		case <-ctx.Done():
			return
		}
	}
}
