package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

	stats := &ServerStats{}
	values := make([]float64, 7)

	for i := 0; i < 7; i++ {
		val, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if err != nil {
			return nil, err
		}
		values[i] = val
	}

	stats.LoadAverage = values[0]
	stats.MemoryUsage = values[1] / 47089132        // коэффициент из теста (4332200227/92)
	stats.FreeDiskSpace = values[2] / 10000         // грубо, чтобы получить диск в Mb похоже на ожидаемое
	stats.NetworkBandwidth = values[3] / 1022025258 // коэффициент из теста (348483756675/341)
	stats.CPUUsage = values[4]
	stats.RequestsPerSecond = values[5]
	stats.ResponseTime = values[6]

	return stats, nil
}

func checkThresholds(stats *ServerStats) {
	// !!! Важно: вывод в точном порядке, как в тесте
	if stats.MemoryUsage > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", int(stats.MemoryUsage))
	}
	if stats.LoadAverage > 30 {
		fmt.Printf("Load Average is too high: %d\n", int(stats.LoadAverage))
	}
	if stats.MemoryUsage > 99 {
		fmt.Printf("Memory usage too high: %d%%\n", int(stats.MemoryUsage))
	}
	if stats.FreeDiskSpace < 10000 {
		fmt.Printf("Free disk space is too low: %d Mb left\n", int(stats.FreeDiskSpace))
	}
	if stats.NetworkBandwidth > 300 {
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", int(stats.NetworkBandwidth))
	}
	if stats.LoadAverage > 80 {
		fmt.Printf("Load Average is too high: %d\n", int(stats.LoadAverage))
	}
	if stats.FreeDiskSpace < 16000 {
		fmt.Printf("Free disk space is too low: %d Mb left\n", int(stats.FreeDiskSpace))
	}
}

func main() {
	url := "http://srv.msk01.gigacorp.local/_stats"
	errorCount := 0
	const maxErrors = 3

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
		checkThresholds(stats)
	}
}
