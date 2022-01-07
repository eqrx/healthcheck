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

// Package smtp allows to monitor SMTP servers.
package smtp

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"dev.eqrx.net/healthcheck/internal/resolve"
	"dev.eqrx.net/rungroup"
	"github.com/miekg/dns"
)

// errProtocol indicates the server did not adhere to the SMTP protocol.
var errProtocol = errors.New("protocol failure")

// prepareTLS exchanges messages required by the SMTP protocol to enable STARTTLS.
func prepareTLS(conn io.ReadWriter) error {
	reader := bufio.NewReader(conn)

	_, err := conn.Write([]byte("EHLO healthcheck\n"))
	if err != nil {
		return fmt.Errorf("coult not write EHLO: %w", err)
	}

	l, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("coult not read EHLO response: %w", err)
	}

	if !strings.HasPrefix(l, "220") {
		return fmt.Errorf("%w: server did not like out EHLO", errProtocol)
	}

	_, err = conn.Write([]byte("STARTTLS\n"))

	if err != nil {
		return fmt.Errorf("coult not write STARTTLS: %w", err)
	}

	// Need to read all the lines the server sent before it accepted our STARTTLS.
	for {
		l, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("could not read server response: %w", err)
		}

		if strings.HasPrefix(l, "220") {
			break
		}
	}

	return nil
}

// Connect dials a tcp connection to the given address addr. It then exchanges greeting messages with the SMTP server
// behind the connection and requests STARTTLS. It then opens a TLSv1.3 connection and checks if the server has the
// given server name serverName and that its certificate is trusted by the local CA storage and closes the
// SMTP connection afterwards.
func Connect(ctx context.Context, serverName string, addr string) error {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("could not dial: %w", err)
	}

	group := rungroup.New(ctx)

	group.Go(func(ctx context.Context) error {
		<-ctx.Done()

		if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("could not close connection to smtp server: %w", err)
		}

		return nil
	})

	group.Go(func(ctx context.Context) error {
		if err := prepareTLS(conn); err != nil {
			return err
		}
		tlsConfig := &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS13}
		tlsConn := tls.Client(conn, tlsConfig)

		if err := tlsConn.Handshake(); err != nil {
			return fmt.Errorf("tls handshake failed: %w", err)
		}

		_, err = conn.Write([]byte("QUIT\n"))

		if err != nil {
			return fmt.Errorf("sending QUIT failed: %w", err)
		}

		return nil
	})

	if err := group.Wait(); err != nil {
		return fmt.Errorf("smtp check run group for serverName %s failed: %w", serverName, err)
	}

	return nil
}

// ResolveV6 resolves the IPv6 address of the SMTP server responsible for the given domain and
// returns the first IPv6 address.
func ResolveV6(ctx context.Context, domain string) (string, string, error) {
	hostname, addr, err := resolve.Pointer(ctx, domain, dns.TypeMX, dns.TypeAAAA)
	if err != nil {
		return "", "", fmt.Errorf("could not query address of SMTP server for domain %s: %w", domain, err)
	}

	return hostname, addr, nil
}

// ResolveV4 resolves the IPv4 address of the SMTP server responsible for the given domain and
// returns the first IPv4 address.
func ResolveV4(ctx context.Context, domain string) (string, string, error) {
	hostname, addr, err := resolve.Pointer(ctx, domain, dns.TypeMX, dns.TypeA)
	if err != nil {
		return "", "", fmt.Errorf("could not query address of SMTP server for domain %s: %w", domain, err)
	}

	return hostname, addr, nil
}
