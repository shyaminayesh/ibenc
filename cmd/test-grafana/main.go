package main

import (
	"flag"
	"log"
	"time"

	"github.com/prometheus/client_model/go"
	"ibenc/config"
	"ibenc/remote"
)

func main() {
	configPath := flag.String("config", "ibenc.yaml", "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfigWithDefaults(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v\n", err)
	}

	log.Println("Creating mock metrics for testing...")

	// Create mock metrics
	metrics := createMockMetrics(cfg)

	// Send to Grafana Cloud
	writer := remote.NewWriter(remote.Config{
		PrometheusURL: cfg.Prometheus.URL,
		Username:      cfg.Prometheus.Username,
		Password:      cfg.Prometheus.Password,
	})

	log.Printf("Sending mock metrics to %s\n", cfg.Prometheus.URL)
	if err := writer.WriteMetrics(metrics); err != nil {
		log.Fatalf("Failed to send metrics: %v\n", err)
	}

	log.Println("Mock metrics sent successfully!")
}

func createMockMetrics(cfg *config.Config) []*io_prometheus_client.MetricFamily {
	metrics := make([]*io_prometheus_client.MetricFamily, 0)
	now := time.Now().UnixMilli()

	// Mock download speed
	downloadName := "ibenc_download_speed_mbps"
	downloadHelp := "Download speed in Mbps"
	downloadValue := 85.5
	downloadTimestamp := now

	metrics = append(metrics, &io_prometheus_client.MetricFamily{
		Name: &downloadName,
		Help: &downloadHelp,
		Type: io_prometheus_client.MetricType_GAUGE.Enum(),
		Metric: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{
					{Name: stringPtr("location"), Value: &cfg.Metrics.Location},
					{Name: stringPtr("isp_name"), Value: &cfg.Metrics.ISPName},
					{Name: stringPtr("package_name"), Value: &cfg.Metrics.PackageName},
				},
				Gauge:       &io_prometheus_client.Gauge{Value: &downloadValue},
				TimestampMs: &downloadTimestamp,
			},
		},
	})

	// Mock upload speed
	uploadName := "ibenc_upload_speed_mbps"
	uploadHelp := "Upload speed in Mbps"
	uploadValue := 42.3
	uploadTimestamp := now

	metrics = append(metrics, &io_prometheus_client.MetricFamily{
		Name: &uploadName,
		Help: &uploadHelp,
		Type: io_prometheus_client.MetricType_GAUGE.Enum(),
		Metric: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{
					{Name: stringPtr("location"), Value: &cfg.Metrics.Location},
					{Name: stringPtr("isp_name"), Value: &cfg.Metrics.ISPName},
					{Name: stringPtr("package_name"), Value: &cfg.Metrics.PackageName},
				},
				Gauge:       &io_prometheus_client.Gauge{Value: &uploadValue},
				TimestampMs: &uploadTimestamp,
			},
		},
	})

	// Mock latency
	latencyName := "ibenc_latency_ms"
	latencyHelp := "Latency in milliseconds"
	latencyValue := 45.23
	latencyTimestamp := now

	metrics = append(metrics, &io_prometheus_client.MetricFamily{
		Name: &latencyName,
		Help: &latencyHelp,
		Type: io_prometheus_client.MetricType_GAUGE.Enum(),
		Metric: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{
					{Name: stringPtr("location"), Value: &cfg.Metrics.Location},
					{Name: stringPtr("isp_name"), Value: &cfg.Metrics.ISPName},
					{Name: stringPtr("package_name"), Value: &cfg.Metrics.PackageName},
				},
				Gauge:       &io_prometheus_client.Gauge{Value: &latencyValue},
				TimestampMs: &latencyTimestamp,
			},
		},
	})

	return metrics
}

func stringPtr(s string) *string {
	return &s
}
