// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package telemetry

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/sdk/resource"
)

func TestParseOTLPEndpoint(t *testing.T) {
	cases := []struct {
		name         string
		endpoint     string
		wantAddr     string
		wantInsecure bool
	}{
		{"http scheme strips and is insecure", "http://192.0.2.1:4317", "192.0.2.1:4317", true},
		{"https scheme strips and keeps TLS", "https://otel.example.com:4317", "otel.example.com:4317", false},
		{"bare host:port unchanged and keeps TLS", "localhost:4317", "localhost:4317", false},
		{"uppercase HTTP scheme strips and is insecure", "HTTP://192.0.2.1:4317", "192.0.2.1:4317", true},
		{"mixed-case Https scheme strips and keeps TLS", "Https://otel.example.com:4317", "otel.example.com:4317", false},
		{"endpoint shorter than scheme prefix is unchanged", "ht", "ht", false},
		{"http scheme with trailing slash is trimmed", "http://192.0.2.1:4317/", "192.0.2.1:4317", true},
		{"https scheme with trailing slash is trimmed", "https://otel.example.com:4317/", "otel.example.com:4317", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			addr, insecure := parseOTLPEndpoint(tc.endpoint)
			if addr != tc.wantAddr {
				t.Errorf("addr = %q, want %q", addr, tc.wantAddr)
			}
			if insecure != tc.wantInsecure {
				t.Errorf("insecure = %v, want %v", insecure, tc.wantInsecure)
			}
		})
	}
}

func TestNewStdoutMetricExporter(t *testing.T) {
	exp, err := newStdoutMetricExporter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp == nil {
		t.Error("expected non-nil exporter")
	}
}

func TestInitConsoleProviders(t *testing.T) {
	meterProvider = nil
	shutdownFuncs = nil
	defer func() {
		for _, fn := range shutdownFuncs {
			_ = fn(context.Background())
		}
		meterProvider = nil
		shutdownFuncs = nil
	}()

	initConsoleProviders(resource.Default())
	if meterProvider == nil {
		t.Error("expected meterProvider to be set after initConsoleProviders")
	}
	if len(shutdownFuncs) != 1 {
		t.Errorf("expected 1 shutdown func, got %d", len(shutdownFuncs))
	}
}

func TestInitOTLPProviders_InvalidEndpoint(t *testing.T) {
	meterProvider = nil
	shutdownFuncs = nil
	defer func() {
		for _, fn := range shutdownFuncs {
			_ = fn(context.Background())
		}
		meterProvider = nil
		shutdownFuncs = nil
	}()

	cfg := Config{
		Exporter:     "otlp",
		OTLPEndpoint: "localhost:0",
	}
	initOTLPProviders(context.Background(), resource.Default(), cfg)
	if meterProvider == nil {
		t.Error("expected meterProvider to be set (OTLP exporter creation is lazy)")
	}
}

func TestInitOTLPHTTPProviders_InvalidEndpoint(t *testing.T) {
	meterProvider = nil
	shutdownFuncs = nil
	defer func() {
		for _, fn := range shutdownFuncs {
			_ = fn(context.Background())
		}
		meterProvider = nil
		shutdownFuncs = nil
	}()

	cfg := Config{
		Exporter:     "otlp",
		OTLPEndpoint: "localhost:0",
		OTLPProtocol: "http/protobuf",
	}
	initOTLPHTTPProviders(context.Background(), resource.Default(), cfg)
	if meterProvider == nil {
		t.Error("expected meterProvider to be set (OTLP HTTP exporter creation is lazy)")
	}
}

func TestInitOTLPProviders_ProtocolRouting(t *testing.T) {
	cases := []struct {
		name        string
		protocol    string
		wantWarning bool // default branch emits a gRPC fallback warning
	}{
		{"grpc default", "grpc", false},
		{"empty defaults to grpc", "", false},
		{"http/protobuf routes to http", "http/protobuf", false},
		{"http/json routes to http", "http/json", false},
		{"unknown falls back to grpc", "foo", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			meterProvider = nil
			shutdownFuncs = nil

			// Capture stderr to assert on the fallback warning.
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			cfg := Config{
				Exporter:     "otlp",
				OTLPEndpoint: "localhost:0",
				OTLPProtocol: tc.protocol,
			}
			initOTLPProviders(context.Background(), resource.Default(), cfg)
			if err := w.Close(); err != nil {
				t.Fatalf("close stderr pipe: %v", err)
			}
			os.Stderr = oldStderr

			defer func() {
				for _, fn := range shutdownFuncs {
					_ = fn(context.Background())
				}
				meterProvider = nil
				shutdownFuncs = nil
			}()

		if meterProvider == nil {
			t.Error("expected meterProvider to be set")
		}
		if len(shutdownFuncs) != 1 {
			t.Errorf("expected 1 shutdown func, got %d", len(shutdownFuncs))
					t.Errorf("expected gRPC fallback warning in stderr, got %q", stderrOut)
				}
			} else {
				if strings.Contains(stderrOut, "falling back to gRPC") {
					t.Errorf("expected no fallback warning, got %q", stderrOut)
				}
			}
		})
	}
}

func TestOTLPSignalURL(t *testing.T) {
	cases := []struct {
		name     string
		endpoint string
		signal   string
		want     string
	}{
		{"http scheme keeps its scheme", "http://192.0.2.1:4318", "/v1/metrics", "http://192.0.2.1:4318/v1/metrics"},
		{"https scheme keeps its scheme", "https://otel.example.com:4318", "/v1/metrics", "https://otel.example.com:4318/v1/metrics"},
		{"bare host:port gains https to keep TLS", "localhost:4318", "/v1/metrics", "https://localhost:4318/v1/metrics"},
		{"a base path is preserved", "http://langfuse:3000/api/public/otel", "/v1/metrics", "http://langfuse:3000/api/public/otel/v1/metrics"},
		{"a trailing slash does not double up", "http://192.0.2.1:4318/", "/v1/metrics", "http://192.0.2.1:4318/v1/metrics"},
		{"an endpoint already naming the signal is left alone", "http://192.0.2.1:4318/v1/metrics", "/v1/metrics", "http://192.0.2.1:4318/v1/metrics"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := otlpSignalURL(tc.endpoint, tc.signal); got != tc.want {
				t.Errorf("otlpSignalURL(%q, %q) = %q, want %q", tc.endpoint, tc.signal, got, tc.want)
			}
		})
	}
}

