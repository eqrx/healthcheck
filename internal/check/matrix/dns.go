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
	"fmt"
	"net/netip"
	"net/url"

	"github.com/miekg/dns"
)

// dnsServer defines the port and address from which to query DNS records.
// Set to cloudflare DNS to avoid as much record caching as possible.
const dnsServer = "[2606:4700:4700::1111]:53"

func (c Check) resolveSRVTargets(ctx context.Context) ([]target, error) {
	question := (&dns.Msg{}).SetQuestion("_matrix._tcp."+c.Domain+".", dns.TypeSRV)

	answer, _, err := (&dns.Client{}).ExchangeContext(ctx, question, dnsServer)
	if err != nil {
		return nil, fmt.Errorf("dns exchange: %w", err)
	}

	targets := []target{}

	for i := range answer.Answer {
		srvRecord, ok := answer.Answer[i].(*dns.SRV)
		if !ok {
			return nil, fmt.Errorf("answer type not matching request: exptected SRV, got %T", answer.Answer[i])
		}

		addrs, err := c.resolveAddr(ctx, srvRecord.Target)
		if err != nil {
			return nil, fmt.Errorf("resolve %s: %w", srvRecord.Target, err)
		}

		url, err := url.Parse("https://" + srvRecord.Target + ":" + fmt.Sprint(srvRecord.Port))
		if err != nil {
			return nil, fmt.Errorf("url: %w", err)
		}

		for _, addr := range addrs {
			targets = append(targets, target{*url, netip.AddrPortFrom(addr, srvRecord.Port)})
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("addr not found")
	}

	return targets, nil
}

func (c Check) resolveAddr(ctx context.Context, name string) ([]netip.Addr, error) {
	question := (&dns.Msg{}).SetQuestion(name, c.targetRRType)

	answer, _, err := (&dns.Client{}).ExchangeContext(ctx, question, dnsServer)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	serverAddrs := []netip.Addr{}

	for i := range answer.Answer {
		var addr netip.Addr

		var addrOk bool

		switch record := answer.Answer[i].(type) {
		case *dns.A:
			if c.targetRRType == dns.TypeA {
				addr, addrOk = netip.AddrFromSlice(record.A)
			}
		case *dns.AAAA:
			if c.targetRRType == dns.TypeAAAA {
				addr, addrOk = netip.AddrFromSlice(record.AAAA)
			}
		}

		if !addrOk || addr.Is4() != c.IPV4 {
			panic("invalid ip")
		}

		serverAddrs = append(serverAddrs, addr)
	}

	return serverAddrs, nil
}
