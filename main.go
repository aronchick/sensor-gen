package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// SensorReading represents a single IoT/OT sensor data point from pipeline infrastructure
type SensorReading struct {
	SensorID    string  `json:"sensor_id"`
	Timestamp   string  `json:"timestamp"`
	Type        string  `json:"type"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Location    Location `json:"location"`
	PipelineID  string  `json:"pipeline_id"`
	Status      string  `json:"status"`
	Quality     float64 `json:"quality_score"`
	AlertLevel  string  `json:"alert_level,omitempty"`
}

type Location struct {
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	MilePost float64 `json:"mile_post"`
}

var sensorTypes = []struct {
	Type string
	Unit string
	Min  float64
	Max  float64
}{
	{"pressure", "psi", 200, 1500},
	{"temperature", "fahrenheit", -20, 180},
	{"flow_rate", "bbl/hr", 0, 50000},
	{"vibration", "mm/s", 0, 25},
	{"corrosion", "mpy", 0, 50},
	{"humidity", "percent", 0, 100},
	{"gas_detector", "ppm", 0, 1000},
	{"valve_position", "percent", 0, 100},
}

var pipelineIDs = []string{
	"PIPE-TX-001", "PIPE-TX-002", "PIPE-OK-001", "PIPE-LA-001",
	"PIPE-NM-001", "PIPE-CO-001", "PIPE-WY-001", "PIPE-ND-001",
}

var statuses = []string{"normal", "normal", "normal", "normal", "warning", "maintenance"}
var alertLevels = []string{"", "", "", "", "", "low", "medium", "high"}

func main() {
	outputFile := flag.String("o", "output.jsonl", "Output file path")
	rate := flag.Int("rate", 10000, "Target entries per second")
	duration := flag.Duration("d", 0, "Duration to run (0 = indefinite)")
	verbose := flag.Bool("v", false, "Verbose output with stats")
	appendMode := flag.Bool("append", false, "Append to existing file instead of overwriting")
	flag.Parse()

	// Open file in truncate (default) or append mode
	var file *os.File
	var err error
	if *appendMode {
		file, err = os.OpenFile(*outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
	} else {
		file, err = os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			os.Exit(1)
		}
	}
	defer file.Close()

	// Buffered writer for performance
	writer := bufio.NewWriterSize(file, 1024*1024) // 1MB buffer
	defer writer.Flush()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	mode := "overwriting"
	if *appendMode {
		mode = "appending"
	}
	fmt.Printf("Generating sensor data to %s (%s) at ~%d entries/sec\n", *outputFile, mode, *rate)
	if *duration > 0 {
		fmt.Printf("Duration: %v\n", *duration)
	}
	fmt.Println("Press Ctrl+C to stop...")

	// Batch for better throughput
	batchSize := 1000
	if *rate < batchSize {
		batchSize = *rate
	}
	if batchSize < 1 {
		batchSize = 1
	}
	// Use float64 to avoid integer division truncation
	interval := time.Duration(float64(time.Second) * float64(batchSize) / float64(*rate))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var endTime time.Time
	if *duration > 0 {
		endTime = time.Now().Add(*duration)
	}

	totalEntries := int64(0)
	startTime := time.Now()
	lastReport := startTime

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		select {
		case <-sigChan:
			writer.Flush()
			printFinalStats(totalEntries, startTime, *outputFile)
			return
		case <-ticker.C:
			if *duration > 0 && time.Now().After(endTime) {
				writer.Flush()
				printFinalStats(totalEntries, startTime, *outputFile)
				return
			}

			// Write batch
			for range batchSize {
				reading := generateReading(rng)
				data, _ := json.Marshal(reading)
				writer.Write(data)
				writer.WriteByte('\n')
				totalEntries++
			}

			// Flush after each batch for real-time observability (tail -f)
			writer.Flush()

			// Periodic stats
			if *verbose && time.Since(lastReport) >= 5*time.Second {
				elapsed := time.Since(startTime).Seconds()
				rate := float64(totalEntries) / elapsed
				fmt.Printf("  %d entries written (%.0f/sec avg)\n", totalEntries, rate)
				lastReport = time.Now()
			}
		}
	}
}

func generateReading(rng *rand.Rand) SensorReading {
	st := sensorTypes[rng.Intn(len(sensorTypes))]
	pipeline := pipelineIDs[rng.Intn(len(pipelineIDs))]
	status := statuses[rng.Intn(len(statuses))]
	alert := alertLevels[rng.Intn(len(alertLevels))]

	// Generate value with occasional anomalies
	value := st.Min + rng.Float64()*(st.Max-st.Min)
	if rng.Float64() < 0.02 { // 2% chance of anomaly
		value = st.Max + rng.Float64()*st.Max*0.2 // Exceed max by up to 20%
		if alert == "" {
			alert = "medium"
		}
	}

	return SensorReading{
		SensorID:   fmt.Sprintf("SNS-%s-%04d", st.Type[:3], rng.Intn(10000)),
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Type:       st.Type,
		Value:      value,
		Unit:       st.Unit,
		PipelineID: pipeline,
		Status:     status,
		Quality:    0.85 + rng.Float64()*0.15,
		AlertLevel: alert,
		Location: Location{
			Lat:      25.0 + rng.Float64()*20, // Roughly US oil/gas regions
			Lon:      -105.0 + rng.Float64()*15,
			MilePost: rng.Float64() * 500,
		},
	}
}

func printFinalStats(total int64, start time.Time, filename string) {
	elapsed := time.Since(start)
	rate := float64(total) / elapsed.Seconds()

	fi, _ := os.Stat(filename)
	sizeMB := float64(fi.Size()) / (1024 * 1024)

	fmt.Printf("\n--- Final Stats ---\n")
	fmt.Printf("Total entries: %d\n", total)
	fmt.Printf("Duration: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("Average rate: %.0f entries/sec\n", rate)
	fmt.Printf("File size: %.2f MB\n", sizeMB)
	fmt.Printf("Avg entry size: %.0f bytes\n", float64(fi.Size())/float64(total))
}
