// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package handlecount // import "opentelemetry.io/collector/receiver/hostmetricsreceiver/internal/scraper/processscraper/internal/handlecount"

type Manager interface {
	Refresh() error
	GetProcessHandleCount(pid int64) (uint32, error)
}
