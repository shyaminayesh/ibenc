package remote

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/prometheus/client_model/go"
)

// WriteMetricsText sends metrics using Prometheus text exposition format (alternative to protobuf)
func (w *Writer) WriteMetricsText(metrics []*io_prometheus_client.MetricFamily) error {
	buffer := &bytes.Buffer{}

	for _, mf := range metrics {
		// Write HELP line
		if mf.Help != nil {
			fmt.Fprintf(buffer, "# HELP %s %s\n", *mf.Name, *mf.Help)
		}
		// Write TYPE line
		if mf.Type != nil {
			typeStr := mf.Type.String()
			// Convert type to prometheus format
			promType := "gauge"
			switch typeStr {
			case "COUNTER":
				promType = "counter"
			case "HISTOGRAM":
				promType = "histogram"
			case "SUMMARY":
				promType = "summary"
			}
			fmt.Fprintf(buffer, "# TYPE %s %s\n", *mf.Name, promType)
		}

		// Write metric samples
		for _, m := range mf.Metric {
			metricName := *mf.Name
			labels := buildLabelsString(m.Label)
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

			timestamp := time.Now().UnixMilli()
			if m.TimestampMs != nil {
				timestamp = *m.TimestampMs
			}

			fmt.Fprintf(buffer, "%s%s %v %d\n", metricName, labels, value, timestamp)
		}
	}

	// Create HTTP request to the text format endpoint
	// Most Prometheus-compatible systems accept text format at /api/v1/write or /metrics
	url := w.config.PrometheusURL + "/metrics/write"
	req, err := http.NewRequest("POST", url, buffer)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication header (basic auth)
	auth := base64.StdEncoding.EncodeToString([]byte(w.config.Username + ":" + w.config.Password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	// Send request
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remote write failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// buildLabelsString creates label string from label pairs
func buildLabelsString(labelPairs []*io_prometheus_client.LabelPair) string {
	if len(labelPairs) == 0 {
		return ""
	}

	labelsStr := "{"
	for i, lp := range labelPairs {
		if i > 0 {
			labelsStr += ","
		}
		labelsStr += fmt.Sprintf(`%s="%s"`, *lp.Name, *lp.Value)
	}
	labelsStr += "}"

	return labelsStr
}
