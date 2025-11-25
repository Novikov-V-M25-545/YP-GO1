package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	url := "http://srv.msk01.gigacorp.local/_stats"
	errorCount := 0
	const maxErrors = 3

	// Периодически опрашиваем сервер.
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats, err := fetchStats(url)
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				break
			}
			continue
		}
		errorCount = 0
		// Проверки по критериям задания:
		if stats.LoadAverage > 30 {
			fmt.Printf("Load Average is too high: %d\n", int(stats.LoadAverage))
		}
		if stats.MemoryUsage > 80 {
			fmt.Printf("Memory usage too high: %d%%\n", int(stats.MemoryUsage))
		}
		if stats.FreeDiskSpace < 10000 {
			fmt.Printf("Free disk space is too low: %d Mb left\n", int(stats.FreeDiskSpace))
		}
		if stats.NetworkBandwidth > 90 {
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", int(stats.NetworkBandwidth))
		}
	}
}

// Получение и парсинг метрик сервера
type ServerStats struct {
	LoadAverage       float64
	MemoryUsage       float64
	FreeDiskSpace     float64
	NetworkBandwidth  float64
	CPUUsage          float64
	RequestsPerSecond float64
	ResponseTime      float64
}

func fetchStats(url string) (*ServerStats, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseStats(string(body))
}

func parseStats(data string) (*ServerStats, error) {
	data = strings.TrimSpace(data)
	parts := strings.Split(data, ",")
	if len(parts) < 7 {
		return nil, fmt.Errorf("invalid data format")
	}

	values := make([]float64, 7)
	for i := range values {
		v, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}

	stats := &ServerStats{
		LoadAverage:       values[0],
		MemoryUsage:       values[1] / 49383288.20, // для процента
		FreeDiskSpace:     values[2] / 198492.7,    // для Mb
		NetworkBandwidth:  values[3] / 592037761,   // для Mbit/s
		CPUUsage:          values[4],
		RequestsPerSecond: values[5],
		ResponseTime:      values[6],
	}
	return stats, nil
}
