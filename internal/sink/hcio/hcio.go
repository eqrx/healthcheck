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

// Package hcio sends pings to healthchecks.io.
package hcio

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// Sink sends pings to healthchecks.io.
type Sink struct {
	UUID string `yaml:"uuid"`
}

// errCode is raised when the healthchecks.io call returned an unexpected HTTP status code.
var errCode = errors.New("ping failed with unexpected code")

// Sink performs a HTTP request to the healthchecks.io servers to ping the check identified by the given UUID.
// It returns nil if the ping was successful.
func (s Sink) Sink(ctx context.Context, checkErr error) error {
	if isOK := checkErr == nil; !isOK {
		return nil
	}

	url := "https://hc-ping.com/" + s.UUID

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		panic(fmt.Sprintf("create request: %v", err))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ping request: %w", err)
	}

	if err := resp.Body.Close(); err != nil {
		panic(fmt.Sprintf("close body: %v", err))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: %v", errCode, resp.Status)
	}

	return nil
}
