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

// Package internal itself provides functionality to bootstrap the healthcheck program.
package internal

import (
	"context"
	"fmt"
	"time"

	"dev.eqrx.net/healthcheck/internal/ceph"
	"dev.eqrx.net/healthcheck/internal/matrix"
	"dev.eqrx.net/healthcheck/internal/smtp"
)

// runTimeout defines how much a run of healthcheck may take.
const runTimeout = 1 * time.Minute

// Run fetches program configuration, determines which check healthcheck shall perform and runs it.
func Run(ctx context.Context) error {
	env, err := readEnv()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	switch env.service {
	case ServiceSMTPv4:
		err = checkConn(ctx, env.key, env.target, smtp.ResolveV4, smtp.Connect)
	case ServiceSMTPv6:
		err = checkConn(ctx, env.key, env.target, smtp.ResolveV6, smtp.Connect)
	case ServiceMatrixv4:
		err = checkConn(ctx, env.key, env.target, matrix.ResolveV4, matrix.Connect)
	case ServiceMatrixv6:
		err = checkConn(ctx, env.key, env.target, matrix.ResolveV6, matrix.Connect)
	case ServiceCeph:
		err = checkChecker(ctx, env.key, ceph.Check)
	default:
		return fmt.Errorf("%w: %v", errUnknownService, env.service)
	}

	if err != nil {
		return fmt.Errorf("%s check failed: %w", env.service, err)
	}

	return nil
}
