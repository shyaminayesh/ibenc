package metrics

import (
	"time"

	"github.com/prometheus/client_model/go"
	"ibenc/iperf3"
)

// MetricLabels holds the labels for metrics
type MetricLabels struct {
	Location    string
	ISPName     string
	PackageName string
}

// ExportMetrics converts test results to Prometheus metrics
func ExportMetrics(result *iperf3.TestResult, labels MetricLabels) []*io_prometheus_client.MetricFamily {
	timestamp := time.Now().UnixMilli()

	metrics := make([]*io_prometheus_client.MetricFamily, 0)

	// Download speed metric
	metrics = append(metrics, createGaugeMetric(
		"ibenc_download_speed_mbps",
		"Download speed in Mbps",
		result.DownloadMbps,
		labels,
		timestamp,
	))

	// Upload speed metric
	metrics = append(metrics, createGaugeMetric(
		"ibenc_upload_speed_mbps",
		"Upload speed in Mbps",
		result.UploadMbps,
		labels,
		timestamp,
	))

	// Jitter metric
	metrics = append(metrics, createGaugeMetric(
		"ibenc_jitter_ms",
		"Jitter in milliseconds",
		result.JitterMs,
		labels,
		timestamp,
	))

	// Latency metric
	metrics = append(metrics, createGaugeMetric(
		"ibenc_latency_ms",
		"Latency in milliseconds",
		result.LatencyMs,
		labels,
		timestamp,
	))

	// Packet loss metric
	metrics = append(metrics, createGaugeMetric(
		"ibenc_packet_loss_percent",
		"Packet loss percentage",
		result.PacketLossPercent,
		labels,
		timestamp,
	))

	return metrics
}

// createGaugeMetric creates a Prometheus gauge metric
func createGaugeMetric(name, help string, value float64, labels MetricLabels, timestamp int64) *io_prometheus_client.MetricFamily {
	mf := &io_prometheus_client.MetricFamily{
		Name: &name,
		Help: &help,
		Type: io_prometheus_client.MetricType_GAUGE.Enum(),
		Metric: []*io_prometheus_client.Metric{
			{
				Label: []*io_prometheus_client.LabelPair{
					{
						Name:  stringPtr("location"),
						Value: &labels.Location,
					},
					{
						Name:  stringPtr("isp_name"),
						Value: &labels.ISPName,
					},
					{
						Name:  stringPtr("package_name"),
						Value: &labels.PackageName,
					},
				},
				Gauge: &io_prometheus_client.Gauge{
					Value: &value,
				},
				TimestampMs: &timestamp,
			},
		},
	}
	return mf
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
