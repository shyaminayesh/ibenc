package main

import (
	"flag"
	"log"

	"ibenc/config"
	"ibenc/iperf3"
	"ibenc/metrics"
	"ibenc/remote"
)

func main() {
	// Command line flags
	configPath := flag.String("config", "ibenc.yaml", "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfigWithDefaults(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v\n", err)
	}

	log.Printf("Starting iperf3 benchmark against %s:%d\n", cfg.Iperf3.Server, cfg.Iperf3.Port)

	// Run iperf3 tests
	testResult, err := iperf3.RunBothTests(cfg.Iperf3.Server, cfg.Iperf3.Port, cfg.Iperf3.Duration)
	if err != nil {
		log.Fatalf("Test failed: %v\n", err)
	}

	log.Printf("Test Results:")
	log.Printf("  Download: %.2f Mbps\n", testResult.DownloadMbps)
	log.Printf("  Upload: %.2f Mbps\n", testResult.UploadMbps)
	log.Printf("  Latency: %.2f ms\n", testResult.LatencyMs)
	log.Printf("  Jitter: %.2f ms\n", testResult.JitterMs)
	log.Printf("  Packet Loss: %.2f %%\n", testResult.PacketLossPercent)

	// Create metrics
	metricLabels := metrics.MetricLabels{
		Location:    cfg.Metrics.Location,
		ISPName:     cfg.Metrics.ISPName,
		PackageName: cfg.Metrics.PackageName,
	}

	metricsData := metrics.ExportMetrics(testResult, metricLabels)

	// Check if test produced any meaningful results
	// Only send metrics if we got at least some valid data
	if testResult.DownloadMbps == 0 && testResult.UploadMbps == 0 {
		log.Println("❌ Error: Test results are 0 - iperf3 connection failed")
		log.Println("   This usually means:")
		log.Println("   - iperf3 server is unreachable")
		log.Println("   - Firewall is blocking port 5201")
		log.Println("   - Network connectivity issue")
		log.Println("")
		log.Println("   No metrics will be sent. Fix the connection and try again.")
		return
	}

	// Send to Grafana Cloud
	writer := remote.NewWriter(remote.Config{
		PrometheusURL: cfg.Prometheus.URL,
		Username:      cfg.Prometheus.Username,
		Password:      cfg.Prometheus.Password,
	})

	log.Printf("Sending metrics to %s\n", cfg.Prometheus.URL)
	if err := writer.WriteMetrics(metricsData); err != nil {
		log.Fatalf("Failed to send metrics: %v\n", err)
	}

	log.Println("Metrics sent successfully!")
}
