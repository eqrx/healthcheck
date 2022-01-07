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

package internal

import (
	"context"
	"fmt"

	"dev.eqrx.net/healthcheck/internal/ping"
)

type (
	// resolverFn shall resolves a predefined service and returns its DNS name and a address string that can be used
	// for net.Dial. If err is nil, the other return values may not be nil.
	resolverFn func(context.Context, string) (string, string, error)
	// connectorFn shall connect to a service behind a DNS name and an address.
	// It shall return an error is the connection is not successful.
	connectorFn func(context.Context, string, string) error
)

// checkConn resolves an address using the resolverFn resolver, calls connectorFn connection with the domain and
// result and IP if it succeeded and pings healthchecks.io with key if that succeeds too. After succeeding
// with that it returns with nil as error. If there is a failure at any point, it returns with an error.
func checkConn(ctx context.Context, key string, target string, resolver resolverFn, connector connectorFn) error {
	_, addr, err := resolver(ctx, target)
	if err != nil {
		return fmt.Errorf("resolve failed: %w", err)
	}

	if err = connector(ctx, target, addr); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	if err := ping.Send(ctx, key); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}

// checkChecker calls the function checker. If it returns nil as error it pings healthchecks.io with key and return nil.
// If there is a failure at any point, it returns with an error.
func checkChecker(ctx context.Context, key string, checker func(context.Context) error) error {
	if err := checker(ctx); err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	if err := ping.Send(ctx, key); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}
