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
	"net/http"
)

//nolint: tagliatelle
type listRoomsResponse struct {
	response
	Rooms []string `json:"joined_rooms"`
}

// Setup matrix and ensure required room is joined.
func (m *Matrix) Setup(ctx context.Context) error {
	url := m.HomeServer + "/_matrix/client/v3/joined_rooms"

	var listRoomsResponse listRoomsResponse

	status, err := m.sendHTTP(ctx, &listRoomsResponse, nil, http.MethodGet, url)
	if err != nil {
		return fmt.Errorf("list joined rooms: %w", err)
	}

	if err := listRoomsResponse.AsError(status); err != nil {
		return err
	}

	for _, room := range listRoomsResponse.Rooms {
		if room == m.RoomID {
			return nil
		}
	}

	url = m.HomeServer + "/_matrix/client/v3/join/" + m.RoomID

	var joinRoomResponse response

	status, err = m.sendHTTP(ctx, &joinRoomResponse, nil, http.MethodPost, url)
	if err != nil {
		return fmt.Errorf("join rooms: %w", err)
	}

	if err := joinRoomResponse.AsError(status); err != nil {
		return err
	}

	return nil
}
