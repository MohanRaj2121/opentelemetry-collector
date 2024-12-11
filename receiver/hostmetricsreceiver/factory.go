// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package hostmetricsreceiver // import "opentelemetry.io/collector/receiver/hostmetricsreceiver"

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/process"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.opentelemetry.io/collector/scraper"

	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/metadata"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/cpuscraper"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/diskscraper"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/filesystemscraper"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/loadscraper"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/memoryscraper"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/networkscraper"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/pagingscraper"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/processesscraper"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/processscraper"
	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/systemscraper"
)

const (
	defaultMetadataCollectionInterval = 5 * time.Minute
)

// This file implements Factory for HostMetrics receiver.
var (
	scraperFactories = map[string]internal.ScraperFactory{
		cpuscraper.TypeStr:        &cpuscraper.Factory{},
		diskscraper.TypeStr:       &diskscraper.Factory{},
		loadscraper.TypeStr:       &loadscraper.Factory{},
		filesystemscraper.TypeStr: &filesystemscraper.Factory{},
		memoryscraper.TypeStr:     &memoryscraper.Factory{},
		networkscraper.TypeStr:    &networkscraper.Factory{},
		pagingscraper.TypeStr:     &pagingscraper.Factory{},
		processesscraper.TypeStr:  &processesscraper.Factory{},
		processscraper.TypeStr:    &processscraper.Factory{},
		systemscraper.TypeStr:     &systemscraper.Factory{},
	}
)

// NewFactory creates a new factory for host metrics receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability),
		receiver.WithLogs(createLogsReceiver, metadata.LogsStability))
}

func getScraperFactory(key string) (internal.ScraperFactory, bool) {
	if factory, ok := scraperFactories[key]; ok {
		return factory, true
	}

	return nil, false
}

// createDefaultConfig creates the default configuration for receiver.
func createDefaultConfig() component.Config {
	return &Config{
		ControllerConfig:           scraperhelper.NewDefaultControllerConfig(),
		MetadataCollectionInterval: defaultMetadataCollectionInterval,
	}
}

// createMetricsReceiver creates a metrics receiver based on provided config.
func createMetricsReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	oCfg := cfg.(*Config)

	addScraperOptions, err := createAddScraperOptions(ctx, set, oCfg, scraperFactories)
	if err != nil {
		return nil, err
	}

	host.EnableBootTimeCache(true)
	process.EnableBootTimeCache(true)

	return scraperhelper.NewScraperControllerReceiver(
		&oCfg.ControllerConfig,
		set,
		consumer,
		addScraperOptions...,
	)
}

func createLogsReceiver(
	_ context.Context, set receiver.Settings, cfg component.Config, consumer consumer.Logs,
) (receiver.Logs, error) {
	return &hostEntitiesReceiver{
		cfg:      cfg.(*Config),
		nextLogs: consumer,
		settings: &set,
	}, nil
}

func createAddScraperOptions(
	ctx context.Context,
	set receiver.Settings,
	config *Config,
	factories map[string]internal.ScraperFactory,
) ([]scraperhelper.ScraperControllerOption, error) {
	scraperControllerOptions := make([]scraperhelper.ScraperControllerOption, 0, len(config.Scrapers))

	for key, cfg := range config.Scrapers {
		hostMetricsScraper, ok, err := createHostMetricsScraper(ctx, set, key, cfg, factories)
		if err != nil {
			return nil, fmt.Errorf("failed to create scraper for key %q: %w", key, err)
		}

		if ok {
			scraperControllerOptions = append(scraperControllerOptions, scraperhelper.AddScraper(metadata.Type, hostMetricsScraper))
			continue
		}

		return nil, fmt.Errorf("host metrics scraper factory not found for key: %q", key)
	}

	return scraperControllerOptions, nil
}

func createHostMetricsScraper(ctx context.Context, set receiver.Settings, key string, cfg internal.Config, factories map[string]internal.ScraperFactory) (s scraper.Metrics, ok bool, err error) {
	factory := factories[key]
	if factory == nil {
		ok = false
		return
	}

	ok = true
	s, err = factory.CreateMetricsScraper(ctx, set, cfg)
	return
}

type environment interface {
	Lookup(k string) (string, bool)
}

type osEnv struct{}

var _ environment = (*osEnv)(nil)

func (e *osEnv) Lookup(k string) (string, bool) {
	return os.LookupEnv(k)
}
