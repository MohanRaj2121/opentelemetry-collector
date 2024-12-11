// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cpuscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper"

import (
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/metadata"
)

// Config relating to CPU Metric Scraper.
type Config struct {
	metadata.MetricsBuilderConfig `mapstructure:",squash"`
	internal.ScraperConfig
}
