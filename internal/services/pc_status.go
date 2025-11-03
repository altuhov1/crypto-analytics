package services

import (
	"context"
	"log/slog"
	"math"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

type SystemMonitor struct {
	lastCPUPercent float64
	ctx            context.Context
}

func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{ctx: context.Background()}
}

func (sm *SystemMonitor) StartStatsReporter() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := collectStats()
			cpuUsage := sm.calculateRealCPUUsage()

			slog.Info("====Системная статистика====",
				"Количество горутин", stats.TotalGoroutines,
				"Ядра CPU", stats.NumCPU,
				"Использование CPU %", math.Round(cpuUsage*10000)/10000,
				"Память выделено МБ", stats.MemoryStats.Alloc/1024/1024,
				"Память всего выделено МБ", stats.MemoryStats.TotalAlloc/1024/1024,
				"Память системы МБ", stats.MemoryStats.Sys/1024/1024,
				"Количество сборок мусора", stats.MemoryStats.NumGC,
				"CPU на сборку мусора %", math.Round(float64(stats.MemoryStats.GCCPUFraction)*100*10000)/10000,
			)

		case <-sm.ctx.Done():
			return
		}
	}
}

func (sm *SystemMonitor) calculateRealCPUUsage() float64 {
	percentages, err := cpu.Percent(0, false) // Мгновенный снимок
	if err != nil {
		slog.Error("Ошибка получения использования CPU", "error", err)
		return sm.lastCPUPercent
	}
	if len(percentages) > 0 {
		sm.lastCPUPercent = percentages[0]
		return sm.lastCPUPercent
	}
	return sm.lastCPUPercent
}

type GoroutineStats struct {
	TotalGoroutines int
	NumCPU          int
	MemoryStats     runtime.MemStats
}

func collectStats() GoroutineStats {
	var stats GoroutineStats
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	stats.TotalGoroutines = runtime.NumGoroutine()
	stats.NumCPU = runtime.NumCPU()
	stats.MemoryStats = memStats

	return stats
}
