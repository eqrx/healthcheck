// Copyright (C) 2022 Alexander Sowitzki
//
// This program is free software: you can redistribute it and/or modify it under the terms of the
// GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
// warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License along with this program.
// If not, see <https://www.gnu.org/licenses/>.

// Package ceph contains health checks for ceph.
package ceph

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"eqrx.net/service"
	"github.com/go-logr/logr"
)

// StatusOK is the value that ceph reports as health status when everything is good.
const StatusOK = "HEALTH_OK"

// Check for the status of a ceph cluster.
type Check struct {
	ClientName string `yaml:"clientName"`
	CredsName  string `yaml:"credsName"`
}

// Check uses exec to execute the command `ceph status -f json` to get the current status of the cluster that is used
// by the host healthcheck is running on. If the exec succeeds the output is unmarshalled into Report.  Lastly if the
// field Report->Health->Status is equal to the constant StatusOK nil is returned.
func (c Check) Check(ctx context.Context, _ logr.Logger) error {
	credDir, err := service.CredsDir()
	if err != nil {
		return fmt.Errorf("ceph key ring: %w", err)
	}

	args := []string{"status", "-n", c.ClientName, "-k", path.Join(credDir, c.CredsName)}
	cmd := exec.CommandContext(ctx, "ceph", args...)

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("query ceph for status: %w", err)
	}

	status := string(out)

	if !strings.Contains(status, StatusOK) {
		return fmt.Errorf("ceph reports unhealthy: %v", status)
	}

	return nil
}
