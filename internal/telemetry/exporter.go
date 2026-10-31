// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package telemetry

import (
	"context"
	"fmt"
	"os"
	"strings"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	otlpmetricgrpc "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otlpmetrichttp "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	stdoutmetric "go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
)

// newStdoutMetricExporter creates a console metric exporter with pretty-print output.
func newStdoutMetricExporter() (sdkmetric.Exporter, error) {
	return stdoutmetric.New(stdoutmetric.WithPrettyPrint())
}

// parseOTLPEndpoint strips a http:// or https:// scheme from the endpoint and
// reports whether the connection should be plaintext (insecure).
// Scheme matching is case-insensitive per RFC 3986. A bare host:port (no
// scheme) is left unchanged and defaults to TLS. Any trailing slash left
// over from a URL-style endpoint (e.g. "http://localhost:4317/") is
// trimmed, since both gRPC and HTTP exporters' WithEndpoint expects a bare
// host:port with no path.
func parseOTLPEndpoint(endpoint string) (addr string, insecure bool) {
	switch {
	case len(endpoint) >= 7 && strings.EqualFold(endpoint[:7], "http://"):
		return strings.TrimRight(endpoint[7:], "/"), true
	case len(endpoint) >= 8 && strings.EqualFold(endpoint[:8], "https://"):
		return strings.TrimRight(endpoint[8:], "/"), false
	default:
		return endpoint, false
	}
}

// otlpSignalURL builds the URL for one signal from the configured endpoint.
//
// OTEL_EXPORTER_OTLP_ENDPOINT is defined by the OpenTelemetry specification as a BASE
// URL for the HTTP protocols, to which the per-signal path is appended: an endpoint of
// "https://backend.example.com" sends metrics to "https://backend.example.com/v1/metrics".
// WithEndpointURL, by contrast, takes the full signal URL — it assigns u.Path straight
// to URLPath — so the signal path has to be appended here rather than left to the SDK.
//
// Two details this has to get right:
//
//   - A bare host:port keeps the meaning parseOTLPEndpoint gave it, TLS, by gaining an
//     https:// scheme. WithEndpointURL derives Insecure from the scheme
//     (`u.Scheme != "https"`), which is what the explicit WithInsecure call used to do,
//     so an existing configuration behaves identically after this change.
//   - An endpoint that already names the signal path is left alone. Someone who worked
//     around the missing path support by writing it out in full should not end up with
//     "/v1/metrics/v1/metrics".
func otlpSignalURL(endpoint, signalPath string) string {
	base := endpoint
	switch {
	case len(base) >= 7 && strings.EqualFold(base[:7], "http://"):
	case len(base) >= 8 && strings.EqualFold(base[:8], "https://"):
	default:
		base = "https://" + base
	}

	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, signalPath) {
		return base
	}
	return base + signalPath
}

// initOTLPProviders dispatches to the gRPC or HTTP exporter based on cfg.OTLPProtocol.
func initOTLPProviders(ctx context.Context, res *resource.Resource, cfg Config) {
	switch cfg.OTLPProtocol {
	case "http/protobuf", "http/json":
		initOTLPHTTPProviders(ctx, res, cfg)
	case "", "grpc":
		initOTLPGRPCProviders(ctx, res, cfg)
	default:
		fmt.Fprintf(os.Stderr, "[ocr] WARNING: unsupported OTLP protocol %q, falling back to gRPC\n", cfg.OTLPProtocol)
		initOTLPGRPCProviders(ctx, res, cfg)
	}
}

// initOTLPGRPCProviders sets up OTLP gRPC metric exporter.
func initOTLPGRPCProviders(ctx context.Context, res *resource.Resource, cfg Config) {
	addr, insecure := parseOTLPEndpoint(cfg.OTLPEndpoint)

	metricOpts := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(addr)}
	if insecure {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}
	metricExp, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ocr] WARNING: failed to create OTLP metric exporter: %v\n", err)
		return
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		sdkmetric.WithResource(res),
	)
	meterProvider = mp
	shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error { return mp.Shutdown(ctx) })
}

// initOTLPHTTPProviders sets up OTLP HTTP metric exporter.
//
// Unlike the gRPC path this uses WithEndpointURL rather than WithEndpoint, so that a
// base path in OTEL_EXPORTER_OTLP_ENDPOINT is honoured. WithEndpoint takes a bare
// host:port and cannot express a path, so every request went to /v1/metrics at the root
// of the host regardless of what the endpoint said.
//
// That silently excludes any backend reached through a path prefix. Langfuse receives
// OTLP at /api/public/otel, for instance, so an endpoint of
// "http://langfuse:3000/api/public/otel" produced requests to
// "http://langfuse:3000/v1/metrics" and every metric was dropped with a 404.
//
// The signal path is appended by otlpSignalURL rather than by the SDK: WithEndpointURL
// assigns the URL's path straight to URLPath, so it expects the full signal URL, while
// OTEL_EXPORTER_OTLP_ENDPOINT is specified as a base. It also derives TLS from the
// scheme, which is what the separate WithInsecure call was doing by hand.
func initOTLPHTTPProviders(ctx context.Context, res *resource.Resource, cfg Config) {
	metricExp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpointURL(otlpSignalURL(cfg.OTLPEndpoint, "/v1/metrics")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ocr] WARNING: failed to create OTLP HTTP metric exporter: %v\n", err)
		return
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		sdkmetric.WithResource(res),
	)
	meterProvider = mp
	shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error { return mp.Shutdown(ctx) })
}

// initConsoleProviders sets up stdout metric exporter.
func initConsoleProviders(res *resource.Resource) {
	metricExp, err := newStdoutMetricExporter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ocr] WARNING: failed to create console metric exporter: %v\n", err)
		return
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		sdkmetric.WithResource(res),
	)
	meterProvider = mp
	shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error { return mp.Shutdown(ctx) })
}
