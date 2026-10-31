// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package telemetry

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Global OTel providers. Set by Init, used via otel.GetMeterProvider().
var (
	meterProvider *sdkmetric.MeterProvider
	shutdownFuncs []func(context.Context) error
	initialized   bool
)

// serviceName holds the name set during Init.
var serviceName = "open-code-review"

// Init initializes global MeterProvider based on
// environment variables and optional config file. Returns true when enabled.
// Safe to call multiple times.
func Init(ctx context.Context) bool {
	if initialized {
		return len(shutdownFuncs) > 0
	}
	initialized = true

	cfg := ResolveConfig(HomeConfigPath())
	serviceName = cfg.ServiceName
	if !cfg.Enabled {
		return false
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithHost(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ocr] WARNING: failed to create OTel resource: %v\n", err)
		res = resource.Default()
	}

	switch cfg.Exporter {
	case "otlp":
		initOTLPProviders(ctx, res, cfg)
	default:
		initConsoleProviders(res)
	}

	otel.SetMeterProvider(meterProvider)

	return len(shutdownFuncs) > 0
}

// IsEnabled returns true when telemetry has been initialized with exporters.
func IsEnabled() bool {
	return initialized && len(shutdownFuncs) > 0
}
