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

// Package resolve allows to resolve DNS records via cloudflare DNS servers.
package resolve

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

const (
	// DNSServer defines the port and address from which to query DNS records.
	// Set to cloudflare DNS to avoid as much record caching as possible.
	DNSServer = "[2606:4700:4700::1111]:53"
	// MXPort defines the port that is assumed for SMTP servers pointed to from within MX records.
	MXPort = 25
)

// errNX indicates that a DNS record is not set.
var errNX = errors.New("dns answer is empty")

// exchange asks the given question to the DNS server set in DNSServer and returns the first answer.
// If there are no answers or something else went wrong this method returns an error.
func exchange(ctx context.Context, question *dns.Msg) (dns.RR, error) { //nolint:ireturn
	response, _, err := (&dns.Client{}).ExchangeContext(ctx, question, DNSServer)
	if err != nil {
		return nil, fmt.Errorf("request DNS answer: %w", err)
	}

	if len(response.Answer) < 1 {
		return nil, fmt.Errorf("%w, %v", errNX, response.Answer)
	}

	return response.Answer[0], nil
}

// add resolves a A or AAAA record with the given name and given type and forms a net.Dial compatible address string
// from the first answer and the given port. If there is no answer or another problem pops up an error is returned.
func addr(ctx context.Context, name string, port uint16, typ uint16) (string, error) {
	answer, err := exchange(ctx, (&dns.Msg{}).SetQuestion(dns.Fqdn(name), typ))
	if err != nil {
		return "", err
	}

	var resolvedIP net.IP

	switch r := answer.(type) {
	case *dns.A:
		resolvedIP = r.A
	case *dns.AAAA:
		resolvedIP = r.AAAA
	default:
		panic("unexpected dns type")
	}

	return net.JoinHostPort(resolvedIP.String(), fmt.Sprintf("%d", port)), nil
}

// Pointer resolves a SRV or MX record with the given name and pointerType. It then picks the first answer
// of the type recordType and returns the server name and a net.Dial compatible address string for it.
// If there is no answer or another problem pops up an error is returned.
func Pointer(ctx context.Context, name string, pointerType, recordType uint16) (string, string, error) {
	answer, err := exchange(ctx, (&dns.Msg{}).SetQuestion(dns.Fqdn(name), pointerType))
	if err != nil {
		return "", "", err
	}

	var serverName string

	var port uint16

	switch record := answer.(type) {
	case *dns.MX:
		serverName = strings.TrimSuffix(record.Mx, ".")
		port = MXPort
	case *dns.SRV:
		serverName = strings.TrimSuffix(record.Target, ".")
		port = record.Port
	default:
		panic("unknown pointer type")
	}

	address, err := addr(ctx, serverName, port, recordType)

	return serverName, address, err
}
