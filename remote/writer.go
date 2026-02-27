package remote

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/client_model/go"
	"github.com/prometheus/prometheus/prompb"
)

// Config holds Grafana Cloud authentication and endpoint details
type Config struct {
	PrometheusURL string // e.g., "https://prometheus-prod-01-eu-west-0.grafana.net/api/prom"
	Username      string // Prometheus instance ID or "prometheus"
	Password      string // API token
}

// Writer sends metrics to Grafana Cloud via remote write API
type Writer struct {
	config Config
	client *http.Client
}

// NewWriter creates a new Grafana Cloud metrics writer
func NewWriter(config Config) *Writer {
	return &Writer{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WriteMetrics sends the metrics to Grafana Cloud using Prometheus remote write protocol
func (w *Writer) WriteMetrics(metrics []*io_prometheus_client.MetricFamily) error {
	// Convert MetricFamily to Prometheus remote write format
	timeseries := make([]prompb.TimeSeries, 0)

	for _, mf := range metrics {
		for _, m := range mf.Metric {
			// Build labels
			labels := make([]prompb.Label, 0)

			// Add metric name
			labels = append(labels, prompb.Label{
				Name:  "__name__",
				Value: *mf.Name,
			})

			// Add all other labels
			for _, lp := range m.Label {
				labels = append(labels, prompb.Label{
					Name:  *lp.Name,
					Value: *lp.Value,
				})
			}

			// Extract value and timestamp
			var value float64
			switch {
			case m.Gauge != nil:
				value = *m.Gauge.Value
			case m.Counter != nil:
				value = *m.Counter.Value
			case m.Untyped != nil:
				value = *m.Untyped.Value
			default:
				continue
			}

			// Use current time in milliseconds
			timestamp := time.Now().UnixMilli()

			// Create time series
			ts := prompb.TimeSeries{
				Labels: labels,
				Samples: []prompb.Sample{
					{
						Value:     value,
						Timestamp: timestamp,
					},
				},
			}

			timeseries = append(timeseries, ts)
		}
	}

	// Create write request
	wr := prompb.WriteRequest{
		Timeseries: timeseries,
	}

	// Marshal to protobuf
	data, err := wr.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf: %w", err)
	}

	// Compress with snappy
	compressed := snappy.Encode(nil, data)

	// Create HTTP request
	url := w.config.PrometheusURL + "/push"
	req, err := http.NewRequest("POST", url, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header (basic auth)
	auth := base64.StdEncoding.EncodeToString([]byte(w.config.Username + ":" + w.config.Password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	// Send request with retry logic for rate limits
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := w.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send request: %w", err)
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			errMsg := string(body)

			// If rate limited (429), retry with backoff
			if resp.StatusCode == 429 {
				if attempt < maxRetries-1 {
					// Wait before retrying (exponential backoff: 1s, 2s, 4s)
					waitTime := time.Duration(1<<uint(attempt)) * time.Second
					log.Printf("Rate limited. Retrying in %v... (attempt %d/%d)", waitTime, attempt+1, maxRetries)
					time.Sleep(waitTime)

					// Recreate the request body for retry (since it was consumed)
					buffer := &bytes.Buffer{}
					buffer.Write(compressed)
					req.Body = io.NopCloser(buffer)
					continue
				}
			}

			return fmt.Errorf("remote write failed with status %d: %s", resp.StatusCode, errMsg)
		}

		return nil
	}

	return fmt.Errorf("failed to send metrics after %d retries", maxRetries)
}
