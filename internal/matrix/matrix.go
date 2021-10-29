// Copyright (C) 2021 Alexander Sowitzki
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

// Package matrix allows to check the status synapse matrix server.
package matrix

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/eqrx/healthcheck/internal/resolve"
	"github.com/miekg/dns"
)

// StatusOK is the value that synapse reports as health status when everything is good.
const StatusOK = "OK"

var (
	// errHealth indicates that the matrix server did not report StatusOK on its health endpoint.
	errHealth = errors.New("matrix reports unhealthy")
	// errCode indicates the matrix server returned some HTTP status code != 200.
	errCode = errors.New("matrix endpoint failed")
)

// Connect dials a tcp connection to the given ip and attempts to establish a TLSv1.3 session over it. It validates the
// servers certificate using the host CA set and the server name using serverName. It then performs a HTTP
// GET on the path /health. If the request is answered with HTML status code 200 and the response payload is equal to
// the StatusOK value, the method returns nil.
func Connect(ctx context.Context, serverName string, addr string) error {
	httpClient := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
			TLSClientConfig:   &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS13},
		},
	}

	url := fmt.Sprintf("https://%s/health", addr)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		panic("could not create request")
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("could not execute http request: %w", err)
	}

	body, err := ioutil.ReadAll(response.Body)

	if err := response.Body.Close(); err != nil {
		return fmt.Errorf("could not close http response body: %w", err)
	}

	if err != nil {
		return fmt.Errorf("could not read http body: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return errCode
	}

	if string(body) != "OK" {
		return errHealth
	}

	return nil
}

// ResolveV6 resolves the IPv6 address of the matrix server responsible for the given domain.
// Is does so by searching for the DNS SRV _matrix._tcp.$domain and returning the first IPv6 address.
func ResolveV6(ctx context.Context, domain string) (string, string, error) {
	name := fmt.Sprintf("_matrix._tcp.%v", domain)

	hostname, addr, err := resolve.Pointer(ctx, name, dns.TypeSRV, dns.TypeAAAA)
	if err != nil {
		return "", "", fmt.Errorf("could not query address of matrix server for domain %s: %w", domain, err)
	}

	return hostname, addr, nil
}

// ResolveV4 resolves the IPv4 address of the matrix server responsible for the given domain.
// Is does so by searching for the DNS SRV _matrix._tcp.$domain and returning the first IPv4 address.
func ResolveV4(ctx context.Context, domain string) (string, string, error) {
	name := fmt.Sprintf("_matrix._tcp.%v", domain)

	hostname, addr, err := resolve.Pointer(ctx, name, dns.TypeSRV, dns.TypeA)
	if err != nil {
		return "", "", fmt.Errorf("could not query address of matrix server for domain %s: %w", domain, err)
	}

	return hostname, addr, nil
}
