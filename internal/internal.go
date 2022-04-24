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

	"eqrx.net/healthcheck/internal/check"
	"eqrx.net/matrix"
	"eqrx.net/matrix/room"
	"eqrx.net/rungroup"
	"eqrx.net/service"
	"github.com/go-logr/logr"
)

const confCredName = "healthcheck"

// MatrixConf defines in which matrix room messages should be sent.
type MatrixConf struct {
	Homeserver string `json:"homeserver"`
	Token      string `json:"token"`
	Room       string `json:"room"`
}

// Conf defines checks and sinks.
type Conf struct {
	Checks []check.Check `yaml:"checks"`
	Matrix MatrixConf    `yaml:"matrix"`
}

// Run loads checks and starts them.
func (c Conf) Run(ctx context.Context, log logr.Logger, service *service.Service) error {
	matrix := matrix.New(c.Matrix.Homeserver, c.Matrix.Token)
	room := room.New(matrix, c.Matrix.Room)

	if err := room.Join(ctx); err != nil {
		return fmt.Errorf("matrix room join: %w", err)
	}

	group := rungroup.New(ctx)

	for i := range c.Checks {
		if err := c.Checks[i].Setup(group, log, room); err != nil {
			return fmt.Errorf("check %s setup: %w", c.Checks[i].Name, err)
		}
	}

	if err := service.MarkReady(); err != nil {
		return fmt.Errorf("systemd notify: %w", err)
	}

	defer func() { _ = service.MarkStopping() }()

	if err := group.Wait(); err != nil {
		return fmt.Errorf("checks: %w", err)
	}

	return nil
}

// Run unmarshals Conf from systemd credentials and calls the Run method on it.
func Run(ctx context.Context, log logr.Logger, service *service.Service) error {
	var conf Conf
	if err := service.UnmarshalYAMLCreds(&conf, confCredName); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return conf.Run(ctx, log, service)
}
