// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package hostmetricsreceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol/otelcoltest"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/filterset"
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

func TestLoadConfig(t *testing.T) {
	factories, err := otelcoltest.NopFactories()
	require.NoError(t, err)

	factory := NewFactory()
	factories.Receivers[metadata.Type] = factory
	// https://opentelemetry.io/collector/issues/33594
	// nolint:staticcheck
	cfg, err := otelcoltest.LoadConfigAndValidate(filepath.Join("testdata", "config.yaml"), factories)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Len(t, cfg.Receivers, 2)

	r0 := cfg.Receivers[component.NewID(metadata.Type)]
	defaultConfigCPUScraper := factory.CreateDefaultConfig()
	defaultConfigCPUScraper.(*Config).Scrapers = map[string]internal.Config{
		cpuscraper.TypeStr: func() internal.Config {
			cfg := (&cpuscraper.Factory{}).CreateDefaultConfig()
			cfg.SetEnvMap(common.EnvMap{})
			return cfg
		}(),
	}

	assert.Equal(t, defaultConfigCPUScraper, r0)

	r1 := cfg.Receivers[component.NewIDWithName(metadata.Type, "customname")].(*Config)
	expectedConfig := &Config{
		MetadataCollectionInterval: 5 * time.Minute,
		ControllerConfig: scraperhelper.ControllerConfig{
			CollectionInterval: 30 * time.Second,
			InitialDelay:       time.Second,
		},
		Scrapers: map[string]internal.Config{
			cpuscraper.TypeStr: func() internal.Config {
				cfg := (&cpuscraper.Factory{}).CreateDefaultConfig()
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			}(),
			diskscraper.TypeStr: func() internal.Config {
				cfg := (&diskscraper.Factory{}).CreateDefaultConfig()
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			}(),
			loadscraper.TypeStr: (func() internal.Config {
				cfg := (&loadscraper.Factory{}).CreateDefaultConfig()
				cfg.(*loadscraper.Config).CPUAverage = true
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			})(),
			filesystemscraper.TypeStr: func() internal.Config {
				cfg := (&filesystemscraper.Factory{}).CreateDefaultConfig()
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			}(),
			memoryscraper.TypeStr: func() internal.Config {
				cfg := (&memoryscraper.Factory{}).CreateDefaultConfig()
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			}(),
			networkscraper.TypeStr: (func() internal.Config {
				cfg := (&networkscraper.Factory{}).CreateDefaultConfig()
				cfg.(*networkscraper.Config).Include = networkscraper.MatchConfig{
					Interfaces: []string{"test1"},
					Config:     filterset.Config{MatchType: "strict"},
				}
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			})(),
			processesscraper.TypeStr: func() internal.Config {
				cfg := (&processesscraper.Factory{}).CreateDefaultConfig()
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			}(),
			pagingscraper.TypeStr: func() internal.Config {
				cfg := (&pagingscraper.Factory{}).CreateDefaultConfig()
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			}(),
			processscraper.TypeStr: (func() internal.Config {
				cfg := (&processscraper.Factory{}).CreateDefaultConfig()
				cfg.(*processscraper.Config).Include = processscraper.MatchConfig{
					Names:  []string{"test2", "test3"},
					Config: filterset.Config{MatchType: "regexp"},
				}
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			})(),
			systemscraper.TypeStr: (func() internal.Config {
				cfg := (&systemscraper.Factory{}).CreateDefaultConfig()
				cfg.SetEnvMap(common.EnvMap{})
				return cfg
			})(),
		},
	}

	assert.Equal(t, expectedConfig, r1)
}

func TestLoadInvalidConfig_NoScrapers(t *testing.T) {
	factories, err := otelcoltest.NopFactories()
	require.NoError(t, err)

	factory := NewFactory()
	factories.Receivers[metadata.Type] = factory
	// https://opentelemetry.io/collector/issues/33594
	// nolint:staticcheck
	_, err = otelcoltest.LoadConfigAndValidate(filepath.Join("testdata", "config-noscrapers.yaml"), factories)

	require.ErrorContains(t, err, "must specify at least one scraper when using hostmetrics receiver")
}

func TestLoadInvalidConfig_InvalidScraperKey(t *testing.T) {
	factories, err := otelcoltest.NopFactories()
	require.NoError(t, err)

	factory := NewFactory()
	factories.Receivers[metadata.Type] = factory
	// https://opentelemetry.io/collector/issues/33594
	// nolint:staticcheck
	_, err = otelcoltest.LoadConfigAndValidate(filepath.Join("testdata", "config-invalidscraperkey.yaml"), factories)

	require.ErrorContains(t, err, "error reading configuration for \"hostmetrics\": invalid scraper key: invalidscraperkey")
}
