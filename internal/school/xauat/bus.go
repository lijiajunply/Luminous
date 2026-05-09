package xauat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"luminous/internal/util"
)

var busCache = util.NewCacheWithName("bus")

type busRecord struct {
	LineName           string `json:"lineName"`
	Description        string `json:"description"`
	DepartureStation   string `json:"departureStation"`
	ArrivalStation     string `json:"arrivalStation"`
	RunTime            string `json:"runTime"`
	ArrivalStationTime string `json:"arrivalStationTime"`
}

func mapBusItems(records []busRecord) []BusItem {
	items := make([]BusItem, 0, len(records))
	for _, r := range records {
		items = append(items, BusItem{
			LineName:           r.LineName,
			Description:        r.Description,
			DepartureStation:   r.DepartureStation,
			ArrivalStation:     r.ArrivalStation,
			RunTime:            r.RunTime,
			ArrivalStationTime: r.ArrivalStationTime,
		})
	}
	return items
}

// GetBus 获取校车时刻表（先尝试旧接口，再尝试新接口）
func GetBus(date, loc string) (*BusResponse, error) {
	cacheKey := fmt.Sprintf("bus:%s:%s", date, loc)
	val, err := busCache.GetOrSet(cacheKey, 12*time.Hour, func() (interface{}, error) {
		bus, err := getOldBus(date)
		if err != nil {
			bus, err = getNewBus(date, loc)
			if err != nil {
				return nil, err
			}
		}
		return bus, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(*BusResponse), nil
}

func getOldBus(date string) (*BusResponse, error) {
	url := fmt.Sprintf("%s/api/school/bus/user/runPlanPage?current=1&size=30&keyWord=&lineId=&date=%s", oldBusURL, date)
	resp, err := util.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Records []busRecord `json:"records"`
		Total   int         `json:"total"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse bus json: %w", err)
	}

	return &BusResponse{Records: mapBusItems(result.Records), Total: result.Total}, nil
}

func getNewBus(date, loc string) (*BusResponse, error) {
	reqBody, err := json.Marshal(map[string]string{
		"type":   loc,
		"nowDay": date,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal bus request: %w", err)
	}

	req, err := http.NewRequest("POST", newBusURL+"/api/openapi/getDayBusPlans", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", util.RandomUserAgent())
	req.Body = io.NopCloser(strings.NewReader(string(reqBody)))

	resp, err := util.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Records []busRecord `json:"records"`
		Total   int         `json:"total"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse new bus json: %w", err)
	}

	return &BusResponse{Records: mapBusItems(result.Records), Total: result.Total}, nil
}
