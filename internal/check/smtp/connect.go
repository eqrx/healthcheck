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
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"eqrx.net/rungroup"
)

// errProtocol indicates the server did not adhere to the SMTP protocol.
var errProtocol = errors.New("protocol failure")

// prepareTLS exchanges messages required by the SMTP protocol to enable STARTTLS.
func prepareTLS(conn io.ReadWriter) error {
	reader := bufio.NewReader(conn)

	_, err := conn.Write([]byte("EHLO healthcheck\n"))
	if err != nil {
		return fmt.Errorf("connect: prepare TLS: write EHLO: %w", err)
	}

	l, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("connect: prepare TLS: read EHLO response: %w", err)
	}

	if !strings.HasPrefix(l, "220") {
		return fmt.Errorf("connect: prepare TLS: %w: server didn't like EHLO", errProtocol)
	}

	_, err = conn.Write([]byte("STARTTLS\n"))

	if err != nil {
		return fmt.Errorf("connect: prepare TLS: write STARTTLS: %w", err)
	}

	// Need to read all the lines the server sent before it accepted our STARTTLS.
	for {
		l, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("connect: prepare TLS: read server response: %w", err)
		}

		if strings.HasPrefix(l, "220") {
			break
		}
	}

	return nil
}

func (c Check) connect(ctx context.Context, addr string) error {
	conn, err := (&net.Dialer{}).DialContext(ctx, c.network, addr)
	if err != nil {
		return fmt.Errorf("connect: dial: %w", err)
	}

	group := rungroup.New(ctx)

	group.Go(func(ctx context.Context) error {
		<-ctx.Done()

		if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("connect: close: %w", err)
		}

		return nil
	})

	group.Go(func(ctx context.Context) error {
		if err := prepareTLS(conn); err != nil {
			return err
		}
		tlsConfig := &tls.Config{ServerName: c.Domain, MinVersion: tls.VersionTLS13}
		tlsConn := tls.Client(conn, tlsConfig)

		if err := tlsConn.Handshake(); err != nil {
			return fmt.Errorf("connect: tls: %w", err)
		}

		_, err = conn.Write([]byte("QUIT\n"))

		if err != nil {
			return fmt.Errorf("connect: sending QUIT: %w", err)
		}

		return nil
	})

	if err := group.Wait(); err != nil {
		return fmt.Errorf("domain %s: %w", c.Domain, err)
	}

	return nil
}
