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

package sink

import (
	"context"
	"errors"

	"eqrx.net/healthcheck/internal/sink/hcio"
	matrixsink "eqrx.net/healthcheck/internal/sink/matrix"
)

var errConcrete = errors.New("more or less than one concrete types set for sink")

// Sink contains concrete sink implementations.
type Sink struct {
	Matrix *matrixsink.Sink                   `yaml:"matrix"`
	HCIO   *hcio.Sink                         `yaml:"hcio"`
	Sink   func(context.Context, error) error `yaml:"-"`
}

// Setup the given sink for sending.
func (s *Sink) Setup(ctx context.Context, name string) error {
	if s.Matrix != nil {
		if s.Sink != nil {
			return errConcrete
		}

		if err := s.Matrix.Setup(ctx, name); err != nil {
			return err
		}

		s.Sink = s.Matrix.Sink
	}

	if s.HCIO != nil {
		if s.Sink != nil {
			return errConcrete
		}

		s.Sink = s.HCIO.Sink
	}

	if s.Sink == nil {
		return errConcrete
	}

	return nil
}
