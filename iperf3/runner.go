package iperf3

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

// TestResult contains the parsed iperf3 results
type TestResult struct {
	DownloadMbps  float64
	UploadMbps    float64
	JitterMs      float64
	LatencyMs     float64
	PacketLossPercent float64
}

// Iperf3Output is the structure of iperf3 JSON output
type Iperf3Output struct {
	Start struct {
		Connected []struct {
			Socket int `json:"socket"`
			LocalAddr string `json:"local_address"`
			LocalPort int `json:"local_port"`
			RemoteAddr string `json:"remote_address"`
			RemotePort int `json:"remote_port"`
		} `json:"connected"`
	} `json:"start"`
	Intervals []struct {
		Streams []struct {
			Socket int `json:"socket"`
			Start float64 `json:"start"`
			End float64 `json:"end"`
			Seconds float64 `json:"seconds"`
			Bytes int64 `json:"bytes"`
			BitsPerSecond float64 `json:"bits_per_second"`
			Retransmits int `json:"retransmits"`
			Snd_Cwnd int `json:"snd_cwnd"`
			Rtt int `json:"rtt"`
			Rttvar int `json:"rttvar"`
			Pmtu int `json:"pmtu"`
			Omitted bool `json:"omitted"`
		} `json:"streams"`
		Sum struct {
			Start float64 `json:"start"`
			End float64 `json:"end"`
			Seconds float64 `json:"seconds"`
			Bytes int64 `json:"bytes"`
			BitsPerSecond float64 `json:"bits_per_second"`
			Retransmits int `json:"retransmits"`
			Omitted bool `json:"omitted"`
		} `json:"sum"`
	} `json:"intervals"`
	End struct {
		Streams []struct {
			Socket int `json:"socket"`
			Start float64 `json:"start"`
			End float64 `json:"end"`
			Seconds float64 `json:"seconds"`
			Bytes int64 `json:"bytes"`
			BitsPerSecond float64 `json:"bits_per_second"`
			Retransmits int `json:"retransmits"`
			Snd_Cwnd int `json:"snd_cwnd"`
			Rtt int `json:"rtt"`
			Rttvar int `json:"rttvar"`
			Pmtu int `json:"pmtu"`
		} `json:"streams"`
		Sum struct {
			Start float64 `json:"start"`
			End float64 `json:"end"`
			Seconds float64 `json:"seconds"`
			Bytes int64 `json:"bytes"`
			BitsPerSecond float64 `json:"bits_per_second"`
			Retransmits int `json:"retransmits"`
		} `json:"sum"`
		SumSent struct {
			Start float64 `json:"start"`
			End float64 `json:"end"`
			Seconds float64 `json:"seconds"`
			Bytes int64 `json:"bytes"`
			BitsPerSecond float64 `json:"bits_per_second"`
			Retransmits int `json:"retransmits"`
			Sender bool `json:"sender"`
		} `json:"sum_sent"`
		SumReceived struct {
			Start float64 `json:"start"`
			End float64 `json:"end"`
			Seconds float64 `json:"seconds"`
			Bytes int64 `json:"bytes"`
			BitsPerSecond float64 `json:"bits_per_second"`
			Sender bool `json:"sender"`
		} `json:"sum_received"`
	} `json:"end"`
}

// RunTest executes iperf3 test against the server
func RunTest(server string, port int, duration int, reverse bool) (*TestResult, error) {
	args := []string{
		"-c", server,
		"-p", strconv.Itoa(port),
		"-t", strconv.Itoa(duration),
		"-J", // JSON output
	}

	if reverse {
		args = append(args, "-R") // Reverse test (server sends to client - download)
	}

	cmd := exec.Command("iperf3", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("iperf3 command failed: %w, output: %s", err, string(output))
	}

	var iperf3Out Iperf3Output
	if err := json.Unmarshal(output, &iperf3Out); err != nil {
		return nil, fmt.Errorf("failed to parse iperf3 JSON output: %w", err)
	}

	result := &TestResult{}

	// Extract throughput from end summary
	if len(iperf3Out.End.Streams) > 0 {
		// Get the bits per second value
		// For reverse tests, the data is in SumReceived; for normal tests, it's in Sum
		var bitsPerSecond float64
		if reverse && iperf3Out.End.SumReceived.BitsPerSecond > 0 {
			// Reverse test - use the received (downloaded) data
			bitsPerSecond = iperf3Out.End.SumReceived.BitsPerSecond
		} else if !reverse && iperf3Out.End.SumSent.BitsPerSecond > 0 {
			// Normal test - use the sent (uploaded) data
			bitsPerSecond = iperf3Out.End.SumSent.BitsPerSecond
		} else {
			// Fallback to Sum field
			bitsPerSecond = iperf3Out.End.Sum.BitsPerSecond
		}

		mbps := bitsPerSecond / 1_000_000

		if reverse {
			result.DownloadMbps = mbps
		} else {
			result.UploadMbps = mbps
		}

		// Extract RTT (latency) and jitter from last stream
		lastStream := iperf3Out.End.Streams[len(iperf3Out.End.Streams)-1]
		result.LatencyMs = float64(lastStream.Rtt) / 1000.0 // Convert microseconds to ms
		result.JitterMs = float64(lastStream.Rttvar) / 1000.0 // Rttvar is jitter
	}

	// Packet loss is typically calculated from retransmits, but iperf3 doesn't directly provide it
	// We'll set it to 0 for now - you might need to enhance this with additional metrics
	result.PacketLossPercent = 0

	return result, nil
}

// RunBothTests runs both download and upload tests, with graceful fallback and retries
func RunBothTests(server string, port int, duration int) (*TestResult, error) {
	result := &TestResult{}
	maxRetries := 2

	// Try download test (reverse) first
	var downloadResult *TestResult
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		downloadResult, err = RunTest(server, port, duration, true)
		if err == nil {
			result.DownloadMbps = downloadResult.DownloadMbps
			result.LatencyMs = downloadResult.LatencyMs
			result.JitterMs = downloadResult.JitterMs
			break
		}
		if attempt < maxRetries-1 {
			log.Printf("Download test attempt %d failed, retrying...", attempt+1)
		}
	}
	if err != nil {
		log.Printf("Warning: Download test failed after %d attempts", maxRetries)
	}

	// Try upload test (normal) with retries
	var uploadResult *TestResult
	for attempt := 0; attempt < maxRetries; attempt++ {
		uploadResult, err = RunTest(server, port, duration, false)
		if err == nil {
			break
		}
		if attempt < maxRetries-1 {
			log.Printf("Upload test attempt %d failed, retrying...", attempt+1)
		}
	}

	if err != nil {
		log.Printf("Upload test failed after %d attempts", maxRetries)
		// If both tests failed, return error
		if result.DownloadMbps == 0 {
			return nil, fmt.Errorf("both download and upload tests failed")
		}
		// But if download succeeded, we can still return partial results
		return result, nil
	}

	// Combine results
	result.UploadMbps = uploadResult.UploadMbps

	// Use better latency/jitter if available
	if uploadResult.LatencyMs > 0 && (result.LatencyMs == 0 || uploadResult.LatencyMs < result.LatencyMs) {
		result.LatencyMs = uploadResult.LatencyMs
	}
	if uploadResult.JitterMs > 0 && (result.JitterMs == 0 || uploadResult.JitterMs < result.JitterMs) {
		result.JitterMs = uploadResult.JitterMs
	}

	return result, nil
}
