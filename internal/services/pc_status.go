package services

import (
	"context"
	"log/slog"
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

			slog.Info("System statistics",
				"goroutines", stats.TotalGoroutines,
				"cpu_cores", stats.NumCPU,
				"cpu_usage_percent", cpuUsage,
				"alloc_memory_mb", stats.MemoryStats.Alloc/1024/1024,
				"total_alloc_memory_mb", stats.MemoryStats.TotalAlloc/1024/1024,
				"sys_memory_mb", stats.MemoryStats.Sys/1024/1024,
				"num_gc", stats.MemoryStats.NumGC,
				"gc_cpu_percent", float64(stats.MemoryStats.GCCPUFraction)*100,
			)

		case <-sm.ctx.Done():
			return
		}
	}
}

func (sm *SystemMonitor) calculateRealCPUUsage() float64 {
	percentages, err := cpu.Percent(0, false) // Мгновенный снимок
	if err != nil {
		slog.Error("Failed to get CPU usage", "error", err)
		return sm.lastCPUPercent // Возвращаем последнее известное значение
	}
	if len(percentages) > 0 {
		sm.lastCPUPercent = percentages[0]
		return sm.lastCPUPercent
	}
	return sm.lastCPUPercent
}

// Остальные функции без изменений
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
