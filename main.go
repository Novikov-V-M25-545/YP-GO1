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
	DiskUsagePercent  float64
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

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 status code")
	}

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

	for i := 0; i < 7; i++ {
		v, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}

	stats := &ServerStats{}
	stats.LoadAverage = values[0]

	// values[1] = всего памяти, values[2] = использованной памяти
	if values[1] > 0 {
		stats.MemoryUsage = (values[2] / values[1]) * 100
	}

	// values[3] = всего диска, values[4] = использованного диска
	freeDiskBytes := values[3] - values[4]
	stats.FreeDiskSpace = freeDiskBytes / 1024 / 1024

	// Процент использования диска
	if values[3] > 0 {
		stats.DiskUsagePercent = (values[4] / values[3]) * 100
	}

	// values[5] = пропускная способность, values[6] = загруженность
	if values[5] > 0 {
		bandwidthUsagePercent := (values[6] / values[5]) * 100

		if bandwidthUsagePercent > 90 {
			freeBandwidthBytes := values[5] - values[6]
			stats.NetworkBandwidth = freeBandwidthBytes / 1_000_000
		}
	}

	stats.CPUUsage = values[1]
	stats.RequestsPerSecond = values[5]
	stats.ResponseTime = values[6]

	return stats, nil
}

func checkThresholds(stats *ServerStats) {
	if stats.LoadAverage > 30 {
		fmt.Printf("Load Average is too high: %d\n", int(stats.LoadAverage))
	}

	if stats.MemoryUsage > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", int(stats.MemoryUsage))
	}

	// Проверяем если использование диска > 90%
	if stats.DiskUsagePercent > 90 {
		fmt.Printf("Free disk space is too low: %d Mb left\n", int(stats.FreeDiskSpace))
	}

	if stats.NetworkBandwidth > 0 {
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", int(stats.NetworkBandwidth))
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

	// Небольшая задержка перед выходом для гарантии обработки последних запросов
	time.Sleep(100 * time.Millisecond)
}
