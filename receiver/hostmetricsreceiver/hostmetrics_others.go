// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package hostmetricsreceiver // import "opentelemetry.io/collector/receiver/hostmetricsreceiver"

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/common"
)

func validateRootPath(rootPath string) error {
	if rootPath == "" {
		return nil
	}
	return fmt.Errorf("root_path is supported on linux only")
}

func setGoPsutilEnvVars(_ string, _ environment) common.EnvMap {
	return common.EnvMap{}
}
