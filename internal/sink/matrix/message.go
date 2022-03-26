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
	"sync/atomic"
)

type message struct {
	Body    string `json:"body"`
	MsgType string `json:"msgtype"`
}

// SendText to the configured room.
func (m *Matrix) SendText(ctx context.Context, text string) error {
	txID := fmt.Sprint(atomic.AddUint64(&m.lastTxID, 1))
	url := m.HomeServer + "/_matrix/client/v3/rooms/" + m.RoomID + "/send/m.room.message/" + txID
	msg := message{text, "m.text"}

	var response response

	status, err := m.sendHTTP(ctx, &response, msg, http.MethodPut, url)
	if err != nil {
		return err
	}

	if err := response.AsError(status); err != nil {
		return err
	}

	return nil
}
