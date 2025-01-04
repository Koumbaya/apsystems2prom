package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

const (
	LOGIN_URL     = "https://www.apsystemsema.com/ema/intoDemoUser.action?id="
	DATA_URL      = "https://www.apsystemsema.com/ema/ajax/getReportApiAjax/getPowerOnCurrentDayAjax"
	DASHBOARD_URL = "https://www.apsystemsema.com/ema/ajax/getDashboardApiAjax/getDashboardProductionInfoAjax"
	PANELS_URL    = "https://www.apsystemsema.com/ema/ajax/getViewAjax/getViewPowerByViewAjax"
)

var ErrNoData = errors.New("no data yet") //silent error for beginning of day where ap systems respond but there are no data points.

type scrapper struct {
	username string
	systemId string
	ecuId    string
	vid      string
}

// Response from DATA_URL
type APResponse struct {
	Time   []int64  `json:"time"`
	Power  []string `json:"power"`
	Energy []string `json:"energy"`
	Total  string   `json:"total"`
	Max    string   `json:"max"`
}

// Response from DASHBOARD_URL
type DashResponse struct {
	Lifetime string `json:"lifetime"`
}

type panelData struct {
	id    string
	power float64
}

type dataPoint struct {
	time     time.Time
	total    float64
	power    int64
	energy   float64
	max      float64
	lifetime float64
	panels   []panelData
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
	q.Add("queryDate", time.Now().UTC().Format("20060102"))
	q.Add("systemId", s.systemId)
	q.Add("selectedValue", s.ecuId)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:52.0) Chrome/50.0.2661.102 Firefox/62.0")

	resp, err := client.Do(req)
	if err != nil {
		return dataPoint{}, fmt.Errorf("failed to execute data request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent {
			return dataPoint{}, ErrNoData
		}
		return dataPoint{}, fmt.Errorf("data fetch failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return dataPoint{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var data APResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return dataPoint{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(data.Time) == 0 || len(data.Power) == 0 || len(data.Energy) == 0 {
		return dataPoint{}, ErrNoData
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

	total, err := strconv.ParseFloat(data.Total, 64)
	if err != nil {
		return dataPoint{}, fmt.Errorf("failed to parse total: %w", err)
	}

	maxi, err := strconv.ParseFloat(data.Max, 64)
	if err != nil {
		return dataPoint{}, fmt.Errorf("failed to parse total: %w", err)
	}

	lifetime, err := fetchLifetime(ctx, client)
	if err != nil {
		return dataPoint{}, err
	}

	var panels []panelData
	if s.vid != "" {
		panels, err = s.fetchPanels(ctx, client)
		if err != nil {
			return dataPoint{}, err
		}
	}

	return dataPoint{
		time:     tmstp,
		power:    power,
		energy:   energy,
		total:    total,
		max:      maxi,
		lifetime: lifetime,
		panels:   panels,
	}, nil
}

func fetchLifetime(ctx context.Context, client *http.Client) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", DASHBOARD_URL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create data request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:52.0) Chrome/50.0.2661.102 Firefox/62.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute data request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("data fetch failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var data DashResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	lifetime, err := strconv.ParseFloat(data.Lifetime, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse lifetime: %w", err)
	}

	return lifetime, nil
}

func (s *scrapper) fetchPanels(ctx context.Context, client *http.Client) ([]panelData, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", PANELS_URL, nil)
	if err != nil {
		return []panelData{}, fmt.Errorf("failed to create data request: %w", err)
	}

	q := req.URL.Query()
	q.Add("sid", s.systemId)
	q.Add("vid", s.vid)
	q.Add("iid", "")
	q.Add("date", time.Now().UTC().Format("20060102"))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:52.0) Chrome/50.0.2661.102 Firefox/62.0")

	resp, err := client.Do(req)
	if err != nil {
		return []panelData{}, fmt.Errorf("failed to execute data request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []panelData{}, fmt.Errorf("data fetch failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []panelData{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var data struct {
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return []panelData{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if data.Detail == "" {
		return []panelData{}, nil
	}

	pnls := strings.Split(data.Detail, "&")
	paneldata := make([]panelData, 0)
	for _, pnl := range pnls {
		components := strings.Split(pnl, "/")
		if len(components) != 2 {
			continue
		}
		vals := strings.Split(components[1], ",")
		last := vals[len(vals)-1:][0]
		lastf, err := strconv.ParseFloat(last, 64)
		if err != nil {
			continue
		}
		paneldata = append(paneldata, panelData{
			id:    components[0],
			power: lastf,
		})
	}

	return paneldata, nil
}
