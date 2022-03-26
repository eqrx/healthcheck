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

// Package smtp contains health checks for SMTP homeservers.
package smtp

import (
	"context"
	"errors"
	"fmt"
	"net"

	"eqrx.net/rungroup"
	"github.com/go-logr/logr"
	"github.com/miekg/dns"
)

var errNX = errors.New("no servers defined for addr")

// Check resolves an SMTP server and tess it TLS function.
type Check struct {
	IPV4         bool   `yaml:"ipv4"`
	Domain       string `yaml:"domain"`
	targetRRType uint16 `yaml:"-"`
	network      string `yaml:"-"`
}

// Setup prepares often used values.
func (c *Check) Setup() {
	c.targetRRType = dns.TypeAAAA

	if c.IPV4 {
		c.targetRRType = dns.TypeA
	}

	c.network = "tcp6"
	if c.IPV4 {
		c.network = "tcp4"
	}
}

// Check resolves the SMTP of the domain and connects to it via TLS.
// TODO: DMARC and all the fun stuff.
// TODO: validate all dns names.
func (c Check) Check(ctx context.Context, _ logr.Logger) error {
	addrs, err := c.resolveServer(ctx)
	if err != nil {
		return fmt.Errorf("smtp check: resolve server: %w", err)
	}

	if len(addrs) == 0 {
		return fmt.Errorf("smtp check: %w", errNX)
	}

	group := rungroup.New(ctx)

	for i := range addrs {
		hostPort := net.JoinHostPort(addrs[i], smtpPort)

		group.Go(func(ctx context.Context) error {
			if err := c.connect(ctx, hostPort); err != nil {
				return fmt.Errorf("connect %s: %w", hostPort, err)
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("smtp check: %w", err)
	}

	return nil
}
