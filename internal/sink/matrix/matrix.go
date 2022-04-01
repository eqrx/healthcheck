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

// Matrix contains the configuration for a matrix room.
type Matrix struct {
	HomeServer string `yaml:"homeServer"`
	RoomID     string `yaml:"roomId"`
	Token      string `yaml:"token"`
	lastTxID   uint64 `yaml:"-"`
}
