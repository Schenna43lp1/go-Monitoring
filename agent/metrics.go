package main

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type SystemMetrics struct {
	CPUPercent  float64
	RAMPercent  float64
	DiskPercent float64
	RAMTotalMB  float64
	RAMUsedMB   float64
	DiskTotalGB float64
	DiskUsedGB  float64
}

func collectMetrics() SystemMetrics {
	if runtime.GOOS == "windows" {
		return collectWindowsMetrics()
	}
	return collectUnixMetrics()
}

func collectWindowsMetrics() SystemMetrics {
	var result struct {
		CPUPercent  float64 `json:"cpu_percent"`
		RAMPercent  float64 `json:"ram_percent"`
		DiskPercent float64 `json:"disk_percent"`
		RAMTotalMB  float64 `json:"ram_total_mb"`
		RAMUsedMB   float64 `json:"ram_used_mb"`
		DiskTotalGB float64 `json:"disk_total_gb"`
		DiskUsedGB  float64 `json:"disk_used_gb"`
	}

	drive := os.Getenv("SystemDrive")
	if drive == "" {
		drive = "C:"
	}

	script := `
$os = Get-CimInstance Win32_OperatingSystem
$cpu = (Get-CimInstance Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average
$disk = Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='` + drive + `'"
$ramTotalMb = [math]::Round($os.TotalVisibleMemorySize / 1024, 2)
$ramFreeMb = [math]::Round($os.FreePhysicalMemory / 1024, 2)
$diskTotalGb = [math]::Round($disk.Size / 1GB, 2)
$diskFreeGb = [math]::Round($disk.FreeSpace / 1GB, 2)
[pscustomobject]@{
  cpu_percent = [math]::Round($cpu, 2)
  ram_percent = [math]::Round((($ramTotalMb - $ramFreeMb) / $ramTotalMb) * 100, 2)
  disk_percent = [math]::Round((($diskTotalGb - $diskFreeGb) / $diskTotalGb) * 100, 2)
  ram_total_mb = $ramTotalMb
  ram_used_mb = [math]::Round($ramTotalMb - $ramFreeMb, 2)
  disk_total_gb = $diskTotalGb
  disk_used_gb = [math]::Round($diskTotalGb - $diskFreeGb, 2)
} | ConvertTo-Json -Compress
`

	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		return SystemMetrics{}
	}
	_ = json.Unmarshal(out, &result)
	return SystemMetrics{
		CPUPercent:  result.CPUPercent,
		RAMPercent:  result.RAMPercent,
		DiskPercent: result.DiskPercent,
		RAMTotalMB:  result.RAMTotalMB,
		RAMUsedMB:   result.RAMUsedMB,
		DiskTotalGB: result.DiskTotalGB,
		DiskUsedGB:  result.DiskUsedGB,
	}
}

func collectUnixMetrics() SystemMetrics {
	metrics := SystemMetrics{}

	total, available := readMemInfo()
	if total > 0 {
		used := total - available
		metrics.RAMTotalMB = round(float64(total)/1024, 2)
		metrics.RAMUsedMB = round(float64(used)/1024, 2)
		metrics.RAMPercent = round(float64(used)/float64(total)*100, 2)
	}

	metrics.CPUPercent = round(readCPUPercent(), 2)

	if totalKB, usedKB := readDiskUsage(); totalKB > 0 {
		metrics.DiskTotalGB = round(float64(totalKB)/1024/1024, 2)
		metrics.DiskUsedGB = round(float64(usedKB)/1024/1024, 2)
		metrics.DiskPercent = round(float64(usedKB)/float64(totalKB)*100, 2)
	}

	return metrics
}

func readMemInfo() (totalKB int64, availableKB int64) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, _ := strconv.ParseInt(fields[1], 10, 64)
		switch strings.TrimSuffix(fields[0], ":") {
		case "MemTotal":
			totalKB = value
		case "MemAvailable":
			availableKB = value
		}
	}
	return totalKB, availableKB
}

func readCPUPercent() float64 {
	idle1, total1 := readCPUStat()
	time.Sleep(100 * time.Millisecond)
	idle2, total2 := readCPUStat()

	totalDelta := total2 - total1
	if totalDelta <= 0 {
		return 0
	}
	return (1 - float64(idle2-idle1)/float64(totalDelta)) * 100
}

func readCPUStat() (idle int64, total int64) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, 0
	}
	line := strings.SplitN(string(data), "\n", 2)[0]
	fields := strings.Fields(line)
	for i, field := range fields[1:] {
		value, _ := strconv.ParseInt(field, 10, 64)
		total += value
		if i == 3 || i == 4 {
			idle += value
		}
	}
	return idle, total
}

func readDiskUsage() (totalKB int64, usedKB int64) {
	out, err := exec.Command("df", "-k", "/").Output()
	if err != nil {
		return 0, 0
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0, 0
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 3 {
		return 0, 0
	}
	totalKB, _ = strconv.ParseInt(fields[1], 10, 64)
	usedKB, _ = strconv.ParseInt(fields[2], 10, 64)
	return totalKB, usedKB
}

func round(value float64, places int) float64 {
	multiplier := math.Pow(10, float64(places))
	return math.Round(value*multiplier) / multiplier
}
