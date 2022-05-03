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

package matrix

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
)

func (c Check) resolveWellKnownTargets(ctx context.Context) ([]target, error) {
	requestURL := "https://" + c.Domain + "/.well-known/matrix/client"

	addrs, err := c.resolveAddr(ctx, c.Domain+".")
	if err != nil {
		return nil, err
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("addr not found")
	}

	targets := []target{}

	for i := range addrs {
		newTargets, err := c.resolveWellKnownTarget(ctx, requestURL, addrs[i])
		if err != nil {
			return nil, err
		}

		targets = append(targets, newTargets...)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("addr not found")
	}

	return targets, nil
}

func (c Check) resolveWellKnownTarget(ctx context.Context, requestURL string, addr netip.Addr) ([]target, error) {
	var wellKnown wellKnown

	hostPort := net.JoinHostPort(c.Domain, "443")
	hostPortResolved := net.JoinHostPort(addr.String(), "443")

	if err := c.loadHTTP(ctx, requestURL, hostPort, hostPortResolved, &wellKnown); err != nil {
		return nil, fmt.Errorf("%s: %w", hostPortResolved, err)
	}

	url, err := url.Parse(wellKnown.Homeserver.URL)
	if err != nil {
		return nil, fmt.Errorf("url: %w", err)
	}

	hostPart, portPart, err := net.SplitHostPort(url.Host)
	if err != nil {
		return nil, fmt.Errorf("url: %w", err)
	}

	addrs, err := c.resolveAddr(ctx, hostPart+".")
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portPart)
	if err != nil {
		return nil, fmt.Errorf("url: %w", err)
	}

	targets := []target{}
	for _, addr := range addrs {
		targets = append(targets, target{*url, netip.AddrPortFrom(addr, uint16(port))})
	}

	return targets, nil
}

func (c Check) loadHTTP(ctx context.Context, url, hostPort, ipPort string, dst interface{}) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		panic(fmt.Sprintf("create http request: %v", err))
	}

	response, err := c.httpClient(ipPort, hostPort).Do(request)
	if err != nil {
		return fmt.Errorf("execute http request: %w", err)
	}

	readErr := json.NewDecoder(response.Body).Decode(dst)

	closeErr := response.Body.Close()

	switch {
	case readErr != nil && closeErr != nil:
		return fmt.Errorf("read body: %w; close body: %v", readErr, closeErr)
	case readErr != nil:
		return fmt.Errorf("read body: %w", readErr)
	case closeErr != nil:
		return fmt.Errorf("close body: %w", closeErr)
	case response.StatusCode < 200 || response.StatusCode >= 300:
		return fmt.Errorf("unexpected http response status: %v", response.StatusCode)
	default:
		return nil
	}
}
