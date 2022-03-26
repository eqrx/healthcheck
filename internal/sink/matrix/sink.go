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
)

// Sink messages into a matrix room. Does lazy deduplication.
type Sink struct {
	matrix      *Matrix `yaml:"-"`
	lastMessage string  `yaml:"-"`
	name        string  `yaml:"-"`
}

// Setup the sink with values.
func (s *Sink) Setup(name string, matrix *Matrix) {
	s.matrix = matrix
	s.name = name
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

	if err := s.matrix.SendText(ctx, message); err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	s.lastMessage = message

	return nil
}
