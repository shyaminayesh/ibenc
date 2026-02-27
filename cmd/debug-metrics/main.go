package main

import (
	"flag"
	"fmt"
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

	log.Println("Creating debug metrics...")

	// Create single metric for debugging
	metricName := "ibenc_download_speed_mbps"
	help := "Download speed in Mbps"
	value := 85.5
	now := time.Now().UnixMilli()

	metric := &io_prometheus_client.MetricFamily{
		Name: &metricName,
		Help: &help,
		Type: io_prometheus_client.MetricType_GAUGE.Enum(),
		Metric: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{
					{Name: stringPtr("location"), Value: &cfg.Metrics.Location},
					{Name: stringPtr("isp_name"), Value: &cfg.Metrics.ISPName},
					{Name: stringPtr("package_name"), Value: &cfg.Metrics.PackageName},
				},
				Gauge:       &io_prometheus_client.Gauge{Value: &value},
				TimestampMs: &now,
			},
		},
	}

	fmt.Printf("Metric Name: %s\n", *metric.Name)
	fmt.Printf("Metric Help: %s\n", *metric.Help)
	fmt.Printf("Metric Type: %v\n", *metric.Type)
	fmt.Printf("Metric Value: %v\n", *metric.Metric[0].Gauge.Value)
	fmt.Printf("Metric Timestamp: %d\n", *metric.Metric[0].TimestampMs)
	fmt.Printf("Labels:\n")
	for _, label := range metric.Metric[0].Label {
		fmt.Printf("  %s = %s\n", *label.Name, *label.Value)
	}

	metrics := []*io_prometheus_client.MetricFamily{metric}

	// Send to Grafana Cloud
	writer := remote.NewWriter(remote.Config{
		PrometheusURL: cfg.Prometheus.URL,
		Username:      cfg.Prometheus.Username,
		Password:      cfg.Prometheus.Password,
	})

	log.Printf("Sending metrics to %s\n", cfg.Prometheus.URL)
	if err := writer.WriteMetrics(metrics); err != nil {
		log.Fatalf("Failed to send metrics: %v\n", err)
	}

	log.Println("Metrics sent successfully!")
}

func stringPtr(s string) *string {
	return &s
}
