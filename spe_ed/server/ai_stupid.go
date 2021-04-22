// SPDX-License-Identifier: Apache-2.0
// Copyright 2020,2021 Philipp Naumann, Marcus Soll
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"math/rand"
	"sync"
)

func init() {
	err := RegisterAI("StupidAI", func() AI { return new(StupidAI) })
	if err != nil {
		panic(err)
	}
}

// StupidAI always sends "change_nothing" except to avoid walls by turning.
type StupidAI struct {
	l sync.Mutex
	i chan string
}

// GetChannel receives the answer channel.
func (s *StupidAI) GetChannel(c chan string) {
	s.l.Lock()
	defer s.l.Unlock()
	s.i = c
}

// GetState gets the game state and computes an answer.
func (s *StupidAI) GetState(g *Game) {
	s.l.Lock()
	defer s.l.Unlock()

	if s.i == nil {
		return
	}

	if g.Running {
		p := g.Players[g.You]
		if s.isFree(p, g) {
			select {
			case s.i <- ActionNOOP:
			default:
			}
			return
		}

		if rand.Float64() < 0.5 {

			// Turn left
			switch p.Direction {
			case DirectionLeft:
				p.Direction = DirectionDown
			case DirectionRight:
				p.Direction = DirectionUp
			case DirectionUp:
				p.Direction = DirectionLeft
			case DirectionDown:
				p.Direction = DirectionRight
			}

			if s.isFree(p, g) {
				select {
				case s.i <- ActionTurnLeft:
				default:
				}
				return
			}

			// Last option: Turn right
			select {
			case s.i <- ActionTurnRight:
			default:
			}
			return
		}
		// Turn left
		switch p.Direction {
		case DirectionLeft:
			p.Direction = DirectionUp
		case DirectionRight:
			p.Direction = DirectionDown
		case DirectionUp:
			p.Direction = DirectionRight
		case DirectionDown:
			p.Direction = DirectionLeft
		}

		if s.isFree(p, g) {
			select {
			case s.i <- ActionTurnRight:
			default:
			}
			return
		}

		// Last option: Turn right
		select {
		case s.i <- ActionTurnLeft:
		default:
		}
	}
}

// isFree returns whether the next movement of the player is on a free cell.
// Not safe for concurrent usage on the same player or game.
func (s *StupidAI) isFree(p *Player, g *Game) bool {
	var x, y int
	switch p.Direction {
	case DirectionUp:
		x, y = p.X, p.Y-1
	case DirectionDown:
		x, y = p.X, p.Y+1
	case DirectionLeft:
		x, y = p.X-1, p.Y
	case DirectionRight:
		x, y = p.X+1, p.Y
	}
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return false
	}
	return g.Cells[y][x] == 0
}

// Name returns the name of the AI.
func (s *StupidAI) Name() string {
	return "StupidAI"
}
