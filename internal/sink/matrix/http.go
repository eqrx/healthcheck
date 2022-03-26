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

// Package matrix sinks reports to a matrix room.
package matrix

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type response struct {
	ErrCode string `json:"errcode"`
	ErrMsg  string `json:"error"`
}

func (r response) AsError(httpStatus int) error {
	if r.ErrCode != "" || r.ErrMsg != "" || httpStatus != http.StatusOK {
		return fmt.Errorf("%w: %s: %s: %s", errMatrix, http.StatusText(httpStatus), r.ErrCode, r.ErrMsg)
	}

	return nil
}

var errMatrix = errors.New("matrix reports")

func (m *Matrix) sendHTTP(ctx context.Context, response, request interface{}, method string, url string) (int, error) {
	requestBody := &bytes.Buffer{}

	if request != nil {
		if err := json.NewEncoder(requestBody).Encode(request); err != nil {
			return -1, fmt.Errorf("marshal request: %w", err)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		panic(fmt.Sprintf("new http req: %v", err))
	}

	httpReq.Header.Set("Authorization", "Bearer "+m.Token)

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return -1, fmt.Errorf("http: %w", err)
	}

	err = json.NewDecoder(httpResp.Body).Decode(response)
	cErr := httpResp.Body.Close()

	switch {
	case err != nil && cErr != nil:
		return httpResp.StatusCode, fmt.Errorf("resp unmarshal: %w, resp close: %v", err, cErr)
	case err != nil:
		return httpResp.StatusCode, fmt.Errorf("resp unmarshal: %w", err)
	case cErr != nil:
		return httpResp.StatusCode, fmt.Errorf("resp close: %w", cErr)
	default:
		return httpResp.StatusCode, nil
	}
}
