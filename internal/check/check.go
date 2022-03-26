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

// Package check contains service checks.
package check

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eqrx.net/healthcheck/internal/check/ceph"
	"eqrx.net/healthcheck/internal/check/matrix"
	"eqrx.net/healthcheck/internal/check/smtp"
	"eqrx.net/healthcheck/internal/sink"
	matrixsink "eqrx.net/healthcheck/internal/sink/matrix"
	"eqrx.net/rungroup"
	"github.com/go-logr/logr"
)

var errConcrete = errors.New("more or less than one concrete types set for check")

// Check contains concrete check implementations.
type Check struct {
	Matrix   *matrix.Check                            `yaml:"matrix"`
	SMTP     *smtp.Check                              `yaml:"smtp"`
	Ceph     *ceph.Check                              `yaml:"ceph"`
	Sinks    []sink.Sink                              `yaml:"sinks"`
	Interval time.Duration                            `yaml:"interval"`
	Name     string                                   `yaml:"name"`
	checkCB  func(context.Context, logr.Logger) error `yaml:"-"`
}

// Setup starts the check.
func (c *Check) Setup(group *rungroup.Group, log logr.Logger, matrix *matrixsink.Matrix) error {
	if c.Matrix != nil {
		if c.checkCB != nil {
			return errConcrete
		}

		c.Matrix.Setup()

		c.checkCB = c.Matrix.Check
	}

	if c.SMTP != nil {
		if c.checkCB != nil {
			return errConcrete
		}

		c.SMTP.Setup()

		c.checkCB = c.SMTP.Check
	}

	if c.Ceph != nil {
		if c.checkCB != nil {
			return errConcrete
		}

		c.checkCB = c.Ceph.Check
	}

	if c.checkCB == nil {
		return errConcrete
	}

	for i := range c.Sinks {
		if err := c.Sinks[i].Setup(c.Name, matrix); err != nil {
			return fmt.Errorf("setup sink: %w", err)
		}
	}

	group.Go(func(ctx context.Context) error {
		c.poll(ctx, log, c.Interval)

		return nil
	})

	return nil
}

func (c Check) poll(ctx context.Context, log logr.Logger, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		checkErr := c.check(ctx, log)

		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := c.sink(ctx, checkErr); err != nil {
			log.Error(err, "sinks error")
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (c Check) check(ctx context.Context, log logr.Logger) error {
	timeout := c.Interval / 2 //nolint: gomnd
	ctx, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()

	return c.checkCB(ctx, log)
}

func (c Check) sink(ctx context.Context, checkErr error) error {
	timeout := c.Interval / 2 //nolint: gomnd
	ctx, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()

	group := rungroup.New(ctx)

	for i := range c.Sinks {
		sink := &c.Sinks[i]

		group.Go(func(ctx context.Context) error {
			return sink.Sink(ctx, checkErr) //nolint:wrapcheck
		}, rungroup.NeverCancel)
	}

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}
