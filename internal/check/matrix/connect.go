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
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
)

var errCode = errors.New("matrix endpoint failed")

func (c Check) connect(ctx context.Context, url url.URL, addr netip.AddrPort) error {
	httpClient := c.httpClient(addr.String(), url.Host)

	request, err := http.NewRequestWithContext(ctx, http.MethodHead, url.String(), nil)
	if err != nil {
		panic(fmt.Sprintf("create http request: %v", err))
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("execute http request: %w", err)
	}

	if err := response.Body.Close(); err != nil {
		return fmt.Errorf("close http response body: %w", err)
	}

	if response.StatusCode != http.StatusNotFound {
		return errCode
	}

	return nil
}
