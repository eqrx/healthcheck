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

// Package internal contains all the code.
package internal

import (
	"context"
	"fmt"
	"os"

	"eqrx.net/healthcheck/internal/check"
	"eqrx.net/rungroup"
	"eqrx.net/service"
	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
)

// ConfigPath defines from where the configuration file is loaded.
const ConfigPath = "/etc/healthcheck/conf"

// Conf defines checks and sinks.
type Conf struct {
	Checks []check.Check `yaml:"checks"`
}

// Run loads checks and starts them.
func (c Conf) Run(ctx context.Context, log logr.Logger, service service.Service) error {
	group := rungroup.New(ctx)

	for i := range c.Checks {
		if err := c.Checks[i].Setup(ctx, group, log); err != nil {
			return fmt.Errorf("check %s setup: %w", c.Checks[i].Name, err)
		}
	}

	group.Go(service.RunNotify)

	if err := group.Wait(); err != nil {
		return fmt.Errorf("checks: %w", err)
	}

	return nil
}

// Run unmarshals Conf from systemd credentials and calls the Run method on it.
func Run(ctx context.Context, log logr.Logger, service service.Service) error {
	var conf Conf

	confBytes, err := os.ReadFile(ConfigPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(confBytes, &conf); err != nil {
		return err
	}

	return conf.Run(ctx, log, service)
}
