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

// Package matrix contains health checks for matrix homeservers.
package matrix

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"

	"eqrx.net/rungroup"
	"github.com/go-logr/logr"
	"github.com/miekg/dns"
)

// StatusOK is the value that synapse reports as health status when everything is good.
const StatusOK = "OK"

type target struct {
	url  url.URL
	addr netip.AddrPort
}

type wellKnown struct {
	Homeserver struct {
		URL string `json:"base_url"`
	} `json:"m.homeserver"`
}

// Check for testing if a homeserver is reachable via HTTPS.
type Check struct {
	IPV4         bool   `yaml:"ipv4"`
	Domain       string `yaml:"domain"`
	targetRRType uint16 `yaml:"-"`
	network      string `yaml:"-"`
}

// Setup the check by preparing often used values.
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

// Check resolved the well-known info and the SRV record of the given domain.
// If both match all homeservers are connected to via HTTP.
func (c Check) Check(ctx context.Context, log logr.Logger) error {
	srvTargets, err := c.resolveSRVTargets(ctx)
	if err != nil {
		return fmt.Errorf("matrix: resolve SRV: %w", err)
	}

	wellKnownTargets, err := c.resolveWellKnownTargets(ctx)
	if err != nil {
		return fmt.Errorf("matrix: check well-known: %w", err)
	}

	targets := []target{}

	for _, newTarget := range append(srvTargets, wellKnownTargets...) {
		exists := false
		for oldIdx := range targets {
			exists = targets[oldIdx] == newTarget
			if exists {
				break
			}
		}

		if !exists {
			targets = append(targets, newTarget)
		}
	}

	group := rungroup.New(ctx)

	for i := range targets {
		url := targets[i].url
		ipPort := targets[i].addr

		group.Go(func(ctx context.Context) error {
			err = c.connect(ctx, url, ipPort)
			if err != nil {
				return fmt.Errorf("connect to server: %w", err)
			}

			return nil
		}, rungroup.NeverCancel)
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("matrix: %w", err)
	}

	return nil
}

func (c Check) httpClient(toAddr, fromAddr string) *http.Client {
	network := "tcp6"
	if c.IPV4 {
		network = "tcp4"
	}

	transport := &http.Transport{
		IdleConnTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, _, actual string) (net.Conn, error) {
			if actual != fromAddr {
				panic(fmt.Sprintf("expected to be called with %s, got %s", fromAddr, actual))
			}

			return (&net.Dialer{}).DialContext(ctx, network, toAddr)
		},
	}

	return &http.Client{Transport: transport}
}
