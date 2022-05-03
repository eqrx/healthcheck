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
	"sync/atomic"
	"time"

	"eqrx.net/matrix"
	"eqrx.net/matrix/room"
	"eqrx.net/service"
)

// CrendentialsName is the name of the systemd credentials to load into Credentials.
const CrendentialsName = "matrix"

// Credentials contains information needed to connect to a matrix server.
type Credentials struct {
	Homeserver string `json:"homeserver"`
	Token      string `json:"token"`
}

// Sink messages into a matrix room. Does lazy deduplication.
type Sink struct {
	matrix      matrix.Client `yaml:"-"`
	rooms       []string      `yaml:"-"`
	lastMessage string        `yaml:"-"`
	name        string        `yaml:"-"`
}

var txID = time.Now().UnixMilli() //nolint:gochecknoglobals

// Setup the sink with values.
func (s *Sink) Setup(ctx context.Context, name string) error {
	var creds Credentials
	if err := service.UnmarshalYAMLCreds(CrendentialsName, &creds); err != nil {
		return err
	}

	matrix, err := matrix.New(ctx, creds.Homeserver, creds.Token)
	if err != nil {
		return fmt.Errorf("matrix create client: %w", err)
	}

	rooms, err := room.Joined(ctx, matrix)
	if err != nil {
		return fmt.Errorf("matrix get rooms: %w", err)
	}

	s.rooms = rooms
	s.name = name
	s.matrix = matrix

	return nil
}

// Sink spams messages in a matrix room and talks about received errors.
func (s *Sink) Sink(ctx context.Context, checkErr error) error {
	message := s.name + ": OK"
	if checkErr != nil {
		message = s.name + ": " + checkErr.Error()
	}

	if message == s.lastMessage {
		return nil
	}

	for _, roomID := range s.rooms {
		txID := fmt.Sprint(atomic.AddInt64(&txID, 1))
		if _, err := room.NewTextMessage(roomID, message).Send(ctx, s.matrix, txID); err != nil {
			return fmt.Errorf("send message: %w", err)
		}
	}

	s.lastMessage = message

	return nil
}
