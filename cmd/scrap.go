package apsystems2prom

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"time"
)

const (
	LOGIN_URL = "https://www.apsystemsema.com/ema/intoDemoUser.action?id="
	DATA_URL  = "https://www.apsystemsema.com/ema/ajax/getReportApiAjax/getPowerOnCurrentDayAjax"
)

type scrapper struct {
	username string
	systemId string
	ecuId    string
}

type APResponse struct {
	Time   []int64  `json:"time"`
	Power  []string `json:"power"`
	Energy []string `json:"energy"`
}

type dataPoint struct {
	time   time.Time
	power  int64
	energy float64
}

func (s *scrapper) scrape(ctx context.Context) (dataPoint, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return dataPoint{}, err
	}
	client := &http.Client{Timeout: 10 * time.Second, Jar: jar}

	// login to hydrate cookies
	if err := s.login(ctx, client); err != nil {
		return dataPoint{}, err
	}

	return s.fetchLatest(ctx, client)
}

func (s *scrapper) login(ctx context.Context, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, "GET", LOGIN_URL+s.username, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:52.0) Chrome/50.0.2661.102 Firefox/62.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}
	return nil
}

func (s *scrapper) fetchLatest(ctx context.Context, client *http.Client) (dataPoint, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", DATA_URL, nil)
	if err != nil {
		return dataPoint{}, fmt.Errorf("failed to create data request: %w", err)
	}
	q := req.URL.Query()
	q.Add("queryDate", time.Now().Format("20060102"))
	q.Add("systemId", s.systemId)
	q.Add("selectedValue", s.ecuId)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:52.0) Chrome/50.0.2661.102 Firefox/62.0")

	resp, err := client.Do(req)
	if err != nil {
		return dataPoint{}, fmt.Errorf("failed to execute data request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dataPoint{}, fmt.Errorf("data fetch failed with status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dataPoint{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var data APResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return dataPoint{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(data.Time) == 0 || len(data.Power) == 0 || len(data.Energy) == 0 {
		return dataPoint{}, fmt.Errorf("failed to fetch latest data")
	}

	tmstp := time.UnixMilli(data.Time[len(data.Time)-1:][0])
	power, err := strconv.ParseInt(data.Power[len(data.Power)-1:][0], 10, 64)
	if err != nil {
		return dataPoint{}, fmt.Errorf("failed to parse power value: %w", err)
	}

	energy, err := strconv.ParseFloat(data.Energy[len(data.Energy)-1:][0], 64)
	if err != nil {
		return dataPoint{}, fmt.Errorf("failed to parse energy: %w", err)
	}

	return dataPoint{
		time:   tmstp,
		power:  power,
		energy: energy,
	}, nil
}
