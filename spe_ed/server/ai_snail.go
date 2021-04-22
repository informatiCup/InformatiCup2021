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
	err := RegisterAI("SnailAI", func() AI { return new(SnailAI) })
	if err != nil {
		panic(err)
	}
}

// SnailAI is an AI that tries to maximise space usage by always 'holding one hand to the wall'. It will usually perform a snail-like pattern at the beginning, thus the name.
type SnailAI struct {
	l         sync.Mutex
	i         chan string
	direction string
}

// GetChannel receives the answer channel.
func (s *SnailAI) GetChannel(c chan string) {
	s.l.Lock()
	defer s.l.Unlock()

	s.i = c
}

// GetState gets the game state and computes an answer.
func (s *SnailAI) GetState(g *Game) {
	s.l.Lock()
	defer s.l.Unlock()

	if s.i == nil {
		return
	}

	if s.direction == "" {
		if rand.Float32() < 0.5 {
			s.direction = DirectionLeft
		} else {
			s.direction = DirectionRight
		}
	}

	if g.Running {
		if s.direction == DirectionLeft {
			var nextX, nextY int
			switch g.Players[g.You].Direction {
			case DirectionUp:
				nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
			case DirectionDown:
				nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
			case DirectionLeft:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
			case DirectionRight:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
			}
			if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
				select {
				case s.i <- ActionTurnRight:
				default:
				}
				return
			}

			switch g.Players[g.You].Direction {
			case DirectionUp:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
			case DirectionDown:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
			case DirectionLeft:
				nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
			case DirectionRight:
				nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
			}
			if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
				select {
				case s.i <- ActionNOOP:
				default:
				}
				return
			}

			switch g.Players[g.You].Direction {
			case DirectionUp:
				nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
			case DirectionDown:
				nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
			case DirectionLeft:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
			case DirectionRight:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
			}
			if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
				select {
				case s.i <- ActionTurnLeft:
				default:
				}
				return
			}
		}
		if s.direction == DirectionRight {
			var nextX, nextY int
			switch g.Players[g.You].Direction {
			case DirectionUp:
				nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
			case DirectionDown:
				nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
			case DirectionLeft:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
			case DirectionRight:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
			}
			if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
				select {
				case s.i <- ActionTurnLeft:
				default:
				}
				return
			}

			switch g.Players[g.You].Direction {
			case DirectionUp:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
			case DirectionDown:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
			case DirectionLeft:
				nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
			case DirectionRight:
				nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
			}
			if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
				select {
				case s.i <- ActionNOOP:
				default:
				}
				return
			}

			switch g.Players[g.You].Direction {
			case DirectionUp:
				nextX, nextY = g.Players[g.You].X+1, g.Players[g.You].Y
			case DirectionDown:
				nextX, nextY = g.Players[g.You].X-1, g.Players[g.You].Y
			case DirectionLeft:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y-1
			case DirectionRight:
				nextX, nextY = g.Players[g.You].X, g.Players[g.You].Y+1
			}
			if nextX >= 0 && nextX < g.Width && nextY >= 0 && nextY < g.Height && g.Cells[nextY][nextX] == 0 {
				select {
				case s.i <- ActionTurnRight:
				default:
				}
				return
			}
		}
		select {
		case s.i <- ActionNOOP:
		default:
		}
	}
}

// Name returns the name of the AI.
func (s *SnailAI) Name() string {
	return "SnailAI"
}
