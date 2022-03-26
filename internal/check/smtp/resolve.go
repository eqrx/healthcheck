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

package smtp

import (
	"context"
	"fmt"

	"github.com/miekg/dns"
)

const (
	// dnsServer defines the port and address from which to query DNS records.
	// Set to cloudflare DNS to avoid as much record caching as possible.
	dnsServer = "[2606:4700:4700::1111]:53"
	// smtpPort defines the port that is assumed for SMTP servers pointed to from within MX records.
	smtpPort = "25"
)

func (c Check) resolveServer(ctx context.Context) ([]string, error) {
	serverNames, err := c.resolveMX(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve MX: %w", err)
	}

	serverAddrs := []string{}

	for _, serverName := range serverNames {
		question := (&dns.Msg{}).SetQuestion(serverName, c.targetRRType)

		answer, _, err := (&dns.Client{}).ExchangeContext(ctx, question, dnsServer)
		if err != nil {
			return nil, fmt.Errorf("dns exchange: %w", err)
		}

		for i := range answer.Answer {
			switch record := answer.Answer[i].(type) {
			case *dns.A:
				if c.targetRRType == dns.TypeA {
					serverAddrs = append(serverAddrs, record.A.String())
				}
			case *dns.AAAA:
				if c.targetRRType == dns.TypeAAAA {
					serverAddrs = append(serverAddrs, record.AAAA.String())
				}
			}
		}
	}

	return serverAddrs, nil
}

func (c Check) resolveMX(ctx context.Context) ([]string, error) {
	question := (&dns.Msg{}).SetQuestion(dns.Fqdn(c.Domain), dns.TypeMX)

	answer, _, err := (&dns.Client{}).ExchangeContext(ctx, question, dnsServer)
	if err != nil {
		return nil, fmt.Errorf("dns exchange: %w", err)
	}

	serverNames := []string{}

	for i := range answer.Answer {
		mxRecord, ok := answer.Answer[i].(*dns.MX)
		if ok {
			serverNames = append(serverNames, mxRecord.Mx)
		}
	}

	return serverNames, nil
}
